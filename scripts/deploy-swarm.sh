#!/bin/bash

# Docker Swarm Production Deployment Script
# Task 26: Production deployment infrastructure with container orchestration

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
COMPOSE_FILE="docker-compose.swarm.yml"
ENV_FILE=".env.production"

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

log_debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        echo -e "${CYAN}[DEBUG]${NC} $1"
    fi
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    # Check if Docker Compose is available
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available. Please install Docker Compose."
        exit 1
    fi

    # Check if running in swarm mode
    if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
        log_error "Docker is not running in swarm mode. Please initialize swarm first."
        log_info "Run: docker swarm init --advertise-addr <MANAGER-IP>"
        exit 1
    fi

    # Check if compose file exists
    if [[ ! -f "$PROJECT_ROOT/$COMPOSE_FILE" ]]; then
        log_error "Compose file not found: $COMPOSE_FILE"
        exit 1
    fi

    # Check if environment file exists
    if [[ ! -f "$PROJECT_ROOT/$ENV_FILE" ]]; then
        log_warning "Environment file not found: $ENV_FILE"
        log_info "Creating default environment file..."
        create_default_env_file
    fi

    log_success "Prerequisites check completed"
}

# Create default environment file
create_default_env_file() {
    cat > "$PROJECT_ROOT/$ENV_FILE" << 'EOF'
# Production Environment Configuration
# Task 26: Production deployment infrastructure

# Domain Configuration
DOMAIN_NAME=bookmark-sync.example.com
SITE_URL=https://bookmark-sync.example.com

# Database Configuration
POSTGRES_PASSWORD=your-secure-postgres-password
AUTH_DB_PASSWORD=your-secure-auth-db-password
AUTHENTICATOR_PASSWORD=your-secure-authenticator-password
REALTIME_DB_PASSWORD=your-secure-realtime-db-password

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-token-with-at-least-32-characters-long
JWT_EXPIRY=3600

# Redis Configuration
REDIS_PASSWORD=your-secure-redis-password

# Supabase Configuration
SUPABASE_ANON_KEY=your-supabase-anon-key
ADDITIONAL_REDIRECT_URLS=https://bookmark-sync.example.com/auth/callback

# OAuth Configuration (optional)
GITHUB_ENABLED=false
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_ENABLED=false
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Email Configuration
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASS=your-smtp-password
ADMIN_EMAIL=admin@example.com
MAILER_AUTOCONFIRM=false
MAILER_SECURE_EMAIL_CHANGE=true

# MinIO Configuration
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your-secure-minio-password
MINIO_BROWSER_REDIRECT_URL=https://bookmark-sync.example.com/storage
MINIO_SERVER_URL=https://bookmark-sync.example.com/storage

# Search Configuration
TYPESENSE_API_KEY=your-secure-typesense-api-key

# Realtime Configuration
REALTIME_ENC_KEY=your-secure-realtime-encryption-key
SECRET_KEY_BASE=your-super-secret-key-base-with-at-least-64-characters-long

# Docker Registry Configuration
DOCKER_REGISTRY=localhost:5000
API_VERSION=latest
WEB_VERSION=latest

# Security Configuration
DISABLE_SIGNUP=false

# PostgREST Configuration
PGRST_DB_SCHEMAS=public
PGRST_DB_ANON_ROLE=anon
EOF

    log_warning "Please edit $ENV_FILE with your actual configuration values"
    log_warning "Make sure to set secure passwords and proper domain names"
}

# Initialize Docker Swarm secrets
init_swarm_secrets() {
    log_step "Initializing Docker Swarm secrets..."

    # Source environment variables
    if [[ -f "$PROJECT_ROOT/$ENV_FILE" ]]; then
        source "$PROJECT_ROOT/$ENV_FILE"
    fi

    # Create secrets if they don't exist
    create_secret_if_not_exists "postgres_password" "$POSTGRES_PASSWORD"
    create_secret_if_not_exists "jwt_secret" "$JWT_SECRET"
    create_secret_if_not_exists "redis_password" "$REDIS_PASSWORD"
    create_secret_if_not_exists "typesense_api_key" "$TYPESENSE_API_KEY"
    create_secret_if_not_exists "minio_root_password" "$MINIO_ROOT_PASSWORD"

    log_success "Docker Swarm secrets initialized"
}

# Helper function to create secret if it doesn't exist
create_secret_if_not_exists() {
    local secret_name="$1"
    local secret_value="$2"

    if docker secret ls --format "{{.Name}}" | grep -q "^${secret_name}$"; then
        log_debug "Secret $secret_name already exists"
    else
        echo "$secret_value" | docker secret create "$secret_name" -
        log_info "Created secret: $secret_name"
    fi
}

# Label nodes for service placement
label_nodes() {
    log_step "Labeling nodes for service placement..."

    # Get current node ID (assuming single node for demo)
    local node_id=$(docker node ls --format "{{.ID}}" --filter "role=manager" | head -n1)

    if [[ -n "$node_id" ]]; then
        # Label manager node for database and cache
        docker node update --label-add database=true "$node_id"
        docker node update --label-add cache=true "$node_id"
        docker node update --label-add search=true "$node_id"
        docker node update --label-add storage=true "$node_id"
        docker node update --label-add zone=zone1 "$node_id"

        log_success "Labeled manager node: $node_id"
    fi

    # Label worker nodes if they exist
    local worker_nodes=$(docker node ls --format "{{.ID}}" --filter "role=worker")
    local zone_counter=1

    for worker_id in $worker_nodes; do
        docker node update --label-add zone="zone$zone_counter" "$worker_id"
        log_info "Labeled worker node $worker_id with zone$zone_counter"
        ((zone_counter++))
    done

    log_success "Node labeling completed"
}

# Create required directories
create_directories() {
    log_step "Creating required directories..."

    local base_dir="/opt/bookmark-sync"
    local dirs=(
        "$base_dir/data/postgres"
        "$base_dir/data/redis"
        "$base_dir/data/typesense"
        "$base_dir/data/minio"
        "$base_dir/cache/minio"
        "$base_dir/logs/nginx"
        "$base_dir/ssl"
        "$base_dir/backups"
    )

    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            sudo mkdir -p "$dir"
            sudo chown -R 1000:1000 "$dir"
            log_info "Created directory: $dir"
        else
            log_debug "Directory already exists: $dir"
        fi
    done

    log_success "Required directories created"
}

# Build and push images to registry
build_and_push_images() {
    log_step "Building and pushing images..."

    # Source environment variables
    if [[ -f "$PROJECT_ROOT/$ENV_FILE" ]]; then
        source "$PROJECT_ROOT/$ENV_FILE"
    fi

    local registry="${DOCKER_REGISTRY:-localhost:5000}"
    local api_version="${API_VERSION:-latest}"
    local web_version="${WEB_VERSION:-latest}"

    # Build API image
    log_info "Building API image..."
    docker build -t "${registry}/bookmark-sync-api:${api_version}" -f Dockerfile.prod .

    # Push API image
    log_info "Pushing API image..."
    docker push "${registry}/bookmark-sync-api:${api_version}"

    # Build web image if web directory exists
    if [[ -d "$PROJECT_ROOT/web" ]]; then
        log_info "Building web image..."
        docker build -t "${registry}/bookmark-sync-web:${web_version}" -f web/Dockerfile.prod web/

        log_info "Pushing web image..."
        docker push "${registry}/bookmark-sync-web:${web_version}"
    fi

    log_success "Images built and pushed successfully"
}

# Deploy stack to swarm
deploy_stack() {
    log_step "Deploying stack to Docker Swarm..."

    cd "$PROJECT_ROOT"

    # Deploy the stack
    docker stack deploy \
        --compose-file "$COMPOSE_FILE" \
        --with-registry-auth \
        "$STACK_NAME"

    log_success "Stack deployed successfully"
}

# Wait for services to be ready
wait_for_services() {
    log_step "Waiting for services to be ready..."

    local max_attempts=60
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        local running_services=$(docker stack services "$STACK_NAME" --format "{{.Replicas}}" | grep -c "/")
        local ready_services=$(docker stack services "$STACK_NAME" --format "{{.Replicas}}" | grep -c "^[0-9]*/[0-9]*$" | grep -v "0/")

        if [[ $running_services -eq $ready_services ]] && [[ $ready_services -gt 0 ]]; then
            log_success "All services are ready"
            return 0
        fi

        log_info "Waiting for services... ($((attempt + 1))/$max_attempts)"
        sleep 10
        ((attempt++))
    done

    log_warning "Some services may not be ready yet. Check with: docker stack services $STACK_NAME"
}

# Show deployment status
show_status() {
    log_step "Deployment Status"

    echo ""
    log_info "Stack Services:"
    docker stack services "$STACK_NAME"

    echo ""
    log_info "Node Information:"
    docker node ls

    echo ""
    log_info "Service Tasks:"
    docker stack ps "$STACK_NAME" --no-trunc

    echo ""
    log_info "Network Information:"
    docker network ls --filter "name=${STACK_NAME}"

    echo ""
    log_success "Deployment completed successfully!"
    log_info "Access your application at: https://${DOMAIN_NAME:-localhost}"
}

# Cleanup function
cleanup() {
    log_step "Cleaning up..."

    if [[ "${1:-}" == "remove" ]]; then
        log_warning "Removing stack: $STACK_NAME"
        docker stack rm "$STACK_NAME"

        log_info "Waiting for stack removal..."
        sleep 30

        log_success "Stack removed successfully"
    fi
}

# Health check
health_check() {
    log_step "Performing health check..."

    local services=(
        "${STACK_NAME}_api"
        "${STACK_NAME}_nginx"
        "${STACK_NAME}_supabase-db"
        "${STACK_NAME}_redis"
    )

    local healthy_count=0

    for service in "${services[@]}"; do
        local replicas=$(docker service ls --filter "name=$service" --format "{{.Replicas}}")
        if [[ "$replicas" =~ ^[1-9][0-9]*/[1-9][0-9]*$ ]]; then
            log_success "Service $service is healthy: $replicas"
            ((healthy_count++))
        else
            log_warning "Service $service may have issues: $replicas"
        fi
    done

    if [[ $healthy_count -eq ${#services[@]} ]]; then
        log_success "All core services are healthy"
    else
        log_warning "Some services may need attention"
    fi
}

# Scale services
scale_services() {
    log_step "Scaling services..."

    local api_replicas="${API_REPLICAS:-5}"
    local nginx_replicas="${NGINX_REPLICAS:-2}"
    local auth_replicas="${AUTH_REPLICAS:-2}"
    local rest_replicas="${REST_REPLICAS:-3}"

    docker service scale \
        "${STACK_NAME}_api=$api_replicas" \
        "${STACK_NAME}_nginx=$nginx_replicas" \
        "${STACK_NAME}_supabase-auth=$auth_replicas" \
        "${STACK_NAME}_supabase-rest=$rest_replicas"

    log_success "Services scaled successfully"
}

# Main function
main() {
    local command="${1:-deploy}"

    case "$command" in
        "deploy")
            log_info "Starting Docker Swarm deployment..."
            check_prerequisites
            init_swarm_secrets
            label_nodes
            create_directories
            build_and_push_images
            deploy_stack
            wait_for_services
            show_status
            ;;
        "status")
            show_status
            ;;
        "health")
            health_check
            ;;
        "scale")
            scale_services
            ;;
        "remove")
            cleanup "remove"
            ;;
        "secrets")
            init_swarm_secrets
            ;;
        "labels")
            label_nodes
            ;;
        *)
            echo "Usage: $0 {deploy|status|health|scale|remove|secrets|labels}"
            echo ""
            echo "Commands:"
            echo "  deploy  - Deploy the complete stack to Docker Swarm"
            echo "  status  - Show deployment status"
            echo "  health  - Perform health check"
            echo "  scale   - Scale services"
            echo "  remove  - Remove the stack"
            echo "  secrets - Initialize Docker secrets"
            echo "  labels  - Label nodes for service placement"
            exit 1
            ;;
    esac
}

# Handle script interruption
trap 'log_error "Script interrupted"; exit 1' INT TERM

# Run main function
main "$@"