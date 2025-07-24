# Bookmark Sync Service Backend

這是書籤同步服務的後端部分，使用 Go 語言和 Gin 框架開發。

## 功能

- 用戶認證和授權
- 用戶個人資料管理
- 書籤管理
- 收藏夾管理
- 跨瀏覽器同步
- 搜索和發現
- 社區功能

## 技術堆棧

- **語言**: Go 與 Gin Web 框架
- **數據庫**: 自託管 Supabase PostgreSQL 與 GORM ORM
- **緩存**: Redis 與 Pub/Sub 用於實時同步
- **搜索**: Typesense 支持中文語言
- **存儲**: MinIO (所有文件的主要存儲)
- **認證**: 自託管 Supabase Auth 與 JWT
- **實時**: 自託管 Supabase Realtime + WebSocket 與 Gorilla WebSocket 庫

## 目錄結構

```
backend/
├── cmd/                   # 應用程序入口點
│   ├── api/              # API 服務器
│   ├── sync/             # 同步服務
│   ├── worker/           # 後台工作者
│   └── migrate/          # 數據庫遷移
├── internal/             # 私有應用程序代碼
│   ├── auth/             # 認證邏輯
│   ├── bookmark/         # 書籤業務邏輯
│   ├── sync/             # 同步邏輯
│   ├── community/        # 社交功能
│   ├── search/           # 搜索集成
│   └── storage/          # 文件存儲邏輯
├── pkg/                  # 公共包
│   ├── database/         # 數據庫模型和連接
│   ├── redis/            # Redis 客戶端操作
│   ├── websocket/        # WebSocket 管理
│   └── utils/            # 共享工具
├── api/                  # API 定義
│   └── v1/               # API v1 路由
├── config/               # 配置文件
├── migrations/           # 數據庫模式遷移
└── docker/               # Docker 配置
```

## 開發設置

### 前提條件

- Go 1.21+
- Docker 和 Docker Compose
- Make

### 安裝

1. 克隆存儲庫：

```bash
git clone https://github.com/yourusername/bookmark-sync-service.git
cd bookmark-sync-service
```

2. 啟動開發環境：

```bash
make dev
```

這將啟動所有必要的服務（Supabase PostgreSQL、Redis、MinIO、Typesense）。

3. 運行後端：

```bash
make run
```

### 測試

運行所有測試：

```bash
make test
```

運行特定包的測試：

```bash
go test -v ./backend/internal/auth/...
```

生成測試覆蓋率報告：

```bash
make coverage
```

### API 文檔

API 文檔可在 `http://localhost:8080/swagger/index.html` 獲取（當服務器運行時）。

## 部署

### 使用 Docker Compose

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 擴展 API 服務

```bash
docker-compose up --scale api=3
```

## 配置

配置通過環境變量或配置文件提供。查看 `.env.example` 文件了解可用的配置選項。

## 貢獻

1. Fork 存儲庫
2. 創建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打開拉取請求

## 許可證

[MIT](LICENSE)