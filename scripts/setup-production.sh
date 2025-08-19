#!/bin/bash

# Production Setup Script for Bookmark Sync Service
# This script automates the initial setup process for production deployment

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
DOMAIN_NAME="${DOMAIN_NAME:-bookmark-sync.example.com}"
SSL_EMAIL="${SSL_EMAIL:-admin@example.com}"

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

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root for security reasons"
        log_info "Please run as a regular user with sudo privileges"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Installing Docker..."
        install_docker
    else
        log_success "Docker is installed"
    fi

    # Check if Docker Compose is available
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available. Please install Docker Compose."
        exit 1
    else
        log_success "Docker Compose is available"
    fi

    # Check if Git is installed
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed. Installing Git..."
        sudo apt-get update && sudo apt-get install -y git
    else
        log_success "Git is installed"
    fi

    # Check if OpenSSL is installed
    if ! command -v openssl &> /dev/null; then
        log_error "OpenSSL is not installed. Installing OpenSSL..."
        sudo apt-get update && sudo apt-get install -y openssl
    else
        log_success "OpenSSL is installed"
    fi

    log_success "Prerequisites check completed"
}

# Install Docker
install_docker() {
    log_info "Installing Docker..."

    # Update package index
    sudo apt-get update

    # Install packages to allow apt to use a repository over HTTPS
    sudo apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release

    # Add Docker's official GPG key
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

    # Set up the repository
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker Engine
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    # Add current user to docker group
    sudo usermod -aG docker $USER

    log_success "Docker installed successfully"
    log_warning "Please log out and log back in for Docker group changes to take effect"
}

# Initialize Docker Swarm
init_swarm() {
    log_step "Initializing Docker Swarm..."

    if docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
        log_success "Docker Swarm is already initialized"
        return 0
    fi

    # Get the primary IP address
    local primary_ip=$(ip route get 8.8.8.8 | awk '{print $7; exit}')

    log_info "Initializing Docker Swarm with advertise address: $primary_ip"
    docker swarm init --advertise-addr "$primary_ip"

    log_success "Docker Swarm initialized successfully"
}

# Setup environment configuration
setup_environment() {
    log_step "Setting up environment configuration..."

    cd "$PROJECT_ROOT"

    # Copy environment template if it doesn't exist
    if [[ ! -f ".env.production" ]]; then
        if [[ -f ".env.production.example" ]]; then
            cp .env.production.example .env.production
            log_success "Created .env.production from template"
        else
            log_error "Environment template not found"
            exit 1
        fi
    else
        log_info ".env.production already exists"
    fi

    # Update domain name if provided
    if [[ -n "$DOMAIN_NAME" && "$DOMAIN_NAME" != "bookmark-sync.example.com" ]]; then
        sed -i "s/bookmark-sync.example.com/$DOMAIN_NAME/g" .env.production
        log_success "Updated domain name to: $DOMAIN_NAME"
    fi

    # Update SSL email if provided
    if [[ -n "$SSL_EMAIL" && "$SSL_EMAIL" != "admin@example.com" ]]; then
        sed -i "s/admin@example.com/$SSL_EMAIL/g" .env.production
        log_success "Updated SSL email to: $SSL_EMAIL"
    fi

    log_warning "Please edit .env.production with your actual configuration values"
}

# Setup secrets
setup_secrets() {
    log_step "Setting up Docker Swarm secrets..."

    cd "$PROJECT_ROOT/secrets"

    # List of secrets to generate
    local secrets=(
        "postgres_password"
        "jwt_secret"
        "redis_password"
        "typesense_api_key"
        "minio_root_password"
    )

    for secret in "${secrets[@]}"; do
        local secret_file="${secret}.txt"

        if [[ ! -f "$secret_file" ]]; then
            log_info "Generating secure $secret..."

            # Generate different length secrets based on type
            case "$secret" in
                "jwt_secret")
                    openssl rand -base64 64 > "$secret_file"
                    ;;
                *)
                    openssl rand -base64 32 > "$secret_file"
                    ;;
            esac

            # Set secure permissions
            chmod 600 "$secret_file"
            chown $(whoami):$(whoami) "$secret_file"

            log_success "Generated $secret_file"
        else
            log_info "$secret_file already exists"
        fi
    done

    log_success "Secrets setup completed"
}

# Setup SSL certificates
setup_ssl() {
    log_step "Setting up SSL certificates..."

    # Check if certbot is installed
    if ! command -v certbot &> /dev/null; then
        log_info "Installing Certbot..."
        sudo apt-get update
        sudo apt-get install -y certbot python3-certbot-nginx
    fi

    # Check if nginx is installed
    if ! command -v nginx &> /dev/null; then
        log_info "Installing Nginx..."
        sudo apt-get update
        sudo apt-get install -y nginx
    fi

    # Create basic nginx configuration
    sudo tee /etc/nginx/sites-available/bookmark-sync > /dev/null <<EOF
server {
    listen 80;
    server_name $DOMAIN_NAME;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://\$server_name\$request_uri;
    }
}
EOF

    # Enable the site
    sudo ln -sf /etc/nginx/sites-available/bookmark-sync /etc/nginx/sites-enabled/
    sudo nginx -t && sudo systemctl reload nginx

    # Generate SSL certificate
    log_info "Generating SSL certificate for $DOMAIN_NAME..."
    sudo certbot --nginx -d "$DOMAIN_NAME" --email "$SSL_EMAIL" --agree-tos --non-interactive

    # Setup auto-renewal
    echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -

    log_success "SSL certificates configured"
}

# Configure firewall
setup_firewall() {
    log_step "Configuring firewall..."

    # Install UFW if not present
    if ! command -v ufw &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y ufw
    fi

    # Configure UFW rules
    sudo ufw --force reset
    sudo ufw default deny incoming
    sudo ufw default allow outgoing

    # Allow essential services
    sudo ufw allow 22/tcp comment 'SSH'
    sudo ufw allow 80/tcp comment 'HTTP'
    sudo ufw allow 443/tcp comment 'HTTPS'

    # Allow Docker Swarm ports
    sudo ufw allow 2377/tcp comment 'Docker Swarm management'
    sudo ufw allow 7946/tcp comment 'Docker Swarm communication'
    sudo ufw allow 7946/udp comment 'Docker Swarm communication'
    sudo ufw allow 4789/udp comment 'Docker Swarm overlay network'

    # Enable firewall
    sudo ufw --force enable

    log_success "Firewall configured"
}

# Create backup directories
setup_backup_dirs() {
    log_step "Creating backup directories..."

    local backup_dirs=(
        "/opt/bookmark-sync/backups/database"
        "/opt/bookmark-sync/backups/storage"
        "/opt/bookmark-sync/backups/config"
    )

    for dir in "${backup_dirs[@]}"; do
        sudo mkdir -p "$dir"
        sudo chown $(whoami):$(whoami) "$dir"
        log_info "Created backup directory: $dir"
    done

    log_success "Backup directories created"
}

# Setup monitoring
setup_monitoring() {
    log_step "Setting up basic monitoring..."

    # Install basic monitoring tools
    sudo apt-get update
    sudo apt-get install -y htop iotop nethogs

    # Setup log rotation
    sudo tee /etc/logrotate.d/bookmark-sync > /dev/null <<EOF
/opt/bookmark-sync/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $(whoami) $(whoami)
}
EOF

    log_success "Basic monitoring setup completed"
}

# Validate setup
validate_setup() {
    log_step "Validating setup..."

    local validation_errors=0

    # Check Docker Swarm
    if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
        log_error "Docker Swarm is not active"
        ((validation_errors++))
    fi

    # Check secrets
    cd "$PROJECT_ROOT/secrets"
    local required_secrets=("postgres_password.txt" "jwt_secret.txt" "redis_password.txt" "typesense_api_key.txt" "minio_root_password.txt")
    for secret in "${required_secrets[@]}"; do
        if [[ ! -f "$secret" ]]; then
            log_error "Secret file missing: $secret"
            ((validation_errors++))
        fi
    done

    # Check environment file
    if [[ ! -f "$PROJECT_ROOT/.env.production" ]]; then
        log_error "Environment file missing: .env.production"
        ((validation_errors++))
    fi

    # Check SSL certificate
    if [[ -n "$DOMAIN_NAME" && "$DOMAIN_NAME" != "bookmark-sync.example.com" ]]; then
        if ! sudo certbot certificates | grep -q "$DOMAIN_NAME"; then
            log_warning "SSL certificate not found for $DOMAIN_NAME"
        fi
    fi

    if [[ $validation_errors -eq 0 ]]; then
        log_success "Setup validation passed"
        return 0
    else
        log_error "Setup validation failed with $validation_errors errors"
        return 1
    fi
}

# Display next steps
show_next_steps() {
    log_step "Setup completed! Next steps:"

    echo ""
    echo -e "${GREEN}✅ Production setup completed successfully!${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "1. Review and update configuration:"
    echo "   nano $PROJECT_ROOT/.env.production"
    echo ""
    echo "2. Deploy the application:"
    echo "   cd $PROJECT_ROOT"
    echo "   ./scripts/deploy-swarm.sh deploy"
    echo ""
    echo "3. Monitor the deployment:"
    echo "   docker stack services bookmark-sync"
    echo "   docker service logs bookmark-sync_api"
    echo ""
    echo "4. Access your application:"
    echo "   https://$DOMAIN_NAME"
    echo ""
    echo -e "${CYAN}For detailed configuration options, see:${NC}"
    echo "- PRODUCTION_SETUP_GUIDE.md"
    echo "- SECURITY_CHECKLIST.md"
    echo "- DEPLOYMENT_SECURITY_FIXES.md"
    echo ""
    echo -e "${RED}Important:${NC}"
    echo "- Review all configuration files before deployment"
    echo "- Ensure DNS is properly configured for your domain"
    echo "- Test the deployment in a staging environment first"
    echo ""
}

# Main function
main() {
    local command="${1:-setup}"

    case "$command" in
        "setup")
            log_info "Starting production setup for Bookmark Sync Service..."
            check_root
            check_prerequisites
            init_swarm
            setup_environment
            setup_secrets
            setup_firewall
            setup_backup_dirs
            setup_monitoring
            if validate_setup; then
                show_next_steps
            else
                log_error "Setup validation failed. Please check the errors above."
                exit 1
            fi
            ;;
        "ssl")
            setup_ssl
            ;;
        "secrets")
            setup_secrets
            ;;
        "firewall")
            setup_firewall
            ;;
        "validate")
            validate_setup
            ;;
        *)
            echo "Usage: $0 {setup|ssl|secrets|firewall|validate}"
            echo ""
            echo "Commands:"
            echo "  setup     - Complete production setup"
            echo "  ssl       - Setup SSL certificates only"
            echo "  secrets   - Generate secrets only"
            echo "  firewall  - Configure firewall only"
            echo "  validate  - Validate current setup"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"