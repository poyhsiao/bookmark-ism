# GitHub Actions CI/CD Configuration

This directory contains the complete CI/CD pipeline configuration for the Bookmark Sync Service project.

## 🚀 Workflows Overview

### 1. CI Pipeline (`.github/workflows/ci.yml`)
**Trigger**: Push to main/develop/task* branches, Pull Requests
**Purpose**: Continuous Integration with comprehensive testing

**Jobs**:
- **Backend Testing**: Go tests with PostgreSQL, Redis, MinIO, Typesense services
- **Frontend Testing**: Browser extension tests and linting
- **Security Scanning**: Gosec and Trivy vulnerability scanning
- **Code Quality**: golangci-lint and formatting checks
- **Docker Build**: Container image build testing
- **Integration Testing**: End-to-end API testing

**Features**:
- ✅ Multi-service testing environment
- ✅ Coverage reporting with 70% threshold
- ✅ Security vulnerability scanning
- ✅ Code quality enforcement
- ✅ Docker build validation
- ✅ Artifact uploads for debugging

### 2. CD Pipeline (`.github/workflows/cd.yml`)
**Trigger**: Push to main branch, version tags, manual dispatch
**Purpose**: Continuous Deployment to staging and production

**Jobs**:
- **Build & Push**: Docker images to GitHub Container Registry
- **Deploy Staging**: Automatic deployment to staging environment
- **Deploy Production**: Deployment to production (tags only)
- **Rollback**: Automatic rollback on deployment failure
- **Cleanup**: Old image cleanup

**Features**:
- ✅ Multi-architecture Docker builds (amd64, arm64)
- ✅ Kubernetes deployment with health checks
- ✅ Automatic rollback on failure
- ✅ GitHub Releases with detailed changelogs
- ✅ SBOM generation for security compliance

### 3. Release Pipeline (`.github/workflows/release.yml`)
**Trigger**: Version tags (v*), manual dispatch
**Purpose**: Create comprehensive releases with assets

**Jobs**:
- **Create Release**: Generate release notes and GitHub release
- **Build Assets**: Cross-platform binaries (Linux, macOS, Windows)
- **Build Extensions**: Browser extension packages
- **Docker Images**: Tagged container images
- **Deployment Package**: Complete deployment bundle

**Features**:
- ✅ Cross-platform binary builds
- ✅ Browser extension packaging
- ✅ Docker image publishing
- ✅ Deployment package creation
- ✅ Automated release notes generation

### 4. Dependency Updates (`.github/workflows/dependency-update.yml`)
**Trigger**: Weekly schedule (Mondays), manual dispatch
**Purpose**: Automated dependency management

**Jobs**:
- **Go Dependencies**: Update Go modules and security scan
- **Node.js Dependencies**: Update npm packages for extensions
- **Docker Images**: Update base images in Dockerfiles
- **Security Scan**: Vulnerability scanning with issue creation

**Features**:
- ✅ Automated dependency updates
- ✅ Security vulnerability detection
- ✅ Automated PR creation
- ✅ Test validation before merge

### 5. Performance Testing (`.github/workflows/performance-test.yml`)
**Trigger**: Weekly schedule (Sundays), manual dispatch
**Purpose**: Performance monitoring and regression detection

**Jobs**:
- **Load Testing**: k6-based load testing with virtual users
- **Database Performance**: PostgreSQL performance benchmarking
- **Resource Usage**: CPU and memory monitoring
- **Performance Summary**: Comprehensive reporting

**Features**:
- ✅ Configurable load testing parameters
- ✅ Database performance benchmarking
- ✅ Resource usage monitoring
- ✅ Performance regression detection
- ✅ Automated issue creation on failures

## 🔧 Configuration Files

### Issue Templates
- **Bug Report** (`.github/ISSUE_TEMPLATE/bug_report.yml`): Structured bug reporting
- **Feature Request** (`.github/ISSUE_TEMPLATE/feature_request.yml`): Feature suggestion template

### Pull Request Template
- **PR Template** (`.github/pull_request_template.md`): Comprehensive PR checklist

### Code Ownership
- **CODEOWNERS** (`.github/CODEOWNERS`): Code review assignments

### Dependency Management
- **Dependabot** (`.github/dependabot.yml`): Automated dependency updates

## 🏗️ Infrastructure

### Kubernetes Deployments
- **Staging** (`k8s/staging/`): Staging environment configuration
- **Production** (`k8s/production/`): Production environment configuration

**Features**:
- ✅ Namespace isolation
- ✅ Resource limits and requests
- ✅ Health checks and probes
- ✅ SSL/TLS termination
- ✅ Rate limiting
- ✅ Rolling updates

### Docker Configuration
- **Backend Dockerfile** (`backend/Dockerfile`): Multi-stage Go application build
- **Security**: Non-root user, minimal base image
- **Health Checks**: Built-in health monitoring

## 🔒 Security Features

### Vulnerability Scanning
- **Gosec**: Go security analyzer
- **Trivy**: Container and filesystem vulnerability scanner
- **SARIF Upload**: Security findings integration with GitHub Security tab

### Code Quality
- **golangci-lint**: Comprehensive Go linting
- **Coverage Threshold**: Minimum 70% test coverage
- **Security Headers**: Proper security configuration

## 📊 Monitoring & Observability

### Health Checks
- Application health endpoints
- Database connectivity checks
- Redis connectivity verification
- Service dependency monitoring

### Performance Monitoring
- Response time tracking
- Error rate monitoring
- Resource usage analysis
- Load testing automation

## 🚀 Getting Started

### Prerequisites
1. **GitHub Secrets**: Configure required secrets in repository settings
2. **Container Registry**: GitHub Container Registry access
3. **Kubernetes Cluster**: EKS or similar for deployments
4. **Domain**: Configure DNS for staging/production

### Required Secrets
```bash
# AWS/Kubernetes
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_REGION
EKS_CLUSTER_NAME

# Database
DATABASE_HOST
DATABASE_USER
DATABASE_PASSWORD
DATABASE_NAME

# Redis
REDIS_HOST
REDIS_PASSWORD

# JWT
JWT_SECRET

# Container Registry
GITHUB_TOKEN (automatically provided)
```

### Deployment Process
1. **Development**: Push to feature branches triggers CI
2. **Staging**: Merge to main triggers staging deployment
3. **Production**: Create version tag triggers production deployment
4. **Monitoring**: Automated performance and security monitoring

## 📈 Metrics & Reporting

### CI/CD Metrics
- Build success rate
- Test coverage trends
- Deployment frequency
- Lead time for changes

### Performance Metrics
- Response time percentiles
- Error rates
- Resource utilization
- Load testing results

### Security Metrics
- Vulnerability scan results
- Dependency update frequency
- Security issue resolution time

## 🔄 Maintenance

### Weekly Tasks (Automated)
- Dependency updates
- Security vulnerability scanning
- Performance testing
- Docker image updates

### Monthly Tasks (Manual)
- Review CI/CD pipeline performance
- Update deployment configurations
- Security audit and review
- Performance optimization

## 📚 Documentation

### Additional Resources
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Security Best Practices](https://docs.github.com/en/code-security)

### Support
For questions or issues with the CI/CD pipeline:
1. Check workflow logs in GitHub Actions tab
2. Review this documentation
3. Create an issue using the bug report template
4. Contact the development team

---

**Note**: This CI/CD configuration is designed for a production-ready application with comprehensive testing, security scanning, and deployment automation. Adjust configurations based on your specific requirements and infrastructure setup.