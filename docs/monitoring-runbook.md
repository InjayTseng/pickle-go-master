# Pickle Go - 監控與運維手冊

**文件版本:** 1.0
**建立日期:** 2026-01-21
**適用環境:** Production (picklego.tw)

---

## 目錄

1. [系統架構概覽](#1-系統架構概覽)
2. [監控工具](#2-監控工具)
3. [關鍵指標 (KPIs)](#3-關鍵指標-kpis)
4. [告警規則](#4-告警規則)
5. [緊急處理程序](#5-緊急處理程序)
6. [常見問題排查](#6-常見問題排查)
7. [定期維護任務](#7-定期維護任務)

---

## 1. 系統架構概覽

```
                                Internet
                                    │
                           ┌────────┴────────┐
                           │   Cloudflare    │
                           │   (CDN/WAF)     │
                           └────────┬────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
              ┌─────┴─────┐   ┌─────┴─────┐        │
              │ Cloud Run │   │ Cloud Run │        │
              │   (Web)   │   │   (API)   │        │
              │ Next.js   │   │  Go/Gin   │        │
              └─────┬─────┘   └─────┬─────┘        │
                    │               │               │
                    │         ┌─────┴─────┐        │
                    │         │ Cloud SQL │        │
                    │         │ PostgreSQL│        │
                    │         │ + PostGIS │        │
                    │         └───────────┘        │
                    │                              │
              ┌─────┴──────────────────────────────┘
              │
        ┌─────┴─────┐
        │  Sentry   │
        │ (Errors)  │
        └───────────┘
```

### 服務清單

| 服務 | 說明 | URL |
|------|------|-----|
| Web (Frontend) | Next.js 14 App Router | https://picklego.tw |
| API (Backend) | Go + Gin REST API | https://api.picklego.tw |
| Database | Cloud SQL PostgreSQL 15 + PostGIS | - |

---

## 2. 監控工具

### 2.1 Sentry (錯誤監控)

**存取連結:** https://sentry.io/organizations/YOUR_ORG/

**專案設置:**
- `pickle-go-web` - 前端錯誤
- `pickle-go-api` - 後端錯誤

**主要功能:**
- 即時錯誤追蹤
- 錯誤聚合與分析
- 效能監控 (Transaction)
- Release 追蹤

**Sentry Dashboard 重點關注:**
```
Issues → 排序: Last Seen → 檢查最新錯誤
Performance → 查看 P95 Transaction 時間
Releases → 確認最新版本的健康狀態
```

**常用篩選:**
```
# 只看 Production 錯誤
environment:production

# 只看特定版本
release:v1.0.0

# 只看高優先級
is:unresolved level:error
```

### 2.2 GCP Cloud Monitoring

**存取連結:** https://console.cloud.google.com/monitoring

**Cloud Run 指標:**
- Request Count
- Request Latency (P50, P95, P99)
- Container Instance Count
- Billable Container Instance Time
- Memory Utilization
- CPU Utilization

**建立 Dashboard:**
```
1. 前往 Monitoring → Dashboards
2. Create Dashboard
3. 添加以下 Widgets:
   - Request count (Cloud Run)
   - Request latency (Cloud Run)
   - Container instance count (Cloud Run)
   - CPU utilization (Cloud Run)
   - Memory utilization (Cloud Run)
```

### 2.3 Cloud SQL 監控

**Cloud SQL 指標:**
- CPU Utilization
- Memory Utilization
- Connections
- Read/Write Operations
- Disk Usage

**存取連結:** https://console.cloud.google.com/sql/instances/pickle-go-db

### 2.4 Uptime Checks

**設置 Uptime Check:**
```
1. 前往 Monitoring → Uptime checks
2. Create Uptime Check:
   - Name: pickle-go-web-health
   - Target: https://picklego.tw
   - Check frequency: 1 minute
   - Response timeout: 10s

3. Create Uptime Check:
   - Name: pickle-go-api-health
   - Target: https://api.picklego.tw/health
   - Check frequency: 1 minute
   - Response timeout: 10s
```

---

## 3. 關鍵指標 (KPIs)

### 3.1 可用性指標

| 指標 | 目標 | 告警閾值 |
|------|------|----------|
| 服務可用性 | >= 99.5% | < 99% |
| API Success Rate | >= 99% | < 98% |
| Error Rate | < 1% | > 2% |

### 3.2 效能指標

| 指標 | 目標 | 告警閾值 |
|------|------|----------|
| API P95 Latency | < 500ms | > 1s |
| API P99 Latency | < 1s | > 2s |
| Web LCP | < 2.5s | > 4s |
| Web FID | < 100ms | > 300ms |

### 3.3 資源指標

| 指標 | 正常範圍 | 告警閾值 |
|------|----------|----------|
| Cloud Run CPU | < 50% | > 80% |
| Cloud Run Memory | < 70% | > 90% |
| Cloud SQL CPU | < 50% | > 80% |
| Cloud SQL Memory | < 70% | > 90% |
| Cloud SQL Connections | < 80% max | > 90% max |

### 3.4 業務指標

| 指標 | 說明 | 追蹤方式 |
|------|------|----------|
| DAU | 每日活躍用戶 | Google Analytics |
| 新建活動數 | 每日新建活動 | 資料庫查詢 |
| 報名數 | 每日報名次數 | 資料庫查詢 |
| 登入成功率 | LINE 登入成功比例 | Sentry/Logs |

---

## 4. 告警規則

### 4.1 Sentry 告警

在 Sentry 設置以下告警:

```yaml
# 高頻錯誤告警
Alert Name: High Error Rate
Conditions:
  - Number of events > 50 in 10 minutes
Actions:
  - Send email notification
  - Send Slack notification (if configured)

# 新錯誤告警
Alert Name: New Issue Alert
Conditions:
  - A new issue is created
  - Event level = error or fatal
Actions:
  - Send email notification
```

### 4.2 GCP Alerting Policies

建立以下告警政策:

```yaml
# 1. API 高延遲告警
Alert Name: API High Latency
Resource: Cloud Run - pickle-go-api
Metric: Request latency (P95)
Condition: Above 1000ms for 5 minutes
Notification: Email, Slack

# 2. API 錯誤率告警
Alert Name: API Error Rate
Resource: Cloud Run - pickle-go-api
Metric: Request count (filtered by response_code >= 500)
Condition: Error rate > 5% for 5 minutes
Notification: Email, Slack

# 3. 服務不可用告警
Alert Name: Service Unavailable
Resource: Uptime Check - pickle-go-api-health
Condition: Failure for 2 consecutive checks
Notification: Email, Slack, PagerDuty (if configured)

# 4. Cloud SQL 高連線數告警
Alert Name: Database High Connections
Resource: Cloud SQL - pickle-go-db
Metric: PostgreSQL connections
Condition: > 90% of max_connections
Notification: Email

# 5. Cloud SQL 磁碟空間告警
Alert Name: Database Disk Usage
Resource: Cloud SQL - pickle-go-db
Metric: Disk utilization
Condition: > 80%
Notification: Email
```

### 4.3 告警通知管道

| 等級 | 條件 | 通知方式 |
|------|------|----------|
| P0 Critical | 服務完全不可用 | Slack + Email + 電話 |
| P1 High | 核心功能受影響 | Slack + Email |
| P2 Medium | 效能下降 | Slack + Email |
| P3 Low | 非緊急問題 | Email |

---

## 5. 緊急處理程序

### 5.1 服務完全不可用

**症狀:**
- https://picklego.tw 無法存取
- Uptime check 失敗
- 用戶回報無法使用

**處理步驟:**

```bash
# 1. 確認服務狀態
gcloud run services describe pickle-go-web --region asia-east1
gcloud run services describe pickle-go-api --region asia-east1

# 2. 檢查最近部署
gcloud run revisions list --service pickle-go-web --region asia-east1 --limit 5
gcloud run revisions list --service pickle-go-api --region asia-east1 --limit 5

# 3. 查看日誌
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=pickle-go-api" --limit 50 --format json

# 4. 如需回滾
gcloud run services update-traffic pickle-go-api \
  --to-revisions=<previous-revision>=100 \
  --region asia-east1

gcloud run services update-traffic pickle-go-web \
  --to-revisions=<previous-revision>=100 \
  --region asia-east1
```

**回滾決策標準:**
- Error rate > 10% 持續 5 分鐘
- P95 latency > 3s 持續 5 分鐘
- 核心功能無法使用

### 5.2 資料庫連線問題

**症狀:**
- API 返回 500 錯誤
- 日誌顯示 "connection refused" 或 "too many connections"

**處理步驟:**

```bash
# 1. 檢查 Cloud SQL 狀態
gcloud sql instances describe pickle-go-db --format="value(state)"

# 2. 檢查連線數
gcloud sql operations list --instance=pickle-go-db --limit 5

# 3. 如需重啟資料庫 (謹慎操作)
gcloud sql instances restart pickle-go-db

# 4. 調整連線數限制 (如有必要)
gcloud sql instances patch pickle-go-db --database-flags max_connections=200
```

### 5.3 LINE 登入失敗

**症狀:**
- 用戶無法完成 LINE 登入
- Sentry 顯示 OAuth 相關錯誤

**處理步驟:**

1. **檢查 LINE Channel 狀態:**
   - 登入 LINE Developers Console
   - 確認 Channel 狀態為 "Published"
   - 確認 Callback URL 設置正確

2. **檢查環境變數:**
   ```bash
   # 確認 Secret Manager 中的值
   gcloud secrets versions access latest --secret=pickle-go-line-channel-id
   gcloud secrets versions access latest --secret=pickle-go-line-channel-secret
   ```

3. **查看相關日誌:**
   ```bash
   gcloud logging read "resource.type=cloud_run_revision AND textPayload:line" --limit 20
   ```

### 5.4 效能下降

**症狀:**
- P95 latency 上升
- 用戶反映載入緩慢

**處理步驟:**

1. **查看 Cloud Run metrics:**
   ```bash
   # 檢查 instance 數量和 CPU 使用率
   gcloud monitoring metrics list --filter="metric.type:run.googleapis.com"
   ```

2. **檢查是否需要擴展:**
   ```bash
   # 調整最大 instance 數
   gcloud run services update pickle-go-api \
     --max-instances 20 \
     --region asia-east1
   ```

3. **檢查慢查詢 (Cloud SQL):**
   - 前往 Cloud SQL → Operations → Query Insights
   - 找出執行時間長的查詢

---

## 6. 常見問題排查

### 6.1 查看應用程式日誌

```bash
# API 日誌
gcloud logging read "resource.type=cloud_run_revision \
  AND resource.labels.service_name=pickle-go-api \
  AND severity>=ERROR" \
  --limit 50 \
  --format="table(timestamp, textPayload)"

# Web 日誌
gcloud logging read "resource.type=cloud_run_revision \
  AND resource.labels.service_name=pickle-go-web \
  AND severity>=WARNING" \
  --limit 50 \
  --format="table(timestamp, textPayload)"
```

### 6.2 檢查特定請求

```bash
# 找出 5xx 錯誤
gcloud logging read "resource.type=cloud_run_revision \
  AND httpRequest.status>=500" \
  --limit 20 \
  --format json

# 找出慢請求 (> 1s)
gcloud logging read "resource.type=cloud_run_revision \
  AND httpRequest.latency>1s" \
  --limit 20 \
  --format json
```

### 6.3 資料庫查詢

```sql
-- 檢查活躍連線數
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- 檢查長時間執行的查詢
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
  AND state = 'active';

-- 終止長時間查詢 (謹慎)
SELECT pg_terminate_backend(pid);
```

### 6.4 常見錯誤代碼

| 錯誤碼 | 可能原因 | 解決方案 |
|--------|----------|----------|
| 401 | JWT Token 無效或過期 | 檢查 JWT_SECRET 設置 |
| 403 | 權限不足 | 檢查 CORS 設置 |
| 500 | 伺服器內部錯誤 | 查看 Sentry 或日誌 |
| 502 | Cloud Run 啟動失敗 | 檢查 container 日誌 |
| 503 | 服務過載或不可用 | 檢查擴展設置 |

---

## 7. 定期維護任務

### 7.1 每日檢查

- [ ] 查看 Sentry 新錯誤
- [ ] 檢查 Uptime 狀態
- [ ] 查看昨日 Error Rate

### 7.2 每週檢查

- [ ] 審查 Cloud Run 資源使用
- [ ] 審查 Cloud SQL 效能
- [ ] 清理已解決的 Sentry issues
- [ ] 檢查磁碟空間使用

### 7.3 每月檢查

- [ ] 審查並優化 Cloud Run 配置
- [ ] 審查 API 回應時間趨勢
- [ ] 檢查 SSL 憑證到期日
- [ ] 審查帳單與成本優化
- [ ] 更新依賴套件 (如有安全更新)

### 7.4 資料庫維護

```sql
-- 每週執行: 更新統計資訊
ANALYZE;

-- 每月執行: 清理死元組
VACUUM ANALYZE;

-- 檢查索引使用情況
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
```

---

## 附錄

### A. 緊急聯絡清單

| 角色 | 姓名 | 電話 | Email |
|------|------|------|-------|
| Tech Lead | | | |
| On-Call Primary | | | |
| On-Call Secondary | | | |

### B. 外部服務聯絡

| 服務 | 支援連結 |
|------|----------|
| GCP Support | https://cloud.google.com/support |
| LINE Developers | https://developers.line.biz/en/ |
| Sentry Support | https://sentry.io/support/ |
| Cloudflare Support | https://support.cloudflare.com |

### C. 相關文件

| 文件 | 連結 |
|------|------|
| 系統架構 | planning/MULTI_AGENT_PLAN-V0.md |
| 回歸測試 | docs/regression-test-checklist.md |
| 上線清單 | docs/launch-checklist.md |

---

*最後更新: 2026-01-21*
