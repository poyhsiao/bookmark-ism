# Docker 構建路徑修復

## 🚨 問題描述

在 GitHub Actions 的 "build and push images" 步驟中出現構建失敗：

```
buildx failed with: ERROR: failed to build: failed to solve: process "/bin/sh -c CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags \"-static\"' -a -installsuffix cgo -o main ./cmd/api" did not complete successfully: exit code: 1
```

## 🔍 根本原因

儘管之前已經修復了 Dockerfile 的部分路徑問題，但構建命令中的目標路徑仍然不正確：

### 問題分析：
1. **錯誤的構建路徑**: Dockerfile 中使用 `./cmd/api`，但實際路徑應該是 `./backend/cmd/api`
2. **項目結構不匹配**: Docker 構建上下文是根目錄，但構建路徑假設在 backend 目錄內
3. **配置文件複製錯誤**: 嘗試複製不存在的 config 目錄

### 項目結構：
```
bookmark-sync-service/
├── go.mod                    # Go 模塊根文件
├── backend/
│   ├── cmd/api/main.go      # 實際的 API 入口點
│   ├── internal/            # 內部包
│   ├── pkg/                 # 公共包
│   └── Dockerfile           # Docker 構建文件
└── ...
```

## 🔧 修復方案

### 修復前的 Dockerfile：
```dockerfile
# 錯誤的構建路徑
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/api  # ❌ 路徑錯誤

# 錯誤的配置複製
COPY --from=builder /app/config ./config  # ❌ config 目錄不存在
```

### 修復後的 Dockerfile：
```dockerfile
# 正確的構建路徑
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./backend/cmd/api  # ✅ 正確路徑

# 註釋掉不存在的配置複製
# COPY --from=builder /app/config ./config  # ✅ 已註釋
```

## 📋 詳細修復內容

### 1. 修正構建目標路徑
```dockerfile
# 修復前
-o main ./cmd/api

# 修復後
-o main ./backend/cmd/api
```

### 2. 移除不存在的配置複製
```dockerfile
# 修復前
COPY --from=builder /app/config ./config

# 修復後
# COPY --from=builder /app/config ./config  # 註釋掉
```

## ✅ 驗證結果

### 本地測試：
```bash
# 構建測試
docker build -f backend/Dockerfile -t test-build-fix .

# 結果：✅ 構建成功
[+] Building 27.2s (19/19) FINISHED

# 二進制測試
docker run --rm test-build-fix --help
# 結果：✅ 二進制文件正常
```

### GitHub Actions 配置：
CD workflow 中的配置保持不變：
```yaml
- name: Build and push backend image
  uses: docker/build-push-action@v5
  with:
    context: .                    # 根目錄作為構建上下文
    file: ./backend/Dockerfile    # Dockerfile 位置
    push: true
    tags: ${{ steps.meta.outputs.tags }}
```

## 🚀 預期結果

修復後，GitHub Actions 的 Docker 構建應該：
- ✅ 成功找到正確的 Go 入口點文件
- ✅ 完成 Docker 鏡像構建
- ✅ 推送鏡像到容器註冊表
- ✅ 不再出現 "exit code: 1" 錯誤

## 📚 相關文件

- `backend/Dockerfile` - 修復的 Docker 構建文件
- `.github/workflows/cd.yml` - CI/CD 配置（無需修改）
- `backend/cmd/api/main.go` - Go 應用程序入口點

## 🔄 部署流程

1. **提交修復**: 將修復提交到新分支
2. **創建 PR**: 創建 Pull Request 進行審查
3. **測試驗證**: GitHub Actions 自動測試構建
4. **合併部署**: 合併後自動部署

這個修復解決了 Docker 構建路徑不匹配的問題，確保 CI/CD 流水線能夠成功構建和部署應用程序。