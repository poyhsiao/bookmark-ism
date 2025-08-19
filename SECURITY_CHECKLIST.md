# Security Checklist for Production Deployment

This checklist ensures your Bookmark Sync Service deployment follows security best practices.

## 🔐 Pre-Deployment Security Checklist

### ✅ Secrets Management
- [ ] All sensitive data moved from environment variables to Docker Swarm secrets
- [ ] Secret files created with secure permissions (600)
- [ ] Strong passwords generated (32+ characters, random)
- [ ] JWT secret is 64+ characters long
- [ ] Actual secret files are gitignored
- [ ] Example secret files are properly documented

### ✅ SSL/TLS Configuration
- [ ] Valid SSL certificate installed (Let's Encrypt or commercial)
- [ ] HTTPS enforced for all connections
- [ ] WebSocket connections use WSS (secure WebSocket)
- [ ] HTTP to HTTPS redirect configured
- [ ] HSTS headers enabled
- [ ] SSL certificate auto-renewal configured

### ✅ Network Security
- [ ] Firewall configured (only necessary ports open)
- [ ] Docker Swarm overlay network encrypted
- [ ] Internal service communication secured
- [ ] Database connections use SSL/TLS
- [ ] Redis authentication enabled
- [ ] MinIO access keys configured

### ✅ Authentication & Authorization
- [ ] Strong admin passwords set
- [ ] OAuth providers configured (if used)
- [ ] JWT expiration times appropriate
- [ ] User registration controls configured
- [ ] Rate limiting enabled
- [ ] Session management secure

### ✅ Database Security
- [ ] Database passwords stored in secrets
- [ ] Database user permissions minimal
- [ ] Database backups encrypted
- [ ] Connection pooling configured
- [ ] SQL injection protection enabled

## 🛡️ Runtime Security Checklist

### ✅ Container Security
- [ ] Containers run as non-root users
- [ ] Container images regularly updated
- [ ] Vulnerability scanning enabled
- [ ] Resource limits configured
- [ ] Security contexts properly set
- [ ] Privileged containers avoided

### ✅ Access Control
- [ ] SSH key-based authentication
- [ ] Sudo access restricted
- [ ] Service accounts properly configured
- [ ] File permissions correctly set
- [ ] Log access controlled

### ✅ Monitoring & Logging
- [ ] Security event logging enabled
- [ ] Log aggregation configured
- [ ] Intrusion detection active
- [ ] Failed login attempts monitored
- [ ] Unusual activity alerts set up

## 🔍 Security Validation Commands

### Check Secret Configuration
```bash
# Verify secrets exist
docker secret ls

# Check secret file permissions
ls -la secrets/*.txt

# Verify no secrets in environment variables
docker service inspect bookmark-sync_api | grep -i password
```

### Validate SSL/TLS
```bash
# Test SSL certificate
openssl s_client -connect your-domain.com:443 -servername your-domain.com

# Check SSL rating
curl -s "https://api.ssllabs.com/api/v3/analyze?host=your-domain.com"

# Verify HSTS headers
curl -I https://your-domain.com
```

### Network Security Tests
```bash
# Check open ports
nmap -sS your-server-ip

# Verify firewall status
sudo ufw status verbose

# Test internal network isolation
docker exec -it $(docker ps -q -f name=bookmark-sync_api) nmap -sn 10.0.1.0/24
```

### Authentication Tests
```bash
# Test rate limiting
for i in {1..100}; do curl -s https://your-domain.com/api/v1/auth/login; done

# Verify JWT expiration
curl -H "Authorization: Bearer expired-token" https://your-domain.com/api/v1/bookmarks

# Test OAuth endpoints
curl https://your-domain.com/auth/v1/authorize?provider=github
```

## 🚨 Security Incident Response

### Immediate Actions
1. **Isolate affected systems**
   ```bash
   # Remove compromised node from swarm
   docker node update --availability drain NODE_ID
   ```

2. **Rotate compromised secrets**
   ```bash
   # Generate new secret
   openssl rand -base64 32 > secrets/new_secret.txt

   # Update Docker secret
   docker secret create new_secret_name secrets/new_secret.txt
   docker service update --secret-rm old_secret --secret-add new_secret service_name
   ```

3. **Review access logs**
   ```bash
   # Check authentication logs
   docker service logs bookmark-sync_supabase-auth | grep -i "failed\|error\|unauthorized"

   # Review API access logs
   docker service logs bookmark-sync_nginx | grep -E "40[0-9]|50[0-9]"
   ```

### Recovery Procedures
1. **Restore from backup**
   ```bash
   # Stop affected services
   docker service scale bookmark-sync_api=0

   # Restore database
   docker exec -i $(docker ps -q -f name=bookmark-sync_supabase-db) \
     psql -U postgres postgres < backup.sql

   # Restart services
   docker service scale bookmark-sync_api=5
   ```

2. **Update security measures**
   ```bash
   # Update all secrets
   ./scripts/rotate-secrets.sh

   # Apply security patches
   docker service update --image new-secure-image bookmark-sync_api
   ```

## 📋 Regular Security Maintenance

### Weekly Tasks
- [ ] Review security logs
- [ ] Check for failed login attempts
- [ ] Verify backup integrity
- [ ] Update security patches
- [ ] Review user access

### Monthly Tasks
- [ ] Rotate secrets and passwords
- [ ] Update SSL certificates (if needed)
- [ ] Security vulnerability scan
- [ ] Review firewall rules
- [ ] Audit user permissions

### Quarterly Tasks
- [ ] Full security assessment
- [ ] Penetration testing
- [ ] Disaster recovery testing
- [ ] Security policy review
- [ ] Staff security training

## 🔧 Security Tools Integration

### Vulnerability Scanning
```bash
# Scan container images
trivy image bookmark-sync-api:latest

# Scan for secrets in code
git-secrets --scan

# Network vulnerability scan
nmap -sV --script vuln your-domain.com
```

### Security Monitoring
```bash
# Setup fail2ban for SSH protection
sudo apt-get install fail2ban
sudo systemctl enable fail2ban

# Configure log monitoring
sudo apt-get install logwatch
sudo logwatch --detail high --mailto admin@your-domain.com
```

### Backup Security
```bash
# Encrypt backups
gpg --symmetric --cipher-algo AES256 backup.sql

# Verify backup integrity
sha256sum backup.sql > backup.sql.sha256
sha256sum -c backup.sql.sha256
```

## 📞 Security Contacts

### Internal Contacts
- **Security Team**: security@your-domain.com
- **DevOps Team**: devops@your-domain.com
- **On-Call Engineer**: +1-XXX-XXX-XXXX

### External Resources
- **Docker Security**: https://docs.docker.com/engine/security/
- **OWASP Guidelines**: https://owasp.org/
- **CVE Database**: https://cve.mitre.org/

## 📝 Compliance Requirements

### Data Protection
- [ ] GDPR compliance (if applicable)
- [ ] Data encryption at rest
- [ ] Data encryption in transit
- [ ] User data export capability
- [ ] Right to be forgotten implementation

### Audit Requirements
- [ ] Access logging enabled
- [ ] Change tracking implemented
- [ ] Audit trail preservation
- [ ] Regular compliance reviews
- [ ] Documentation maintained

Remember: Security is an ongoing process, not a one-time setup. Regularly review and update your security measures to address new threats and vulnerabilities.