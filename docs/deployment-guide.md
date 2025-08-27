# 书签同步服务部署指南

## 部署概述

本指南提供书签同步服务的完整部署说明，包括开发环境、测试环境和生产环境的部署步骤。

## 系统要求

### 硬件要求

#### 最小配置
- CPU: 2核心
- 内存: 4GB RAM
- 存储: 20GB SSD
- 网络: 100Mbps

#### 推荐配置
- CPU: 4核心
- 内存: 8GB RAM
- 存储: 100GB SSD
- 网络: 1Gbps

#### 生产环境
- CPU: 8核心
- 内存: 16GB RAM
- 存储: 500GB SSD
- 网络: 10Gbps

### 软件要求

#### 必需软件
- Go 1.19+
- Node.js 18+
- Docker 20.10+
- Docker Compose 2.0+

#### 数据库
- PostgreSQL 14+
- Redis 6.0+

#### 搜索引擎
- Typesense 0.24+

#### 文件存储
- MinIO (S3兼容)

## 环境配置

### 开发环境

#### 1. 克隆代码库
```bash
git clone https://github.com/your-org/bookmark-sync-service.git
cd bookmark-sync-service
```

#### 2. 安装依赖
```bash
# Go 依赖
go mod download

# 前端依赖
cd frontend
npm install
cd ..
```

#### 3. 配置环境变量
```bash
cp .env.example .env.development
# 编辑 .env.development 文件
```

#### 4. 启动开发服务
```bash
# 启动数据库和依赖服务
docker-compose -f docker-compose.dev.yml up -d

# 启动后端服务
make dev-backend

# 启动前端服务
make dev-frontend
```

### 测试环境

#### 1. 构建镜像
```bash
# 构建所有服务镜像
make build-all

# 或分别构建
make build-backend
make build-frontend
make build-extension
```

#### 2. 部署到测试环境
```bash
# 使用Docker Compose部署
docker-compose -f docker-compose.test.yml up -d

# 运行数据库迁移
make db-migrate

# 运行测试套件
make test-all
```

### 生产环境

#### 使用Docker部署

##### 1. 准备配置文件
```bash
# 复制配置模板
cp configs/production.env.example configs/production.env

# 编辑生产配置
nano configs/production.env
```

##### 2. 配置反向代理
```nginx
# /etc/nginx/sites-available/bookmark-sync
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/ssl/cert.pem;
    ssl_certificate_key /path/to/ssl/key.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

##### 3. 启动生产服务
```bash
# 部署生产环境
docker-compose -f docker-compose.prod.yml up -d

# 等待服务启动
docker-compose logs -f
```

#### 使用Kubernetes部署

##### 1. 准备配置文件
```yaml
# configs/k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bookmark-sync
```

##### 2. 配置ConfigMap和Secret
```yaml
# configs/k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: bookmark-sync-config
  namespace: bookmark-sync
data:
  DATABASE_URL: "postgresql://user:pass@postgres:5432/bookmarkdb"
  REDIS_URL: "redis://redis:6379"
  TYPESENSE_URL: "http://typesense:8108"
```

##### 3. 部署应用
```bash
# 应用配置
kubectl apply -f configs/k8s/

# 检查部署状态
kubectl get pods -n bookmark-sync
kubectl get services -n bookmark-sync
```

## 数据库设置

### PostgreSQL 配置

#### 1. 创建数据库
```sql
CREATE DATABASE bookmarkdb;
CREATE USER bookmarkuser WITH PASSWORD 'securepassword';
GRANT ALL PRIVILEGES ON DATABASE bookmarkdb TO bookmarkuser;
```

#### 2. 运行迁移
```bash
# 生成迁移文件
make db-migration name=initial_schema

# 运行迁移
make db-migrate

# 回滚迁移
make db-rollback
```

#### 3. 性能优化
```sql
-- 创建索引
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at);
CREATE INDEX idx_bookmarks_url_hash ON bookmarks(url_hash);

-- 配置参数
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
```

### Redis 配置

#### 1. 配置文件
```conf
# redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

#### 2. 集群配置
```bash
# 启动Redis集群
redis-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
  --cluster-replicas 1
```

## 监控和日志

### 监控配置

#### 1. Prometheus 配置
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'bookmark-sync'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

#### 2. Grafana 仪表板
```bash
# 导入预配置的仪表板
curl -X POST \
  http://admin:admin@localhost:3000/api/dashboards/db \
  -H 'Content-Type: application/json' \
  -d @configs/grafana/dashboard.json
```

### 日志配置

#### 1. 日志聚合
```yaml
# docker-compose.logging.yml
version: '3.8'
services:
  fluentd:
    image: fluent/fluentd:v1.14
    ports:
      - "24224:24224"
    volumes:
      - ./configs/fluentd:/fluentd/etc
    environment:
      FLUENTD_CONF: fluent.conf
```

#### 2. ELK Stack 配置
```yaml
# configs/elasticsearch/elasticsearch.yml
cluster.name: bookmark-sync-logs
network.host: 0.0.0.0
discovery.type: single-node
```

## 安全配置

### SSL/TLS 配置

#### 1. 证书获取
```bash
# 使用 Let's Encrypt
certbot certonly --webroot \
  -w /var/www/html \
  -d yourdomain.com \
  -d www.yourdomain.com
```

#### 2. 证书自动更新
```bash
# 添加到 crontab
0 12 * * * /usr/bin/certbot renew --quiet
```

### 防火墙配置

#### 1. UFW 配置
```bash
# 启用防火墙
ufw enable

# 允许必要端口
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS

# 拒绝其他端口
ufw default deny incoming
ufw default allow outgoing
```

#### 2. Fail2ban 配置
```conf
# /etc/fail2ban/jail.local
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true

[nginx-http-auth]
enabled = true
```

## 备份和恢复

### 数据库备份

#### 1. 自动备份脚本
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/var/backups/bookmark-sync"
DATE=$(date +%Y%m%d_%H%M%S)

# PostgreSQL 备份
pg_dump -h localhost -U bookmarkuser bookmarkdb | \
  gzip > "${BACKUP_DIR}/postgres_${DATE}.sql.gz"

# Redis 备份
cp /var/lib/redis/dump.rdb "${BACKUP_DIR}/redis_${DATE}.rdb"

# 清理旧备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete
```

#### 2. 定时备份
```bash
# 添加到 crontab
0 2 * * * /opt/bookmark-sync/scripts/backup.sh
```

### 恢复过程

#### 1. 数据库恢复
```bash
# PostgreSQL 恢复
gunzip -c backup.sql.gz | psql -h localhost -U bookmarkuser bookmarkdb

# Redis 恢复
systemctl stop redis
cp backup.rdb /var/lib/redis/dump.rdb
systemctl start redis
```

## 性能优化

### 应用优化

#### 1. Go 应用优化
```go
// 连接池配置
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)

// Redis 连接池
redisPool := &redis.Pool{
    MaxIdle:     10,
    MaxActive:   100,
    IdleTimeout: 240 * time.Second,
}
```

#### 2. 缓存策略
```go
// 多级缓存
type CacheManager struct {
    L1 *ristretto.Cache  // 内存缓存
    L2 *redis.Client     // Redis缓存
    L3 Database          // 数据库
}
```

### 基础设施优化

#### 1. 负载均衡
```nginx
upstream backend {
    least_conn;
    server backend1:8080 weight=3;
    server backend2:8080 weight=2;
    server backend3:8080 weight=1;
}
```

#### 2. CDN 配置
```javascript
// CloudFlare 配置
const cdnConfig = {
    zones: ['static.yourdomain.com'],
    caching: {
        browser_cache_ttl: 86400,
        edge_cache_ttl: 604800
    }
};
```

## 故障排除

### 常见问题

#### 1. 服务无法启动
```bash
# 检查日志
docker-compose logs backend
kubectl logs -n bookmark-sync deployment/backend

# 检查配置
docker-compose config
kubectl describe pod -n bookmark-sync
```

#### 2. 数据库连接问题
```bash
# 测试连接
psql -h localhost -U bookmarkuser -d bookmarkdb

# 检查网络
nc -zv postgres-host 5432
```

#### 3. 性能问题
```bash
# 监控资源使用
top
htop
iotop
netstat -tulpn
```

### 日志分析

#### 1. 应用日志
```bash
# 实时查看日志
tail -f /var/log/bookmark-sync/app.log

# 搜索错误
grep -i error /var/log/bookmark-sync/app.log

# 分析访问模式
awk '{print $7}' /var/log/nginx/access.log | sort | uniq -c | sort -nr
```

#### 2. 系统日志
```bash
# 查看系统日志
journalctl -u bookmark-sync-backend
journalctl -u bookmark-sync-frontend

# 查看资源使用历史
sar -u 1 10
sar -r 1 10
```

## 维护计划

### 定期维护

#### 1. 每日检查
- 监控服务状态
- 检查错误日志
- 验证备份完成
- 监控资源使用

#### 2. 每周维护
- 更新系统补丁
- 清理日志文件
- 检查安全告警
- 性能分析报告

#### 3. 每月维护
- 数据库优化
- 安全审计
- 容量规划
- 灾难恢复测试

### 版本更新

#### 1. 准备工作
```bash
# 备份当前版本
make backup-all

# 测试新版本
make test-deployment
```

#### 2. 滚动更新
```bash
# Kubernetes 滚动更新
kubectl set image deployment/backend backend=new-image:tag -n bookmark-sync

# Docker Compose 更新
docker-compose pull
docker-compose up -d --no-deps backend
```

#### 3. 回滚计划
```bash
# Kubernetes 回滚
kubectl rollout undo deployment/backend -n bookmark-sync

# Docker 回滚
docker-compose down
docker tag old-image:tag current-image:tag
docker-compose up -d
```