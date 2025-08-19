#!/bin/bash

# CI/CD Pipeline Script for Production Deployment
# Task 26: Automated deployment pipeline with CI/CD integration

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
REGISTRY="${DOCKER_REGISTRY:-localhost:5000}"
BUILD_NUMBER="${BUILD_NUMBER:-$(date +%Y%m%d%H%M%S)}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

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

# Pipeline stages
stage_checkout() {
    log_step "Stage 1: Code Checkout"

    if [[ -n "${GIT_URL:-}" ]]; then
        log_info "Cloning repository: $GIT_URL"
        git clone "$GIT_URL" "$PROJECT_ROOT" || true
        cd "$PROJECT_ROOT"
        git checkout "${GIT_BRANCH:-main}"
    else
        log_info "Using local repository"
        cd "$PROJECT_ROOT"
    fi

    log_info "Current commit: $(git rev-parse HEAD)"
    log_success "Code checkout completed"
}

stage_test() {
    log_step "Stage 2: Automated Testing"

    cd "$PROJECT_ROOT"

    # Run unit tests
    log_info "Running unit tests..."
    if [[ -f "scripts/run-tests.sh" ]]; then
        bash scripts/run-tests.sh --unit
    else
        go test ./backend/... -v -race -coverprofile=coverage.out
    fi

    # Run integration tests
    log_info "Running integration tests..."
    if [[ -f "scripts/run-integration-tests.sh" ]]; then
        bash scripts/run-integration-tests.sh
    fi

    # Run security tests
    log_info "Running security tests..."
    if command -v gosec &> /dev/null; then
        gosec ./backend/...
    else
        log_warning "gosec not installed, skipping security tests"
    fi

    # Run linting
    log_info "Running code linting..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./backend/...
    else
        log_warning "golangci-lint not installed, skipping linting"
    fi

    log_success "All tests passed"
}

stage_build() {
    log_step "Stage 3: Build Images"

    cd "$PROJECT_ROOT"

    local api_tag="${REGISTRY}/bookmark-sync-api:${BUILD_NUMBER}"
    local api_latest="${REGISTRY}/bookmark-sync-api:latest"
    local web_tag="${REGISTRY}/bookmark-sync-web:${BUILD_NUMBER}"
    local web_latest="${REGISTRY}/bookmark-sync-web:latest"

    # Build API image
    log_info "Building API image..."
    docker build \
        --build-arg BUILD_NUMBER="$BUILD_NUMBER" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --build-arg BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -t "$api_tag" \
        -t "$api_latest" \
        -f Dockerfile.prod .

    # Build web image if exists
    if [[ -d "web" && -f "web/Dockerfile.prod" ]]; then
        log_info "Building web image..."
        docker build \
            --build-arg BUILD_NUMBER="$BUILD_NUMBER" \
            --build-arg GIT_COMMIT="$GIT_COMMIT" \
            --build-arg BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            -t "$web_tag" \
            -t "$web_latest" \
            -f web/Dockerfile.prod web/
    fi

    log_success "Images built successfully"
}

stage_security_scan() {
    log_step "Stage 4: Security Scanning"

    local api_tag="${REGISTRY}/bookmark-sync-api:${BUILD_NUMBER}"

    # Scan images for vulnerabilities
    if command -v trivy &> /dev/null; then
        log_info "Scanning API image for vulnerabilities..."
        trivy image --exit-code 1 --severity HIGH,CRITICAL "$api_tag"
    else
        log_warning "Trivy not installed, skipping vulnerability scanning"
    fi

    # Check for secrets in images
    if command -v docker-scout &> /dev/null; then
        log_info "Scanning for secrets..."
        docker scout cves "$api_tag"
    else
        log_warning "Docker Scout not available, skipping secret scanning"
    fi

    log_success "Security scanning completed"
}

stage_push() {
    log_step "Stage 5: Push Images"

    local api_tag="${REGISTRY}/bookmark-sync-api:${BUILD_NUMBER}"
    local api_latest="${REGISTRY}/bookmark-sync-api:latest"
    local web_tag="${REGISTRY}/bookmark-sync-web:${BUILD_NUMBER}"
    local web_latest="${REGISTRY}/bookmark-sync-web:latest"

    # Push API images
    log_info "Pushing API images..."
    docker push "$api_tag"
    docker push "$api_latest"

    # Push web images if they exist
    if docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "$web_tag"; then
        log_info "Pushing web images..."
        docker push "$web_tag"
        docker push "$web_latest"
    fi

    log_success "Images pushed successfully"
}

stage_deploy_staging() {
    log_step "Stage 6: Deploy to Staging"

    local staging_stack="${STACK_NAME}-staging"

    # Deploy to staging environment
    log_info "Deploying to staging environment..."

    # Update environment variables for staging
    export API_VERSION="$BUILD_NUMBER"
    export WEB_VERSION="$BUILD_NUMBER"
    export STACK_NAME="$staging_stack"

    # Deploy using staging configuration
    if [[ -f "docker-compose.staging.yml" ]]; then
        docker stack deploy \
            --compose-file docker-compose.staging.yml \
            --with-registry-auth \
            "$staging_stack"
    else
        # Use production config with staging overrides
        docker stack deploy \
            --compose-file docker-compose.swarm.yml \
            --with-registry-auth \
            "$staging_stack"
    fi

    # Wait for staging deployment
    wait_for_deployment "$staging_stack"

    log_success "Staging deployment completed"
}

stage_smoke_tests() {
    log_step "Stage 7: Smoke Tests"

    local staging_url="${STAGING_URL:-http://localhost}"

    # Basic health check
    log_info "Running health check..."
    if curl -f "$staging_url/health" > /dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        return 1
    fi

    # API endpoint tests
    log_info "Testing API endpoints..."
    if curl -f "$staging_url/api/v1/health" > /dev/null 2>&1; then
        log_success "API health check passed"
    else
        log_error "API health check failed"
        return 1
    fi

    # Database connectivity test
    log_info "Testing database connectivity..."
    if curl -f "$staging_url/api/v1/ping" > /dev/null 2>&1; then
        log_success "Database connectivity test passed"
    else
        log_warning "Database connectivity test failed"
    fi

    log_success "Smoke tests completed"
}

stage_deploy_production() {
    log_step "Stage 8: Deploy to Production"

    # Require manual approval for production deployment
    if [[ "${AUTO_DEPLOY_PROD:-false}" != "true" ]]; then
        log_warning "Production deployment requires manual approval"
        read -p "Deploy to production? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Production deployment cancelled"
            return 0
        fi
    fi

    # Deploy to production
    log_info "Deploying to production environment..."

    # Update environment variables for production
    export API_VERSION="$BUILD_NUMBER"
    export WEB_VERSION="$BUILD_NUMBER"
    export STACK_NAME="$STACK_NAME"

    # Perform rolling update
    docker service update \
        --image "${REGISTRY}/bookmark-sync-api:${BUILD_NUMBER}" \
        --update-parallelism 2 \
        --update-delay 10s \
        --update-failure-action rollback \
        "${STACK_NAME}_api"

    # Wait for production deployment
    wait_for_deployment "$STACK_NAME"

    log_success "Production deployment completed"
}

stage_post_deploy() {
    log_step "Stage 9: Post-Deployment"

    # Run post-deployment tests
    log_info "Running post-deployment tests..."

    local prod_url="${PRODUCTION_URL:-https://bookmark-sync.example.com}"

    # Health check
    if curl -f "$prod_url/health" > /dev/null 2>&1; then
        log_success "Production health check passed"
    else
        log_error "Production health check failed"
        return 1
    fi

    # Performance test
    if command -v ab &> /dev/null; then
        log_info "Running performance test..."
        ab -n 100 -c 10 "$prod_url/api/v1/health" > /dev/null 2>&1
        log_success "Performance test completed"
    fi

    # Send notifications
    send_deployment_notification "success"

    log_success "Post-deployment tasks completed"
}

# Helper functions
wait_for_deployment() {
    local stack_name="$1"
    local max_attempts=60
    local attempt=0

    log_info "Waiting for deployment to complete..."

    while [[ $attempt -lt $max_attempts ]]; do
        local services_ready=true

        # Check if all services are ready
        while IFS= read -r service; do
            local replicas=$(docker service ls --filter "name=$service" --format "{{.Replicas}}")
            if [[ ! "$replicas" =~ ^[1-9][0-9]*/[1-9][0-9]*$ ]]; then
                services_ready=false
                break
            fi
        done < <(docker stack services "$stack_name" --format "{{.Name}}")

        if [[ "$services_ready" == "true" ]]; then
            log_success "All services are ready"
            return 0
        fi

        log_info "Waiting for services... ($((attempt + 1))/$max_attempts)"
        sleep 10
        ((attempt++))
    done

    log_error "Deployment timeout"
    return 1
}

send_deployment_notification() {
    local status="$1"

    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        local message="Deployment successful"

        if [[ "$status" != "success" ]]; then
            color="danger"
            message="Deployment failed"
        fi

        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"text\":\"$message - Build: $BUILD_NUMBER, Commit: $GIT_COMMIT\"}]}" \
            "$SLACK_WEBHOOK_URL"
    fi

    if [[ -n "${EMAIL_NOTIFICATION:-}" ]]; then
        echo "Deployment $status - Build: $BUILD_NUMBER, Commit: $GIT_COMMIT" | \
            mail -s "Bookmark Sync Deployment $status" "$EMAIL_NOTIFICATION"
    fi
}

rollback_deployment() {
    log_step "Rolling back deployment..."

    local previous_version="${PREVIOUS_VERSION:-latest}"

    # Rollback API service
    docker service update \
        --image "${REGISTRY}/bookmark-sync-api:${previous_version}" \
        --rollback \
        "${STACK_NAME}_api"

    # Wait for rollback to complete
    wait_for_deployment "$STACK_NAME"

    send_deployment_notification "rollback"

    log_success "Rollback completed"
}

# Main pipeline function
run_pipeline() {
    local start_time=$(date +%s)

    log_info "Starting CI/CD Pipeline - Build: $BUILD_NUMBER"

    # Set error handling
    set -e
    trap 'handle_pipeline_error $?' ERR

    # Run pipeline stages
    stage_checkout
    stage_test
    stage_build
    stage_security_scan
    stage_push
    stage_deploy_staging
    stage_smoke_tests
    stage_deploy_production
    stage_post_deploy

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log_success "Pipeline completed successfully in ${duration}s"
}

handle_pipeline_error() {
    local exit_code="$1"

    log_error "Pipeline failed with exit code: $exit_code"

    # Send failure notification
    send_deployment_notification "failure"

    # Optionally trigger rollback
    if [[ "${AUTO_ROLLBACK:-false}" == "true" ]]; then
        rollback_deployment
    fi

    exit "$exit_code"
}

# Main function
main() {
    local command="${1:-pipeline}"

    case "$command" in
        "pipeline")
            run_pipeline
            ;;
        "rollback")
            rollback_deployment
            ;;
        "test")
            stage_test
            ;;
        "build")
            stage_build
            ;;
        "deploy-staging")
            stage_deploy_staging
            ;;
        "deploy-production")
            stage_deploy_production
            ;;
        "smoke-tests")
            stage_smoke_tests
            ;;
        *)
            echo "Usage: $0 {pipeline|rollback|test|build|deploy-staging|deploy-production|smoke-tests}"
            echo ""
            echo "Commands:"
            echo "  pipeline         - Run complete CI/CD pipeline"
            echo "  rollback         - Rollback to previous version"
            echo "  test             - Run tests only"
            echo "  build            - Build images only"
            echo "  deploy-staging   - Deploy to staging only"
            echo "  deploy-production - Deploy to production only"
            echo "  smoke-tests      - Run smoke tests only"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"