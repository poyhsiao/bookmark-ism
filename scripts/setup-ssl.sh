#!/bin/bash

# SSL Certificate Setup Script for Bookmark Sync Service
# This script sets up SSL certificates using Let's Encrypt

set -e

# Configuration
DOMAIN=${DOMAIN:-"localhost"}
EMAIL=${EMAIL:-"admin@example.com"}
NGINX_CONTAINER=${NGINX_CONTAINER:-"bookmark-nginx"}
CERTBOT_CONTAINER=${CERTBOT_CONTAINER:-"certbot"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if running in production mode
check_production() {
    if [[ "$DOMAIN" == "localhost" ]]; then
        warn "Domain is set to localhost. SSL certificates will be self-signed for development."
        return 1
    fi
    return 0
}

# Create SSL directory structure
create_ssl_directories() {
    log "Creating SSL directory structure..."
    mkdir -p nginx/ssl
    mkdir -p nginx/ssl/live
    mkdir -p nginx/ssl/archive
    chmod 755 nginx/ssl
}

# Generate self-signed certificates for development
generate_self_signed() {
    log "Generating self-signed certificates for development..."

    # Create private key
    openssl genrsa -out nginx/ssl/key.pem 2048

    # Create certificate signing request
    openssl req -new -key nginx/ssl/key.pem -out nginx/ssl/cert.csr -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"

    # Generate self-signed certificate
    openssl x509 -req -days 365 -in nginx/ssl/cert.csr -signkey nginx/ssl/key.pem -out nginx/ssl/cert.pem

    # Create chain file (same as cert for self-signed)
    cp nginx/ssl/cert.pem nginx/ssl/chain.pem

    # Set permissions
    chmod 600 nginx/ssl/key.pem
    chmod 644 nginx/ssl/cert.pem nginx/ssl/chain.pem

    log "Self-signed certificates generated successfully"
}

# Setup Let's Encrypt certificates for production
setup_letsencrypt() {
    log "Setting up Let's Encrypt certificates for domain: $DOMAIN"

    # Check if certbot is available
    if ! command -v certbot &> /dev/null; then
        log "Installing certbot..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install certbot
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y certbot
            elif command -v yum &> /dev/null; then
                sudo yum install -y certbot
            else
                error "Unable to install certbot. Please install it manually."
            fi
        fi
    fi

    # Stop nginx temporarily for certificate generation
    log "Stopping nginx for certificate generation..."
    docker-compose stop nginx || true

    # Generate certificates using standalone mode
    log "Generating Let's Encrypt certificates..."
    certbot certonly \
        --standalone \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --domains "$DOMAIN" \
        --cert-path nginx/ssl/cert.pem \
        --key-path nginx/ssl/key.pem \
        --chain-path nginx/ssl/chain.pem \
        --fullchain-path nginx/ssl/fullchain.pem

    # Copy certificates to nginx directory
    if [[ -d "/etc/letsencrypt/live/$DOMAIN" ]]; then
        cp "/etc/letsencrypt/live/$DOMAIN/cert.pem" nginx/ssl/cert.pem
        cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/key.pem
        cp "/etc/letsencrypt/live/$DOMAIN/chain.pem" nginx/ssl/chain.pem
        cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/fullchain.pem

        # Set permissions
        chmod 600 nginx/ssl/key.pem
        chmod 644 nginx/ssl/cert.pem nginx/ssl/chain.pem nginx/ssl/fullchain.pem

        log "Let's Encrypt certificates installed successfully"
    else
        error "Failed to generate Let's Encrypt certificates"
    fi
}

# Setup certificate renewal
setup_renewal() {
    log "Setting up certificate renewal..."

    # Create renewal script
    cat > scripts/renew-ssl.sh << 'EOF'
#!/bin/bash

# SSL Certificate Renewal Script
set -e

DOMAIN=${DOMAIN:-"localhost"}
NGINX_CONTAINER=${NGINX_CONTAINER:-"bookmark-nginx"}

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

if [[ "$DOMAIN" != "localhost" ]]; then
    log "Renewing SSL certificates for $DOMAIN..."

    # Renew certificates
    certbot renew --quiet

    # Copy renewed certificates
    if [[ -d "/etc/letsencrypt/live/$DOMAIN" ]]; then
        cp "/etc/letsencrypt/live/$DOMAIN/cert.pem" nginx/ssl/cert.pem
        cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/key.pem
        cp "/etc/letsencrypt/live/$DOMAIN/chain.pem" nginx/ssl/chain.pem
        cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/fullchain.pem

        # Set permissions
        chmod 600 nginx/ssl/key.pem
        chmod 644 nginx/ssl/cert.pem nginx/ssl/chain.pem nginx/ssl/fullchain.pem

        # Reload nginx
        docker-compose exec nginx nginx -s reload

        log "SSL certificates renewed successfully"
    fi
else
    log "Skipping renewal for localhost (development mode)"
fi
EOF

    chmod +x scripts/renew-ssl.sh

    # Add to crontab for automatic renewal (production only)
    if check_production; then
        log "Adding certificate renewal to crontab..."
        (crontab -l 2>/dev/null; echo "0 12 * * * $(pwd)/scripts/renew-ssl.sh >> $(pwd)/logs/ssl-renewal.log 2>&1") | crontab -
    fi
}

# Validate certificates
validate_certificates() {
    log "Validating SSL certificates..."

    if [[ -f "nginx/ssl/cert.pem" && -f "nginx/ssl/key.pem" ]]; then
        # Check certificate validity
        if openssl x509 -in nginx/ssl/cert.pem -text -noout > /dev/null 2>&1; then
            log "Certificate is valid"

            # Show certificate details
            log "Certificate details:"
            openssl x509 -in nginx/ssl/cert.pem -text -noout | grep -E "(Subject:|Issuer:|Not Before:|Not After:)"
        else
            error "Certificate validation failed"
        fi

        # Check private key
        if openssl rsa -in nginx/ssl/key.pem -check -noout > /dev/null 2>&1; then
            log "Private key is valid"
        else
            error "Private key validation failed"
        fi

        # Check if certificate and key match
        cert_hash=$(openssl x509 -noout -modulus -in nginx/ssl/cert.pem | openssl md5)
        key_hash=$(openssl rsa -noout -modulus -in nginx/ssl/key.pem | openssl md5)

        if [[ "$cert_hash" == "$key_hash" ]]; then
            log "Certificate and private key match"
        else
            error "Certificate and private key do not match"
        fi
    else
        error "SSL certificate files not found"
    fi
}

# Main execution
main() {
    log "Starting SSL certificate setup..."

    # Create directories
    create_ssl_directories

    # Setup certificates based on environment
    if check_production; then
        setup_letsencrypt
    else
        generate_self_signed
    fi

    # Setup renewal
    setup_renewal

    # Validate certificates
    validate_certificates

    log "SSL certificate setup completed successfully!"
    log "You can now start the services with: docker-compose up -d"
}

# Run main function
main "$@"