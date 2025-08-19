# Bookmark Sync Service - Production Deployment

A comprehensive, secure, and scalable bookmark synchronization service with Docker Swarm orchestration.

## 🚀 Quick Start

### Prerequisites
- Docker 24.0+ with Swarm mode
- Domain name with DNS configured
- Ubuntu 20.04+ / CentOS 8+ / RHEL 8+
- 8GB+ RAM, 4+ CPU cores, 100GB+ SSD

### One-Command Setup
```bash
# Set your configuration
export DOMAIN_NAME=your-domain.com
export SSL_EMAIL=admin@your-domain.com

# Run automated setup
./scripts/setup-production.sh setup

# Deploy to production
./scripts/deploy-swarm.sh deploy
```

## 📋 Manual Setup

### 1. Clone Repository
```bash
git clone https://github.com/your-org/bookmark-sync-service.git
cd bookmark-sync-service
chmod +x scripts/*.sh
```

### 2. Initialize Docker Swarm
```bash
docker swarm init --advertise-addr YOUR_SERVER_IP
```

### 3. Configure Environment
```bash
# Copy and edit environment file
cp .env.production.example .env.production
nano .env.production
```

### 4. Setup Secrets
```bash
cd secrets/

# Generate secure secrets
openssl rand -base64 32 > postgres_password.txt
openssl rand -base64 64 > jwt_secret.txt
openssl rand -base64 32 > redis_password.txt
openssl rand -base64 32 > typesense_api_key.txt
openssl rand -base64 32 > minio_root_password.txt

# Secure permissions
chmod 600 *.txt
cd ..
```

### 5. Deploy Services
```bash
./scripts/deploy-swarm.sh deploy
```

## 🔒 Security Features

### Docker Swarm Secrets
- All sensitive data stored in Docker Swarm secrets
- Zero environment variable exposure
- Automatic secret rotation support
- Encrypted secret storage

### Network Security
- Encrypted overlay networks
- WSS (WebSocket Secure) connections
- SSL/TLS termination with Let's Encrypt
- Firewall configuration included

### Access Control
- Role-based access control (RBAC)
- JWT token authentication
- Rate limiting and circuit breakers
- Security headers and CORS policies

## 📊 Service Architecture

### Core Services
| Service | Purpose | Replicas | Resources |
|---------|---------|----------|-----------|
| **API** | Go backend service | 5 | 2GB RAM, 1 CPU |
| **Database** | PostgreSQL with Supabase | 1 | 4GB RAM, 2 CPU |
| **Cache** | Redis with persistence | 1 | 1GB RAM, 0.5 CPU |
| **Search** | Typesense with Chinese support | 1 | 2GB RAM, 1 CPU |
| **Storage** | MinIO S3-compatible storage | 1 | 2GB RAM, 1 CPU |
| **Load Balancer** | Nginx with SSL termination | 2 | 512MB RAM, 0.5 CPU |

### High Availability Features
- Automatic failover and restart policies
- Rolling updates with zero downtime
- Health checks and monitoring
- Horizontal scaling support
- Load balancing across replicas

## 🔧 Configuration

### Required Environment Variables
```bash
# Domain Configuration
DOMAIN_NAME=your-domain.com
SITE_URL=https://your-domain.com
SSL_EMAIL=admin@your-domain.com

# Database Configuration
AUTH_DB_PASSWORD=secure-auth-password
AUTHENTICATOR_PASSWORD=secure-authenticator-password
REALTIME_DB_PASSWORD=secure-realtime-password

# External Services
SUPABASE_ANON_KEY=your-supabase-anon-key
SMTP_HOST=smtp.your-provider.com
SMTP_USER=noreply@your-domain.com
SMTP_PASS=your-smtp-password
```

### Optional OAuth Configuration
```bash
# GitHub OAuth
GITHUB_ENABLED=true
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Google OAuth
GOOGLE_ENABLED=true
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

## 🛠️ Management Commands

### Deployment Operations
```bash
# Deploy stack
./scripts/deploy-swarm.sh deploy

# Check service health
./scripts/deploy-swarm.sh health

# Scale services
./scripts/deploy-swarm.sh scale

# Remove stack
./scripts/deploy-swarm.sh remove
```

### Secret Management
```bash
# Rotate all secrets
./scripts/rotate-secrets.sh rotate

# Rotate specific secret
./scripts/rotate-secrets.sh rotate postgres_password

# Verify secrets
./scripts/rotate-secrets.sh verify

# List secrets status
./scripts/rotate-secrets.sh list
```

### CI/CD Pipeline
```bash
# Run complete pipeline
./scripts/ci-cd-pipeline.sh pipeline

# Deploy to staging
./scripts/ci-cd-pipeline.sh deploy-staging

# Deploy to production
./scripts/ci-cd-pipeline.sh deploy-production

# Rollback deployment
./scripts/ci-cd-pipeline.sh rollback
```

## 📈 Monitoring and Maintenance

### Health Monitoring
```bash
# Service status
docker stack services bookmark-sync

# Service logs
docker service logs bookmark-sync_api

# Resource usage
docker stats

# Node status
docker node ls
```

### Backup Procedures
```bash
# Database backup
docker exec $(docker ps -q -f name=bookmark-sync_supabase-db) \
  pg_dump -U postgres postgres > backup_$(date +%Y%m%d_%H%M%S).sql

# Storage backup
docker exec $(docker ps -q -f name=bookmark-sync_minio) \
  mc mirror /data /backup/minio_$(date +%Y%m%d_%H%M%S)
```

### Scaling Operations
```bash
# Scale API service
docker service scale bookmark-sync_api=10

# Scale load balancer
docker service scale bookmark-sync_nginx=3

# Scale with environment variables
API_REPLICAS=10 NGINX_REPLICAS=3 ./scripts/deploy-swarm.sh scale
```

## 🔍 Troubleshooting

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

#### SSL Certificate Issues
```bash
# Check certificate status
sudo certbot certificates

# Renew certificates
sudo certbot renew

# Test SSL configuration
openssl s_client -connect your-domain.com:443
```

### Performance Tuning
```bash
# Monitor resource usage
docker stats

# Adjust resource limits
docker service update --limit-memory 4g bookmark-sync_api

# Scale based on load
docker service scale bookmark-sync_api=8
```

## 📚 Documentation

### Complete Guides
- **[Production Setup Guide](PRODUCTION_SETUP_GUIDE.md)**: Detailed deployment instructions
- **[Security Checklist](SECURITY_CHECKLIST.md)**: Security validation procedures
- **[Deployment Security Fixes](DEPLOYMENT_SECURITY_FIXES.md)**: Security improvements and fixes

### API Documentation
- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics`
- **API v1**: `GET /api/v1/*`

### Service Endpoints
- **Web Interface**: `https://your-domain.com`
- **API**: `https://your-domain.com/api/v1`
- **Storage Console**: `https://your-domain.com/storage`
- **Health Check**: `https://your-domain.com/health`

## 🆘 Support

### Log Collection
```bash
# Collect all service logs
mkdir -p logs
for service in api nginx supabase-db redis typesense minio; do
  docker service logs bookmark-sync_$service > logs/$service.log 2>&1
done

# Create support bundle
tar -czf support-bundle-$(date +%Y%m%d_%H%M%S).tar.gz \
  logs/ .env.production docker-compose.swarm.yml
```

### Contact Information
- **Issues**: Create an issue in the repository
- **Security**: security@your-domain.com
- **Support**: support@your-domain.com

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

---

**Note**: This is a production-ready deployment with enterprise-grade security, scalability, and monitoring. For development setup, see the main [README.md](README.md) file.