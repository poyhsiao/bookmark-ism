#!/bin/bash

# Secret Rotation Script for Bookmark Sync Service
# This script safely rotates Docker Swarm secrets with zero downtime

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
STACK_NAME="bookmark-sync"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Check if running in Docker Swarm mode
check_swarm() {
    if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
        log_error "Docker is not running in swarm mode"
        exit 1
    fi
}

# Backup current secrets
backup_secrets() {
    log_step "Backing up current secrets..."

    local backup_dir="$PROJECT_ROOT/secrets/backup/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"

    cd "$PROJECT_ROOT/secrets"

    local secrets=("postgres_password.txt" "jwt_secret.txt" "redis_password.txt" "typesense_api_key.txt" "minio_root_password.txt")

    for secret_file in "${secrets[@]}"; do
        if [[ -f "$secret_file" ]]; then
            cp "$secret_file" "$backup_dir/"
            log_info "Backed up $secret_file"
        fi
    done

    log_success "Secrets backed up to: $backup_dir"
}

# Generate new secret
generate_new_secret() {
    local secret_name="$1"
    local secret_file="$PROJECT_ROOT/secrets/${secret_name}.txt"
    local temp_file="$PROJECT_ROOT/secrets/${secret_name}.new.txt"

    log_info "Generating new $secret_name..."

    case "$secret_name" in
        "jwt_secret")
            openssl rand -base64 64 > "$temp_file"
            ;;
        *)
            openssl rand -base64 32 > "$temp_file"
            ;;
    esac

    chmod 600 "$temp_file"
    chown $(whoami):$(whoami) "$temp_file"

    log_success "Generated new $secret_name"
}

# Rotate single secret
rotate_secret() {
    local secret_name="$1"
    local secret_file="$PROJECT_ROOT/secrets/${secret_name}.txt"
    local temp_file="$PROJECT_ROOT/secrets/${secret_name}.new.txt"
    local new_secret_name="${secret_name}_new"

    log_step "Rotating secret: $secret_name"

    # Generate new secret
    generate_new_secret "$secret_name"

    # Create new Docker secret
    log_info "Creating new Docker secret: $new_secret_name"
    docker secret create "$new_secret_name" "$temp_file"

    # Get services using this secret
    local services=$(docker service ls --format "{{.Name}}" | grep "^${STACK_NAME}_")

    # Update services to use new secret
    for service in $services; do
        local service_config=$(docker service inspect "$service" --format '{{json .Spec.TaskTemplate.ContainerSpec.Secrets}}')

        if echo "$service_config" | grep -q "\"SecretName\":\"$secret_name\""; then
            log_info "Updating service $service to use new secret"

            # Add new secret and remove old one
            docker service update \
                --secret-add "source=$new_secret_name,target=$secret_name" \
                --secret-rm "$secret_name" \
                "$service"

            log_success "Updated $service"
        fi
    done

    # Wait for services to be healthy
    log_info "Waiting for services to stabilize..."
    sleep 30

    # Verify services are healthy
    local unhealthy_services=0
    for service in $services; do
        local replicas=$(docker service ls --filter "name=$service" --format "{{.Replicas}}")
        if [[ ! "$replicas" =~ ^[1-9][0-9]*/[1-9][0-9]*$ ]]; then
            log_warning "Service $service may not be healthy: $replicas"
            ((unhealthy_services++))
        fi
    done

    if [[ $unhealthy_services -eq 0 ]]; then
        # Remove old secret
        log_info "Removing old Docker secret: $secret_name"
        docker secret rm "$secret_name" || log_warning "Failed to remove old secret $secret_name"

        # Rename new secret to original name
        log_info "Creating final Docker secret: $secret_name"
        docker secret create "$secret_name" "$temp_file"
        docker secret rm "$new_secret_name"

        # Update services back to original secret name
        for service in $services; do
            local service_config=$(docker service inspect "$service" --format '{{json .Spec.TaskTemplate.ContainerSpec.Secrets}}')

            if echo "$service_config" | grep -q "\"SecretName\":\"$new_secret_name\""; then
                docker service update \
                    --secret-add "source=$secret_name,target=$secret_name" \
                    --secret-rm "$new_secret_name" \
                    "$service"
            fi
        done

        # Replace old secret file with new one
        mv "$temp_file" "$secret_file"

        log_success "Successfully rotated secret: $secret_name"
    else
        log_error "Some services are unhealthy. Rolling back..."
        rollback_secret "$secret_name" "$new_secret_name"
        return 1
    fi
}

# Rollback secret rotation
rollback_secret() {
    local secret_name="$1"
    local new_secret_name="$2"

    log_warning "Rolling back secret rotation for: $secret_name"

    local services=$(docker service ls --format "{{.Name}}" | grep "^${STACK_NAME}_")

    # Rollback services to original secret
    for service in $services; do
        local service_config=$(docker service inspect "$service" --format '{{json .Spec.TaskTemplate.ContainerSpec.Secrets}}')

        if echo "$service_config" | grep -q "\"SecretName\":\"$new_secret_name\""; then
            docker service update \
                --secret-add "source=$secret_name,target=$secret_name" \
                --secret-rm "$new_secret_name" \
                "$service"
        fi
    done

    # Remove new secret
    docker secret rm "$new_secret_name" || true

    # Remove temporary file
    rm -f "$PROJECT_ROOT/secrets/${secret_name}.new.txt"

    log_success "Rollback completed for: $secret_name"
}

# Rotate all secrets
rotate_all_secrets() {
    log_step "Rotating all secrets..."

    local secrets=("postgres_password" "redis_password" "typesense_api_key" "minio_root_password")
    local failed_rotations=0

    # Backup current secrets first
    backup_secrets

    # Rotate each secret
    for secret in "${secrets[@]}"; do
        if ! rotate_secret "$secret"; then
            log_error "Failed to rotate $secret"
            ((failed_rotations++))
        fi
    done

    # Handle JWT secret separately (requires more careful handling)
    log_warning "JWT secret rotation requires manual verification"
    log_info "To rotate JWT secret, run: $0 rotate jwt_secret"

    if [[ $failed_rotations -eq 0 ]]; then
        log_success "All secrets rotated successfully"
    else
        log_error "$failed_rotations secret rotations failed"
        return 1
    fi
}

# Verify secret rotation
verify_secrets() {
    log_step "Verifying secret rotation..."

    local secrets=("postgres_password" "jwt_secret" "redis_password" "typesense_api_key" "minio_root_password")
    local verification_errors=0

    # Check if secret files exist
    for secret in "${secrets[@]}"; do
        local secret_file="$PROJECT_ROOT/secrets/${secret}.txt"
        if [[ ! -f "$secret_file" ]]; then
            log_error "Secret file missing: $secret_file"
            ((verification_errors++))
        fi
    done

    # Check if Docker secrets exist
    for secret in "${secrets[@]}"; do
        if ! docker secret ls --format "{{.Name}}" | grep -q "^${secret}$"; then
            log_error "Docker secret missing: $secret"
            ((verification_errors++))
        fi
    done

    # Check service health
    local services=$(docker service ls --format "{{.Name}}" | grep "^${STACK_NAME}_")
    for service in $services; do
        local replicas=$(docker service ls --filter "name=$service" --format "{{.Replicas}}")
        if [[ ! "$replicas" =~ ^[1-9][0-9]*/[1-9][0-9]*$ ]]; then
            log_error "Service $service is not healthy: $replicas"
            ((verification_errors++))
        fi
    done

    if [[ $verification_errors -eq 0 ]]; then
        log_success "Secret verification passed"
        return 0
    else
        log_error "Secret verification failed with $verification_errors errors"
        return 1
    fi
}

# List current secrets
list_secrets() {
    log_step "Current secrets status:"

    echo ""
    echo -e "${CYAN}Docker Swarm Secrets:${NC}"
    docker secret ls

    echo ""
    echo -e "${CYAN}Secret Files:${NC}"
    cd "$PROJECT_ROOT/secrets"
    ls -la *.txt 2>/dev/null || echo "No secret files found"

    echo ""
    echo -e "${CYAN}Service Status:${NC}"
    docker service ls --filter "name=${STACK_NAME}_"
}

# Clean up old secret backups
cleanup_backups() {
    log_step "Cleaning up old secret backups..."

    local backup_dir="$PROJECT_ROOT/secrets/backup"
    local retention_days="${BACKUP_RETENTION_DAYS:-30}"

    if [[ -d "$backup_dir" ]]; then
        find "$backup_dir" -type d -mtime +$retention_days -exec rm -rf {} + 2>/dev/null || true
        log_success "Cleaned up backups older than $retention_days days"
    else
        log_info "No backup directory found"
    fi
}

# Main function
main() {
    local command="${1:-help}"
    local secret_name="${2:-}"

    case "$command" in
        "rotate")
            check_swarm
            if [[ -n "$secret_name" ]]; then
                rotate_secret "$secret_name"
            else
                rotate_all_secrets
            fi
            ;;
        "verify")
            check_swarm
            verify_secrets
            ;;
        "list")
            check_swarm
            list_secrets
            ;;
        "backup")
            backup_secrets
            ;;
        "cleanup")
            cleanup_backups
            ;;
        "help"|*)
            echo "Usage: $0 {rotate|verify|list|backup|cleanup} [secret_name]"
            echo ""
            echo "Commands:"
            echo "  rotate [secret]  - Rotate all secrets or specific secret"
            echo "  verify           - Verify secret rotation status"
            echo "  list             - List current secrets and service status"
            echo "  backup           - Backup current secrets"
            echo "  cleanup          - Clean up old secret backups"
            echo ""
            echo "Available secrets:"
            echo "  - postgres_password"
            echo "  - jwt_secret"
            echo "  - redis_password"
            echo "  - typesense_api_key"
            echo "  - minio_root_password"
            echo ""
            echo "Examples:"
            echo "  $0 rotate                    # Rotate all secrets"
            echo "  $0 rotate postgres_password  # Rotate specific secret"
            echo "  $0 verify                    # Verify rotation status"
            echo "  $0 list                      # Show current status"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"