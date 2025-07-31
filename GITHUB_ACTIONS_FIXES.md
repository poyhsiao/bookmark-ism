# GitHub Actions 修復總結

## 問題描述

在 GitHub CI/CD 流水線中遇到以下錯誤：

1. **actions/upload-artifact v3 棄用警告**
   ```
   Error: This request has been automatically failed because it uses a deprecated version of actions/upload-artifact: v3
   ```

2. **Package cleanup 失敗**
   ```
   Error: get versions API failed. Package not found.
   ```

## 修復內容

### 1. 更新 Actions 版本

將所有已棄用的 actions 更新到最新版本：

**修復的文件：**
- `.github/workflows/ci.yml`
- `.github/workflows/cd.yml`
- `.github/workflows/performance-test.yml`

**具體更改：**
```yaml
# 修復前
- uses: actions/upload-artifact@v3
- uses: actions/download-artifact@v3

# 修復後
- uses: actions/upload-artifact@v4
- uses: actions/download-artifact@v4
```

### 2. 修復 Package Cleanup

**問題原因：**
- 使用了錯誤的 package name 格式
- 缺少適當的錯誤處理
- 在包不存在時仍嘗試清理

**修復方案：**

```yaml
# 修復前
- name: Delete old container images
  uses: actions/delete-package-versions@v4
  with:
    package-name: ${{ env.IMAGE_NAME }}  # 錯誤：包含完整路徑
    package-type: 'container'
    min-versions-to-keep: 10
    delete-only-untagged-versions: true

# 修復後
- name: Delete old container images
  uses: actions/delete-package-versions@v4
  with:
    package-name: ${{ github.event.repository.name }}  # 正確：只使用倉庫名
    package-type: 'container'
    min-versions-to-keep: 10
    delete-only-untagged-versions: true
    token: ${{ secrets.GITHUB_TOKEN }}
  continue-on-error: true  # 添加錯誤容忍
```

### 3. 改進的 Cleanup 策略

**新的 cleanup 配置特點：**
- ✅ 簡化的錯誤處理
- ✅ 只在成功構建後執行清理
- ✅ 使用正確的包名格式
- ✅ 添加 `continue-on-error: true` 防止流水線失敗
- ✅ 明確的權限設置

```yaml
cleanup:
  name: Cleanup Old Images
  runs-on: ubuntu-latest
  needs: [build-and-push]
  if: success()
  permissions:
    packages: write

  steps:
  - name: Delete old container images
    uses: actions/delete-package-versions@v4
    with:
      package-name: ${{ github.event.repository.name }}
      package-type: 'container'
      min-versions-to-keep: 10
      delete-only-untagged-versions: true
      token: ${{ secrets.GITHUB_TOKEN }}
    continue-on-error: true
```

## 驗證工具

創建了驗證腳本 `scripts/validate-github-actions.sh` 來檢查：
- YAML 語法正確性
- 已棄用的 actions 版本
- Package cleanup 配置問題

## 使用方法

```bash
# 運行驗證腳本
./scripts/validate-github-actions.sh

# 手動檢查特定工作流
gh workflow list
gh workflow view ci.yml
```

## 預期結果

修復後，GitHub Actions 流水線應該：

1. ✅ 不再出現 v3 棄用警告
2. ✅ Package cleanup 不會導致流水線失敗
3. ✅ 成功清理舊的容器鏡像（如果存在）
4. ✅ 在包不存在時優雅地跳過清理

## 最佳實踐

1. **定期更新 Actions 版本**
   - 訂閱 GitHub Actions 更新通知
   - 使用 Dependabot 自動更新

2. **錯誤處理**
   - 對非關鍵步驟使用 `continue-on-error: true`
   - 添加適當的條件檢查

3. **權限管理**
   - 明確指定所需的權限
   - 使用最小權限原則

4. **監控和維護**
   - 定期檢查工作流運行狀態
   - 及時修復警告和錯誤

## 相關文檔

- [GitHub Actions - upload-artifact](https://github.com/actions/upload-artifact)
- [GitHub Actions - delete-package-versions](https://github.com/actions/delete-package-versions)
- [GitHub Packages API](https://docs.github.com/en/rest/packages)