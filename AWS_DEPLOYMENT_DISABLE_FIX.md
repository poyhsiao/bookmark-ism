# AWS 部署錯誤修復

## 🚨 問題描述

GitHub Actions CD 流水線中出現多個 AWS 相關錯誤：

1. **Deploy to Staging 錯誤**:
   ```
   Error: Input required and not supplied: aws-region
   ```

2. **Rollback Deployment 錯誤**:
   ```
   Error: Input required and not supplied: aws-region
   ```

3. **Notify Deployment Status 錯誤**:
   ```
   ❌ Deployment failed. Check the logs for details.
   Error: Process completed with exit code 1.
   ```

## 🔍 問題分析

### 根本原因
CD workflow 包含了完整的 AWS EKS 部署流程，但缺少必要的 AWS 配置：
- 缺少 `AWS_ACCESS_KEY_ID` secret
- 缺少 `AWS_SECRET_ACCESS_KEY` secret
- 缺少 `AWS_REGION` secret
- 缺少 `EKS_CLUSTER_NAME` secret

### 當前狀況
你目前不需要部署到 AWS，但 workflow 仍然嘗試執行部署步驟，導致失敗。

## 🔧 修復方案

### 1. 添加部署開關控制
使用 GitHub repository variable 來控制是否啟用 AWS 部署：

```yaml
# 修復前 - 總是嘗試部署
if: github.ref == 'refs/heads/main'

# 修復後 - 可選部署
if: github.ref == 'refs/heads/main' && vars.ENABLE_AWS_DEPLOYMENT == 'true'
```

### 2. 添加跳過部署的替代步驟
```yaml
# 新增：當不部署時的通知步驟
skip-deployment:
  name: Skip AWS Deployment
  runs-on: ubuntu-latest
  needs: build-and-push
  if: vars.ENABLE_AWS_DEPLOYMENT != 'true'

  steps:
  - name: Skip deployment notification
    run: |
      echo "🚀 Docker image built and pushed successfully!"
      echo "ℹ️ AWS deployment is disabled."
```

### 3. 修復回滾步驟條件
```yaml
# 修復前 - 總是嘗試回滾
if: failure() && (needs.deploy-staging.result == 'failure' || needs.deploy-production.result == 'failure')

# 修復後 - 只在啟用部署時回滾
if: failure() && vars.ENABLE_AWS_DEPLOYMENT == 'true' && (needs.deploy-staging.result == 'failure' || needs.deploy-production.result == 'failure')
```

### 4. 改進通知邏輯
```yaml
# 修復前 - 假設總是有部署
needs: [deploy-staging, deploy-production]

# 修復後 - 包含所有可能的步驟
needs: [build-and-push, deploy-staging, deploy-production, skip-deployment]
```

## 📋 詳細修復內容

### 修復的步驟：

#### 1. Deploy to Staging
- ✅ 添加 `vars.ENABLE_AWS_DEPLOYMENT == 'true'` 條件
- ✅ 只在啟用時執行 AWS 部署

#### 2. Deploy to Production
- ✅ 添加 `vars.ENABLE_AWS_DEPLOYMENT == 'true'` 條件
- ✅ 只在啟用時執行 AWS 部署

#### 3. Skip Deployment (新增)
- ✅ 當不啟用 AWS 部署時的替代步驟
- ✅ 提供清晰的狀態信息和啟用指南

#### 4. Rollback
- ✅ 只在啟用 AWS 部署且部署失敗時執行
- ✅ 避免在沒有部署時嘗試回滾

#### 5. Notification
- ✅ 重新設計為通用的流水線狀態通知
- ✅ 根據部署啟用狀態顯示不同信息
- ✅ 提供清晰的成功/失敗狀態

## ✅ 修復後的行為

### 當 AWS 部署未啟用時（默認）：
1. ✅ **Docker 構建**: 成功構建和推送鏡像
2. ✅ **SBOM 生成**: 生成軟體組件清單
3. ✅ **跳過部署**: 顯示跳過部署的通知
4. ✅ **清理**: 清理舊的容器鏡像
5. ✅ **通知**: 顯示成功狀態和啟用部署的指南

### 當 AWS 部署啟用時：
1. ✅ **Docker 構建**: 成功構建和推送鏡像
2. ✅ **部署到 Staging**: 部署到 AWS EKS
3. ✅ **部署到 Production**: 部署到生產環境（僅限標籤）
4. ✅ **回滾**: 部署失敗時自動回滾
5. ✅ **通知**: 顯示部署狀態

## 🔧 如何啟用 AWS 部署

如果將來需要啟用 AWS 部署：

### 1. 設置 Repository Variable
在 GitHub 倉庫設置中添加：
- Variable name: `ENABLE_AWS_DEPLOYMENT`
- Value: `true`

### 2. 配置 AWS Secrets
在 GitHub 倉庫 Secrets 中添加：
- `AWS_ACCESS_KEY_ID`: AWS 訪問密鑰 ID
- `AWS_SECRET_ACCESS_KEY`: AWS 秘密訪問密鑰
- `AWS_REGION`: AWS 區域（如 `us-west-2`）
- `EKS_CLUSTER_NAME`: EKS 集群名稱

### 3. 準備 Kubernetes 配置
確保 `k8s/staging/` 和 `k8s/production/` 目錄包含正確的部署配置。

## 🎯 最佳實踐

1. **漸進式部署**: 先啟用 staging，測試成功後再啟用 production
2. **環境隔離**: 使用不同的 AWS 賬戶或命名空間隔離環境
3. **監控和日誌**: 配置適當的監控和日誌收集
4. **回滾策略**: 確保有可靠的回滾機制

## 📚 相關資源

- [GitHub Actions Variables](https://docs.github.com/en/actions/learn-github-actions/variables)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [AWS EKS Documentation](https://docs.aws.amazon.com/eks/)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)

這個修復確保了 CI/CD 流水線在沒有 AWS 配置時也能正常運行，同時保留了將來啟用 AWS 部署的能力。