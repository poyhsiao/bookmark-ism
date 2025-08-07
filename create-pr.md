# 創建 Pull Request 指南

## 🎉 代碼已成功推送！

你的修復已經成功推送到 `task16` 分支。現在需要創建 Pull Request。

## 方法 1: 使用 GitHub 網頁界面（推薦）

1. 打開瀏覽器，訪問：
   ```
   https://github.com/poyhsiao/bookmark-ism/compare/main...task16
   ```

2. 點擊 "Create pull request" 按鈕

3. 填寫 PR 信息：

### 標題：
```
fix: Update GitHub Actions to resolve deprecated warnings and package cleanup errors
```

### 描述：
```markdown
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
```

## 方法 2: 使用 GitHub CLI（需要先登錄）

如果你想使用命令行，需要先完成 GitHub CLI 登錄：

```bash
# 登錄 GitHub CLI
gh auth login

# 創建 PR
gh pr create --title "fix: Update GitHub Actions to resolve deprecated warnings and package cleanup errors" --body-file pr-description.md --base main
```

## 📋 修復摘要

### 修改的文件：
- ✅ `.github/workflows/ci.yml` - 更新 artifact actions 版本
- ✅ `.github/workflows/cd.yml` - 修復 package cleanup 配置
- ✅ `.github/workflows/performance-test.yml` - 更新 artifact actions 版本
- ✅ `GITHUB_ACTIONS_FIXES.md` - 新增修復文檔
- ✅ `scripts/validate-github-actions.sh` - 新增驗證腳本

### Git 提交信息：
```
fix: update GitHub Actions to resolve deprecated warnings and package cleanup errors

- Update actions/upload-artifact from v3 to v4 across all workflows
- Update actions/download-artifact from v3 to v4
- Fix package cleanup configuration in CD pipeline
  - Use correct package name format (repository name only)
  - Add proper error handling with continue-on-error
  - Simplify cleanup logic and improve reliability
- Add validation script for GitHub Actions workflows
- Add comprehensive documentation of fixes and best practices
```

## 🚀 下一步

1. 使用上述任一方法創建 Pull Request
2. 等待 CI/CD 流水線運行驗證修復
3. 如果一切正常，合併 PR 到 main 分支

修復完成後，你的 GitHub Actions 應該不會再出現棄用警告和 package cleanup 錯誤了！