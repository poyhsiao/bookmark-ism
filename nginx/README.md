# Nginx Load Balancer and Gateway

This directory contains the Nginx configuration for the Bookmark Sync Service, providing load balancing, SSL termination, rate limiting, and reverse proxy functionality.

## 📁 Directory Structure

```
nginx/
├── nginx.conf              # Development configuration
├── nginx.prod.conf         # Production configuration with SSL
├── conf.d/                 # Modular configuration files
│   ├── ssl.conf           # SSL/TLS configuration
│   ├── security.conf      # Security headers and rate limiting
│   ├── cache.conf         # Caching configuration
│   └── upstream.conf      # Upstream server definitions
├── ssl/                   # SSL certificates (created by setup-ssl.sh)
└── README.md              # This file
```

## 🚀 Features

### Load Balancing
- **Algorithm**: Least connections with keepalive
- **Health Checks**: Automatic failover with configurable thresholds
- **Scaling**: Support for multiple API instances
- **Session Persistence**: Sticky sessions for WebSocket connections

### SSL/TLS Security
- **Protocols**: TLS 1.2 and 1.3 only
- **Ciphers**: Modern cipher suites with perfect forward secrecy
- **HSTS**: HTTP Strict Transport Security enabled
- **OCSP Stapling**: Certificate validation optimization

### Rate Limiting
- **API Endpoints**: 20 requests/second with burst capacity
- **Authentication**: 10 requests/second for auth endpoints
- **File Uploads**: 5 requests/second for upload endpoints
- **WebSocket**: 30 requests/second for real-time connections

### Security Headers
- **X-Frame-Options**: Clickjacking protection
- **X-XSS-Protection**: Cross-site scripting protection
- **X-Content-Type-Options**: MIME type sniffing protection
- **Content-Security-Policy**: Comprehensive CSP policy
- **Referrer-Policy**: Referrer information control

### Caching
- **API Responses**: 5-minute cache for GET requests
- **Static Assets**: 1-year cache with immutable headers
- **Cache Bypass**: Automatic bypass for authenticated requests
- **Cache Revalidation**: Background updates for stale content

## 🔧 Configuration Files

### nginx.conf (Development)
- Single API instance
- HTTP only
- Relaxed rate limiting
- Debug logging enabled

### nginx.prod.conf (Production)
- Multiple API instances
- HTTPS with SSL termination
- Strict rate limiting
- Optimized for performance

### conf.d/ssl.conf
- SSL session configuration
- Modern TLS settings
- OCSP stapling setup
- Security optimizations

### conf.d/security.conf
- Rate limiting zones
- Security headers
- Attack pattern blocking
- Real IP configuration

### conf.d/cache.conf
- Proxy cache paths
- Cache key configuration
- Cache bypass rules
- Performance optimizations

### conf.d/upstream.conf
- Backend server definitions
- Health check configuration
- Load balancing algorithms
- Connection pooling

## 🛠️ Setup and Configuration

### 1. SSL Certificate Setup

For development (self-signed certificates):
```bash
./scripts/setup-ssl.sh
```

For production (Let's Encrypt):
```bash
DOMAIN=your-domain.com EMAIL=admin@your-domain.com ./scripts/setup-ssl.sh
```

### 2. Performance Tuning

Generate optimized configuration:
```bash
./scripts/nginx-performance-tuning.sh optimize
```

Run performance benchmark:
```bash
./scripts/nginx-performance-tuning.sh benchmark
```

### 3. Health Monitoring

Run health checks:
```bash
./scripts/nginx-health-check.sh check
```

Continuous monitoring:
```bash
./scripts/nginx-health-check.sh monitor
```

### 4. Testing

Comprehensive test suite:
```bash
./scripts/test-nginx.sh test
```

Quick functionality test:
```bash
./scripts/test-nginx.sh quick
```

Security-focused tests:
```bash
./scripts/test-nginx.sh security
```

## 🔄 Upstream Services

### API Backend
- **Service**: Go API server
- **Port**: 8080
- **Health Check**: `/api/v1/health`
- **Load Balancing**: Least connections

### Supabase Auth
- **Service**: GoTrue authentication
- **Port**: 9999
- **Health Check**: `/health`
- **Rate Limiting**: 10 req/s

### Supabase REST
- **Service**: PostgREST API
- **Port**: 3000
- **Health Check**: `/`
- **Caching**: 5-minute cache

### Supabase Realtime
- **Service**: Realtime subscriptions
- **Port**: 4000
- **Protocol**: WebSocket
- **Health Check**: `/api/health`

## 📊 Monitoring and Metrics

### Health Endpoints
- **Nginx Health**: `GET /health`
- **Nginx Status**: `GET http://localhost:8080/nginx_status` (internal)
- **Nginx Metrics**: `GET http://localhost:8080/nginx_metrics` (Prometheus format)

### Log Files
- **Access Log**: `/var/log/nginx/access.log`
- **Error Log**: `/var/log/nginx/error.log`
- **Format**: JSON with performance metrics

### Key Metrics
- Active connections
- Requests per second
- Response times
- Cache hit rates
- Upstream health status

## 🔒 Security Features

### Rate Limiting
```nginx
# API endpoints: 20 req/s with burst of 50
limit_req zone=api burst=50 nodelay;

# Auth endpoints: 10 req/s with burst of 20
limit_req zone=auth burst=20 nodelay;

# Upload endpoints: 5 req/s with burst of 10
limit_req zone=upload burst=10 nodelay;
```

### Security Headers
```nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Content-Security-Policy "default-src 'self'; ..." always;
```

### Attack Protection
- SQL injection pattern blocking
- XSS attempt blocking
- Directory traversal protection
- Suspicious user agent blocking

## 🚀 Production Deployment

### Environment Variables
```bash
# SSL Configuration
DOMAIN=your-domain.com
EMAIL=admin@your-domain.com

# Performance Tuning
WORKER_PROCESSES=auto
WORKER_CONNECTIONS=2048
KEEPALIVE_TIMEOUT=65

# Container Configuration
NGINX_CONTAINER=bookmark-nginx
```

### Docker Compose Integration
```yaml
nginx:
  image: nginx:1.25-alpine
  container_name: bookmark-nginx
  volumes:
    - ./nginx/nginx.prod.conf:/etc/nginx/nginx.conf:ro
    - ./nginx/conf.d:/etc/nginx/conf.d:ro
    - ./nginx/ssl:/etc/nginx/ssl:ro
  ports:
    - "80:80"
    - "443:443"
  depends_on:
    - api
  networks:
    - bookmark-network
```

### Scaling Configuration
For horizontal scaling, update upstream configuration:
```nginx
upstream api_backend_prod {
    least_conn;
    server api_1:8080 max_fails=3 fail_timeout=30s weight=1;
    server api_2:8080 max_fails=3 fail_timeout=30s weight=1;
    server api_3:8080 max_fails=3 fail_timeout=30s weight=1;
    keepalive 64;
}
```

## 🔧 Troubleshooting

### Common Issues

1. **SSL Certificate Errors**
   ```bash
   # Check certificate validity
   openssl x509 -in nginx/ssl/cert.pem -text -noout

   # Regenerate certificates
   ./scripts/setup-ssl.sh
   ```

2. **Rate Limiting Too Aggressive**
   ```bash
   # Check rate limit zones in conf.d/security.conf
   # Adjust burst values as needed
   ```

3. **Upstream Connection Errors**
   ```bash
   # Check upstream health
   ./scripts/nginx-health-check.sh check

   # Verify service connectivity
   docker-compose ps
   ```

4. **Performance Issues**
   ```bash
   # Run performance analysis
   ./scripts/nginx-performance-tuning.sh benchmark

   # Check system resources
   ./scripts/nginx-performance-tuning.sh optimize
   ```

### Log Analysis
```bash
# Check error logs
docker-compose logs nginx

# Monitor access patterns
tail -f nginx/logs/access.log | grep -E "(4[0-9]{2}|5[0-9]{2})"

# Analyze performance
awk '{print $NF}' nginx/logs/access.log | sort -n | tail -10
```

## 📈 Performance Optimization

### System-Level Optimizations
```bash
# Increase file descriptor limits
ulimit -n 65535

# Optimize kernel parameters
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
sysctl -p
```

### Nginx-Level Optimizations
- Worker processes = CPU cores
- Worker connections = 2048 per worker
- Keepalive connections enabled
- Gzip compression for text content
- Open file cache enabled
- Sendfile and TCP optimizations

### Application-Level Optimizations
- API response caching
- Static asset optimization
- Database query optimization
- Connection pooling
- Background job processing

## 🔄 Maintenance

### Certificate Renewal
```bash
# Manual renewal
./scripts/renew-ssl.sh

# Automatic renewal (cron job)
0 12 * * * /path/to/scripts/renew-ssl.sh >> /path/to/logs/ssl-renewal.log 2>&1
```

### Configuration Updates
```bash
# Test configuration
docker exec bookmark-nginx nginx -t

# Reload without downtime
docker exec bookmark-nginx nginx -s reload

# Full restart if needed
docker-compose restart nginx
```

### Log Rotation
```bash
# Manual log rotation
docker exec bookmark-nginx nginx -s reopen

# Automatic rotation (logrotate)
/var/log/nginx/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 nginx nginx
    postrotate
        docker exec bookmark-nginx nginx -s reopen
    endscript
}
```

## 📚 Additional Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [SSL/TLS Best Practices](https://wiki.mozilla.org/Security/Server_Side_TLS)
- [Rate Limiting Guide](https://www.nginx.com/blog/rate-limiting-nginx/)
- [Performance Tuning](https://www.nginx.com/blog/tuning-nginx/)
- [Security Headers](https://securityheaders.com/)

## 🤝 Contributing

When modifying nginx configuration:

1. Test changes in development first
2. Run the test suite: `./scripts/test-nginx.sh`
3. Validate configuration: `nginx -t`
4. Update documentation if needed
5. Monitor performance after deployment

## 📄 License

This configuration is part of the Bookmark Sync Service project and follows the same license terms.