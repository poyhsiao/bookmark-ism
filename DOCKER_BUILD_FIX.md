# Docker 構建修復

## 問題描述

在 GitHub Actions CI/CD 的 "build and push image" 步驟中出現構建失敗：

```
ERROR: failed to build: failed to solve: process "/bin/sh -c CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags \"-static\"' -a -installsuffix cgo -o main ./cmd/api" did not complete successfully: exit code: 1
```

## 根本原因

問題出現在 `backend/Dockerfile` 中的路徑配置錯誤：

1. **構建上下文問題**: GitHub Actions 使用根目錄 (`.`) 作為 Docker 構建上下文
2. **Go 模塊結構**: 項目的 Go 模塊根目錄在項目根目錄，不是 `backend/` 目錄
3. **路徑不匹配**: Dockerfile 中的 COPY 和 build 路徑與實際項目結構不符

## 修復方案

### 修復前的 Dockerfile:
```dockerfile
# 錯誤的路徑配置
COPY backend/go.mod backend/go.sum ./  # 路徑錯誤
COPY backend/ .                        # 只複製 backend 目錄
RUN go build -o main ./cmd/api         # 路徑不正確
```

### 修復後的 Dockerfile:
```dockerfile
# 正確的路徑配置
COPY go.mod go.sum ./                  # 從根目錄複製 go.mod
COPY . .                               # 複製整個項目
RUN go build -o main ./backend/cmd/api # 正確的構建路徑
```

## 詳細修復內容

### 1. 修正 Go 模塊文件複製
```dockerfile
# 修復前
COPY backend/go.mod backend/go.sum ./

# 修復後
COPY go.mod go.sum ./
```

### 2. 修正源代碼複製
```dockerfile
# 修復前
COPY backend/ .

# 修復後
COPY . .
```

### 3. 修正構建路徑
```dockerfile
# 修復前
RUN go build -o main ./cmd/api

# 修復後
RUN go build -o main ./backend/cmd/api
```

## 驗證結果

修復後的 Docker 構建測試：

```bash
# 本地測試構建
docker build -f backend/Dockerfile -t test-build .

# 結果：✅ 構建成功
[+] Building 31.0s (19/19) FINISHED
```

## 項目結構說明

```
bookmark-sync-service/
├── go.mod                    # Go 模塊根文件
├── go.sum                    # Go 依賴鎖定文件
├── backend/
│   ├── cmd/api/main.go      # API 服務入口點
│   ├── internal/            # 內部包
│   ├── pkg/                 # 公共包
│   └── Dockerfile           # Docker 構建文件
├── extensions/              # 瀏覽器擴展
├── web/                     # Web 前端
└── ...
```

## GitHub Actions 配置

CD workflow 中的 Docker 構建配置保持不變：

```yaml
- name: Build and push backend image
  uses: docker/build-push-action@v5
  with:
    context: .                    # 根目錄作為構建上下文
    file: ./backend/Dockerfile    # Dockerfile 位置
    push: true
    tags: ${{ steps.meta.outputs.tags }}
```

## 最佳實踐

1. **統一模塊管理**: 在項目根目錄維護單一的 go.mod 文件
2. **正確的構建上下文**: 確保 Dockerfile 路徑與構建上下文匹配
3. **路徑一致性**: 保持 Dockerfile 中的路徑與實際項目結構一致
4. **本地測試**: 在提交前本地測試 Docker 構建

## 相關文件

- `backend/Dockerfile` - 修復的 Docker 構建文件
- `.github/workflows/cd.yml` - CI/CD 配置（無需修改）
- `go.mod` - Go 模塊配置文件

修復完成後，GitHub Actions 的 Docker 構建應該能夠成功完成。