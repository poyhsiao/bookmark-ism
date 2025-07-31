## 🔧 修復 GitHub Actions CI/CD 問題

### 問題描述
修復了 GitHub Actions 流水線中的兩個主要問題：
1. **actions/upload-artifact v3 棄用警告**
2. **Package cleanup 中的 'Package not found' 錯誤**

### 🎯 修復內容

#### 1. 更新已棄用的 Actions 版本
- ✅ 將 `actions/upload-artifact@v3` 更新到 `v4`
- ✅ 將 `actions/download-artifact@v3` 更新到 `v4`
- 📁 涉及文件：
  - `.github/workflows/ci.yml` (2 處更新)
  - `.github/workflows/cd.yml` (1 處更新)
  - `.github/workflows/performance-test.yml` (5 處更新)

#### 2. 修復 Package Cleanup 配置
- ✅ 修正 `package-name` 參數格式錯誤
- ✅ 添加 `continue-on-error: true` 防止流水線失敗
- ✅ 簡化錯誤處理邏輯
- ✅ 改進權限設置和執行條件

**修復前：**
```yaml
package-name: ${{ env.IMAGE_NAME }}  # 錯誤：包含完整路徑
```

**修復後：**
```yaml
package-name: ${{ github.event.repository.name }}  # 正確：只使用倉庫名
continue-on-error: true  # 防止清理失敗影響整個流水線
```

### 📋 新增工具和文檔

#### 1. 驗證腳本
- 📄 `scripts/validate-github-actions.sh` - 檢查工作流配置的語法和常見問題

#### 2. 修復文檔
- 📄 `GITHUB_ACTIONS_FIXES.md` - 詳細的修復說明和最佳實踐指南
- 📄 `create-pr.md` - Pull Request 創建指南

### ✅ 預期結果

修復後的 CI/CD 流水線將：
- 🚫 不再出現 v3 版本棄用警告
- 🛡️ Package cleanup 不會導致流水線失敗
- 🧹 成功清理舊的容器鏡像（保留最新 10 個版本）
- 🎯 在包不存在時優雅地跳過清理步驟

### 🧪 測試

可以使用新增的驗證腳本來檢查配置：
```bash
./scripts/validate-github-actions.sh
```

### 📚 相關文檔

- [GitHub Actions - upload-artifact](https://github.com/actions/upload-artifact)
- [GitHub Actions - delete-package-versions](https://github.com/actions/delete-package-versions)
- [GitHub Packages API](https://docs.github.com/en/rest/packages)

---

**類型**: 🐛 Bug Fix
**影響範圍**: CI/CD Pipeline
**測試**: ✅ 已驗證配置語法
**文檔**: ✅ 已更新相關文檔