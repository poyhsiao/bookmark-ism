# Copilot Instructions for Bookmark Sync Service

這是一個多用戶的書籤同步服務，提供跨瀏覽器的書籤管理功能，具有類似 Toby 的視覺介面。以下是開發此專案時需要注意的重要事項：

## 系統需求

請參考 `.kiro/specs/bookmark-sync-service/requirements.md` 檔案以獲取詳細的需求和說明。

## 設計需求

請參考 `.kiro/specs/bookmark-sync-service/design.md` 檔案以獲取詳細的設計說明和規格。

## 工作項目和任務

請參考 `.kiro/specs/bookmark-sync-service/tasks.md` 檔案以獲取詳細的工作項目和任務說明。

## 系統架構

- **多組件服務架構**：
  - API 伺服器：`backend/cmd/api/main.go`
  - 資料同步服務：`backend/cmd/sync/main.go`
  - 背景工作服務：`backend/cmd/worker/main.go`
  - 資料庫遷移工具：`backend/cmd/migrate/main.go`

- **關鍵依賴服務**：
  - Supabase (PostgreSQL + Auth)
  - Redis (快取 + Pub/Sub)
  - Typesense (搜尋引擎)
  - MinIO (檔案儲存)

## 開發工作流程

### 建置與運行

```bash
# 完整開發環境設置
make setup

# 開發模式運行（支援熱重載）
make dev

# 檢查服務健康狀態
make health
```

### 資料庫操作

- 資料模型定義在 `backend/pkg/database/models.go`
- 使用 GORM 作為 ORM，支援軟刪除
- 所有模型都繼承自 `BaseModel`，包含標準時間戳記欄位

## 專案特定模式

### 1. 實時同步機制

- 使用 WebSocket (`backend/pkg/websocket`) 處理瀏覽器即時連接
- Redis Pub/Sub 用於跨服務實例的事件廣播
- 參考 `backend/internal/server/server.go` 中的 WebSocket 集成

### 2. 認證流程

- 使用 self-hosted Supabase Auth 進行身份驗證
- JWT 令牌驗證在 `backend/pkg/middleware/auth.go` 中實現
- 用戶模型與 Supabase ID 關聯

### 3. 資料存儲模式

- 使用 MinIO 作為主要檔案存儲
- 書籤相關檔案（favicon、截圖）自動同步到存儲系統
- 元數據使用 JSONB 格式儲存在 PostgreSQL 中

### 4. 錯誤處理約定

- 使用 `backend/pkg/utils/response.go` 中定義的標準回應格式
- 所有 HTTP 回應應該遵循一致的錯誤結構
- 使用 zap logger 進行結構化日誌記錄

## 關鍵檔案參考

- `config/config.yaml`：配置範例和說明
- `backend/internal/config/config.go`：配置結構定義
- `backend/pkg/database/models.go`：資料模型定義
- `backend/internal/server/server.go`：主要服務器邏輯

## 測試與除錯

- 使用 `make test` 運行測試套件
- 開發時使用 `make dev` 啟用熱重載
- 使用 `make docker-logs` 查看容器日誌
- 健康檢查端點可通過 `make health` 訪問
