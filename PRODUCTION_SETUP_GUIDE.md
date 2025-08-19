# Production Setup Guide

This comprehensive guide walks you through setting up the Bookmark Sync Service in a production environment with Docker Swarm.

## 📋 Prerequisites

### System Requirements
- **OS**: Ubuntu 20.04+ / CentOS 8+ / RHEL 8+
- **RAM**: Minimum 8GB, Recommended 16GB+
- **CPU**: Minimum 4 cores, Recommended 8+ cores
- **Storage**: Minimum 100GB SSD, Recommended 500GB+ SSD
- **Network**: Static IP address, Domain name configured

### Software Requirements
- Docker 24.0+
- Docker Compose 2.20+
- Git 2.30+
- OpenSSL 1.1.1+
- Nginx (for SSL termination)

## 🚀 Quick Start

### 1. Clone and Setup Repository
```bash
# Clone the repository
git clone https://github.com/your-org/bookmark-sync-service.git
cd bookmark-sync-service

# Make scripts executable
chmod +x scripts/*.sh
```

### 2. Initialize Docker Swarm
```bash
# Initialize Docker Swarm (replace with your server IP)
docker swarm init --advertise-addr YOUR_SERVER_IP

# Verify swarm status
docker node ls
```

### 3. Configure Environment
```bash
# Copy environment template
cp .env.production.example .env.production

# Edit with your actual values
nano .env.production
```

### 4. Setup Secrets
```bash
# Navigate to secrets directory
cd secrets/

# Copy example files
cp postgres_password.txt.example postgres_password.txt
cp jwt_secret.txt.example jwt_secret.txt
cp redis_password.txt.example redis_password.txt
cp typesense_api_key.txt.example typesense_api_key.txt
cp minio_root_password.txt.example minio_root_password.txt

# Generate secure passwords
openssl rand -base64 32 > postgres_password.txt
openssl rand -base64 64 > jwt_secret.txt
openssl rand -base64 32 > redis_password.txt
openssl rand -base64 32 > typesense_api_key.txt
openssl rand -base64 32 > minio_root_password.txt

# Secure the files
chmod 600 *.txt
cd ..
```

### 5. Deploy to Production
```bash
# Run the deployment script
./scripts/deploy-swarm.sh deploy

# Monitor deployment
docker stack services bookmark-sync
```

## 🔧 Detailed Configuration

### Environment Variables

#### Required Configuration
Update these values in `.env.production`:

```bash
# Domain Configuration
DOMAIN_NAME=your-domain.com
SITE_URL=https://your-domain.com
SSL_EMAIL=admin@your-domain.com

# Database Passwords (used for service connections)
AUTH_DB_PASSWORD=your-secure-auth-password
AUTHENTICATOR_PASSWORD=your-secure-authenticator-password
REALTIME_DB_PASSWORD=your-secure-realtime-password

# Supabase Configuration
SUPABASE_ANON_KEY=your-supabase-anon-key

# Email Configuration
SMTP_HOST=smtp.your-provider.com
SMTP_PORT=587
SMTP_USER=noreply@your-domain.com
SMTP_PASS=your-smtp-password
ADMIN_EMAIL=admin@your-domain.com

# Storage Configuration
MINIO_ROOT_USER=your-minio-admin-user
MINIO_BROWSER_REDIRECT_URL=https://your-domain.com/storage
MINIO_SERVER_URL=https://your-domain.com/storage

# Realtime Configuration
REALTIME_ENC_KEY=your-32-char-encryption-key
SECRET_KEY_BASE=your-64-char-secret-key-base
```

#### Optional OAuth Configuration
```bash
# GitHub OAuth (optional)
GITHUB_ENABLED=true
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Google OAuth (optional)
GOOGLE_ENABLED=true
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

### Docker Swarm Secrets

The system uses Docker Swarm secrets for sensitive data:

| Secret Name | Purpose | File Location |
|-------------|---------|---------------|
| `postgres_password` | PostgreSQL database password | `secrets/postgres_password.txt` |
| `jwt_secret` | JWT token signing secret | `secrets/jwt_secret.txt` |
| `redis_password` | Redis authentication password | `secrets/redis_password.txt` |
| `typesense_api_key` | Typesense search API key | `secrets/typesense_api_key.txt` |
| `minio_root_password` | MinIO storage admin password | `secrets/minio_root_password.txt` |

### Service Configuration

#### Node Labels
The deployment script automatically labels nodes for service placement:

```bash
# Manager node labels
docker node update --label-add database=true NODE_ID
docker node update --label-add cache=true NODE_ID
docker node update --label-add search=true NODE_ID
docker node update --label-add storage=true NODE_ID

# Worker node labels
docker node update --label-add zone=zone1 NODE_ID
```

#### Resource Allocation
Default resource limits per service:

| Service | Memory Limit | CPU Limit | Replicas |
|---------|--------------|-----------|----------|
| API | 2GB | 1.0 | 5 |
| Database | 4GB | 2.0 | 1 |
| Redis | 1GB | 0.5 | 1 |
| Nginx | 512MB | 0.5 | 2 |
| Typesense | 2GB | 1.0 | 1 |
| MinIO | 2GB | 1.0 | 1 |

## 🔒 Security Configuration

### SSL/TLS Setup
```bash
# Install Certbot for Let's Encrypt
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx

# Generate SSL certificates
sudo certbot --nginx -d your-domain.com

# Setup auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Firewall Configuration
```bash
# Configure UFW firewall
sudo ufw enable
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 2377/tcp  # Docker Swarm management
sudo ufw allow 7946/tcp  # Docker Swarm communication
sudo ufw allow 4789/udp  # Docker Swarm overlay network
```

### Security Headers
The Nginx configuration includes security headers:
- HSTS (HTTP Strict Transport Security)
- Content Security Policy
- X-Frame-Options
- X-Content-Type-Options
- Referrer-Policy

## 📊 Monitoring and Maintenance

### Health Checks
```bash
# Check service health
./scripts/deploy-swarm.sh health

# View service logs
docker service logs bookmark-sync_api
docker service logs bookmark-sync_nginx

# Monitor resource usage
docker stats
```

### Backup Procedures
```bash
# Database backup
docker exec $(docker ps -q -f name=bookmark-sync_supabase-db) \
  pg_dump -U postgres postgres > backup_$(date +%Y%m%d_%H%M%S).sql

# MinIO backup
docker exec $(docker ps -q -f name=bookmark-sync_minio) \
  mc mirror /data /backup/minio_$(date +%Y%m%d_%H%M%S)
```

### Scaling Services
```bash
# Scale API service
docker service scale bookmark-sync_api=10

# Scale Nginx load balancer
docker service scale bookmark-sync_nginx=3

# Or use the deployment script
API_REPLICAS=10 NGINX_REPLICAS=3 ./scripts/deploy-swarm.sh scale
```

## 🔄 CI/CD Integration

### Environment Variables for CI/CD
```bash
# CI/CD Configuration
AUTO_DEPLOY_PROD=false
NON_INTERACTIVE=true
AUTO_ROLLBACK=true

# Notification Settings
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
EMAIL_NOTIFICATION=devops@your-domain.com
```

### Pipeline Integration
```bash
# Run CI/CD pipeline
./scripts/ci-cd-pipeline.sh pipeline

# Deploy to staging
./scripts/ci-cd-pipeline.sh deploy-staging

# Deploy to production (with approval)
./scripts/ci-cd-pipeline.sh deploy-production
```

## 🚨 Troubleshooting

### Common Issues

#### Services Not Starting
```bash
# Check service status
docker service ps bookmark-sync_api --no-trunc

# Check logs
docker service logs bookmark-sync_api

# Restart service
docker service update --force bookmark-sync_api
```

#### Secret Access Issues
```bash
# List secrets
docker secret ls

# Recreate secret
docker secret rm postgres_password
docker secret create postgres_password ./secrets/postgres_password.txt
```

#### Network Connectivity Issues
```bash
# Check network
docker network ls
docker network inspect bookmark-sync_bookmark-network

# Test connectivity
docker exec -it $(docker ps -q -f name=bookmark-sync_api) \
  ping supabase-db
```

### Performance Tuning
```bash
# Monitor resource usage
docker stats

# Adjust resource limits
docker service update --limit-memory 4g bookmark-sync_api

# Scale services based on load
docker service scale bookmark-sync_api=8
```

## 📞 Support

### Log Collection
```bash
# Collect all service logs
mkdir -p logs
for service in api nginx supabase-db redis typesense minio; do
  docker service logs bookmark-sync_$service > logs/$service.log 2>&1
done

# Create support bundle
tar -czf support-bundle-$(date +%Y%m%d_%H%M%S).tar.gz logs/ .env.production docker-compose.swarm.yml
```

### Health Check Endpoints
- API Health: `https://your-domain.com/health`
- API Metrics: `https://your-domain.com/metrics`
- Database Status: Check via API health endpoint
- Storage Status: `https://your-domain.com/storage/minio/health/live`

For additional support, please refer to the project documentation or create an issue in the repository.