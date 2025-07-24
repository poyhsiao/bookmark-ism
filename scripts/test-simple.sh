#!/bin/bash

# 簡單測試腳本
# Simple test script

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 檢查是否在專案根目錄
if [ ! -f "go.mod" ]; then
    log_error "請在專案根目錄執行此腳本"
    exit 1
fi

log_info "開始運行測試..."

# 進入 backend 目錄
cd backend

# 下載依賴
log_info "下載依賴..."
go mod download
go mod tidy

# 創建測試結果目錄
mkdir -p ../test-results

# 運行基本測試
log_info "運行資料庫模型測試..."
go test -v ./pkg/database/... || {
    log_error "資料庫模型測試失敗"
    exit 1
}

log_info "運行配置測試..."
go test -v ./internal/config/... || {
    log_error "配置測試失敗"
    exit 1
}

log_info "運行中間件測試..."
go test -v ./pkg/middleware/... || {
    log_error "中間件測試失敗"
    exit 1
}

# 運行所有測試並生成覆蓋率
log_info "運行所有測試並生成覆蓋率報告..."
go test -v -coverprofile=../test-results/coverage.out ./... || {
    log_error "測試執行失敗"
    exit 1
}

# 生成 HTML 覆蓋率報告
go tool cover -html=../test-results/coverage.out -o ../test-results/coverage.html

# 顯示覆蓋率統計
log_info "覆蓋率統計："
go tool cover -func=../test-results/coverage.out | tail -1

cd ..

log_success "測試完成！查看覆蓋率報告：open test-results/coverage.html"