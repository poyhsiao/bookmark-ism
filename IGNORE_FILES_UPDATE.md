# .gitignore 和 .dockerignore 更新說明

## 📋 更新概述

根據當前專案架構，更新了 `.gitignore` 和 `.dockerignore` 文件，以確保不應該提交到 Git 倉庫或包含在 Docker 構建上下文中的文件被正確排除。

## 🔧 .gitignore 更新

### 新增的忽略規則

#### Go 構建產物
```gitignore
# Go build artifacts and binaries
/api
/main
backend/api
backend/main
backend/coverage.out
backend/coverage.html
backend/*.test
backend/*.prof
```

#### 瀏覽器擴展構建產物
```gitignore
# Extension build artifacts
extensions/*/dist/
extensions/*/build/
extensions/*/web-ext-artifacts/
extensions/*.zip
extensions/*.crx
extensions/*.xpi
extensions/*/node_modules/
extensions/*/package-lock.json
```

#### Web 前端構建產物
```gitignore
# Web frontend build artifacts
web/dist/
web/build/
web/node_modules/
web/.next/
web/.nuxt/
```

#### 本地配置和數據目錄
```gitignore
# Local configuration overrides
config/local/
config/dev/
config/*.local.*

# Runtime data directories
data/
storage/
uploads/
screenshots/
avatars/
backups/
```

#### 開發和調試文件
```gitignore
# Test and development files
*_test.go.bak
*.test.bak
test-*
debug-*

# Performance and debugging
*.pprof
*.trace
*.heap
*.cpu

# Temporary fix files (from development)
*_FIX.md.bak
*_TEMP.md
TEMP_*
DEBUG_*
```

#### Kubernetes 和 Docker 本地配置
```gitignore
# Kubernetes local configs
k8s/local/
k8s/*.local.yaml
k8s/secrets/

# Docker development files
docker-compose.override.yml
docker-compose.dev.yml
```

#### Supabase 本地開發文件
```gitignore
# Supabase local development
supabase/.branches/
supabase/.temp/
supabase/logs/
supabase/docker/
```

## 🐳 .dockerignore 更新

### 優化的 Docker 構建排除規則

#### 開發相關文件（不需要在容器中）
```dockerignore
# Frontend source (not needed for backend container)
web/
extensions/

# Development scripts (not needed in container)
scripts/
Makefile

# CI/CD configs (not needed in container)
.github/

# Kiro IDE files (not needed in container)
.kiro/
.roo/
```

#### 測試和調試文件
```dockerignore
# Test files and coverage (not needed in production container)
*_test.go
*.test
coverage.out
coverage.html
*.prof
*.pprof
*.trace
```

#### 構建產物（將在容器內構建）
```dockerignore
# Build artifacts (will be built inside container)
bin/
dist/
build/
/api
/main
backend/api
backend/main
```

#### 外部服務配置
```dockerignore
# Supabase (external service)
supabase/

# Kubernetes configs (not needed in container)
k8s/

# Local development configs
nginx/
config/local/
config/dev/
```

## 🎯 更新的好處

### Git 倉庫優化
1. **減少倉庫大小**: 排除構建產物和臨時文件
2. **避免衝突**: 排除本地配置和 IDE 文件
3. **提高安全性**: 排除可能包含敏感信息的文件
4. **清潔的提交歷史**: 只包含源代碼和必要的配置

### Docker 構建優化
1. **更快的構建速度**: 減少構建上下文大小
2. **更小的鏡像**: 只包含運行時需要的文件
3. **更好的緩存**: 避免不必要的文件變更影響緩存
4. **安全性**: 排除開發工具和敏感配置

## 📁 項目結構對應

### 保留在 Git 中的重要文件
- ✅ 源代碼 (`backend/`, `extensions/`, `web/`)
- ✅ 配置模板 (`.env.example`, `config/config.yaml`)
- ✅ 部署配置 (`k8s/`, `docker-compose.yml`)
- ✅ 文檔 (`README.md`, `docs/`)
- ✅ CI/CD 配置 (`.github/`)

### 排除的文件類型
- ❌ 構建產物 (`bin/`, `dist/`, `*.test`)
- ❌ 依賴目錄 (`node_modules/`, `vendor/`)
- ❌ 臨時文件 (`*.tmp`, `*.log`, `coverage.out`)
- ❌ IDE 配置 (`.vscode/`, `.idea/`)
- ❌ 本地覆蓋 (`*.local.*`, `docker-compose.override.yml`)

## 🔄 維護建議

1. **定期檢查**: 隨著項目發展，定期檢查和更新忽略規則
2. **團隊同步**: 確保團隊成員了解新的忽略規則
3. **測試驗證**: 在添加新的構建產物或工具時，及時更新忽略規則
4. **文檔更新**: 保持文檔與實際配置同步

這些更新確保了項目的清潔性、安全性和構建效率。