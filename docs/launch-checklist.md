# Pickle Go - Go/No-Go 上線檢查清單

**文件版本:** 1.0
**建立日期:** 2026-01-21
**目標上線日期:** ____________

---

## 概述

本文件為 Pickle Go 平台正式上線前的 Go/No-Go 決策清單。所有 **必要條件** 必須全部通過才能執行上線。

---

## 決策摘要

| 類別 | 狀態 | 負責人 | 確認日期 |
|------|------|--------|----------|
| 功能完整性 | [ ] 通過 | | |
| 效能指標 | [ ] 通過 | | |
| 安全性 | [ ] 通過 | | |
| 基礎設施 | [ ] 通過 | | |
| 營運準備 | [ ] 通過 | | |

**最終決策:** [ ] **GO** / [ ] **NO-GO**

**決策日期:** ____________

**決策者簽核:** ____________

---

## 1. 功能完整性 (必要)

### 1.1 核心用戶流程

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| 建立活動 | 團主可成功建立活動並取得分享連結 | [ ] | |
| 分享活動 | 短網址可正確開啟活動頁面 | [ ] | |
| LINE 分享預覽 | 分享到 LINE 顯示正確的 OG 卡片 | [ ] | |
| 瀏覽活動 | 用戶可在地圖上瀏覽附近活動 | [ ] | |
| 報名活動 | 球友可成功報名活動 | [ ] | |
| 滿團處理 | 活動滿員時正確顯示狀態並進入候補 | [ ] | |
| 候補遞補 | 正取取消時候補自動遞補 | [ ] | |
| 取消報名 | 用戶可取消自己的報名 | [ ] | |

### 1.2 LINE Login

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| LINE 授權頁面 | 點擊登入可跳轉到 LINE 授權 | [ ] | |
| OAuth 回調 | 授權後正確建立用戶並發 Token | [ ] | |
| 用戶資訊 | 顯示正確的 LINE 用戶名和頭像 | [ ] | |
| 登出功能 | 可正常登出並清除狀態 | [ ] | |
| Token 更新 | JWT Token 過期處理正常 | [ ] | |

### 1.3 行動裝置體驗

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| iOS Safari | iPhone 12+ 上所有功能正常 | [ ] | |
| Android Chrome | Pixel 5+ 上所有功能正常 | [ ] | |
| 響應式佈局 | 頁面在不同尺寸顯示正常 | [ ] | |
| 觸控操作 | 按鈕和互動元素易於點擊 | [ ] | |
| 地圖操作 | 手機上地圖縮放、移動順暢 | [ ] | |

---

## 2. 效能指標 (必要)

使用 [PageSpeed Insights](https://pagespeed.web.dev/) 或 Lighthouse 測試。

### 2.1 Core Web Vitals

| 指標 | 目標 | 實測值 | 狀態 |
|------|------|--------|------|
| LCP (Largest Contentful Paint) | < 2.5s | ______s | [ ] |
| FID (First Input Delay) | < 100ms | ______ms | [ ] |
| CLS (Cumulative Layout Shift) | < 0.1 | ______ | [ ] |

### 2.2 頁面效能

| 頁面 | Performance Score 目標 | 實測分數 | 狀態 |
|------|------------------------|----------|------|
| 首頁 (地圖) | >= 80 | ______ | [ ] |
| 活動詳情頁 | >= 80 | ______ | [ ] |
| 建立活動頁 | >= 75 | ______ | [ ] |

### 2.3 API 效能

| API | P95 目標 | 實測 P95 | 狀態 |
|-----|----------|----------|------|
| GET /events (地圖查詢) | < 500ms | ______ms | [ ] |
| GET /events/:id | < 200ms | ______ms | [ ] |
| POST /events | < 500ms | ______ms | [ ] |
| POST /events/:id/register | < 300ms | ______ms | [ ] |

---

## 3. 安全性 (必要)

### 3.1 認證與授權

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| JWT 驗證 | API 正確驗證 Token | [ ] | |
| 未授權處理 | 無效 Token 返回 401 | [ ] | |
| 權限檢查 | 只有團主可修改/取消自己的活動 | [ ] | |
| Token 儲存 | 使用 HttpOnly Cookie | [ ] | |

### 3.2 資料保護

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| HTTPS | 所有流量強制使用 HTTPS | [ ] | |
| CORS | 只允許指定來源 | [ ] | |
| SQL Injection | 使用參數化查詢 | [ ] | |
| Secrets 管理 | 敏感資料使用 Secret Manager | [ ] | |

### 3.3 第三方服務

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| LINE Channel | Production Channel 已設置 | [ ] | |
| Google Maps API | API Key 已設置使用限制 | [ ] | |
| Sentry DSN | 已設置正確的 DSN | [ ] | |

---

## 4. 基礎設施 (必要)

### 4.1 GCP Cloud Run

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| API Service | pickle-go-api 部署成功 | [ ] | |
| Web Service | pickle-go-web 部署成功 | [ ] | |
| 自動擴展 | Min: 0, Max: 10 已設置 | [ ] | |
| 資源限制 | CPU/Memory 設置合理 | [ ] | |

### 4.2 Cloud SQL

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| 資料庫連線 | API 可正常連線 Cloud SQL | [ ] | |
| PostGIS | PostGIS 擴展已啟用 | [ ] | |
| 備份策略 | 自動備份已啟用 | [ ] | |
| 連線數限制 | 設置合理的連線數上限 | [ ] | |

### 4.3 網域與 DNS

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| picklego.tw | 指向 Web Service | [ ] | |
| api.picklego.tw | 指向 API Service | [ ] | |
| SSL 憑證 | 自動更新憑證有效 | [ ] | |
| CDN | Cloudflare 已設置 (選用) | [ ] | |

### 4.4 監控

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| Sentry | 前後端錯誤監控已啟用 | [ ] | |
| Cloud Monitoring | GCP 監控儀表板已建立 | [ ] | |
| Uptime Check | 健康檢查已設置 | [ ] | |
| 告警規則 | 異常告警已設置 | [ ] | |

---

## 5. 營運準備 (必要)

### 5.1 種子內容

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| 種子團主 | 至少 5 位團主承諾開團 | [ ] | 名單: |
| 首週活動 | 目標 10 場活動有把握達成 | [ ] | |
| 測試活動 | 已清除所有測試資料 | [ ] | |

### 5.2 文件

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| 運維手冊 | monitoring-runbook.md 已建立 | [ ] | |
| 緊急聯絡 | 緊急處理流程已定義 | [ ] | |
| FAQ | 常見問題已準備 | [ ] | |

### 5.3 團隊準備

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| On-Call 排班 | 上線首週值班表已排定 | [ ] | |
| 通知管道 | Slack/LINE 告警群組已建立 | [ ] | |
| Rollback 流程 | 團隊熟悉回滾步驟 | [ ] | |

---

## 6. Bug 狀態 (必要)

### 6.1 已知 Bug 清單

| Bug ID | 嚴重度 | 說明 | 狀態 | 備註 |
|--------|--------|------|------|------|
| | P0 | | [ ] 已修復 | |
| | P1 | | [ ] 已修復 | |

### 6.2 Bug 統計

| 嚴重度 | 待修復 | 已修復 | 說明 |
|--------|--------|--------|------|
| P0 (Critical) | 0 | | 必須為 0 |
| P1 (High) | 0 | | 必須為 0 |
| P2 (Medium) | | | 可接受 |
| P3 (Low) | | | 可接受 |

---

## 7. 回歸測試 (必要)

| 項目 | 檢查說明 | 狀態 | 備註 |
|------|----------|------|------|
| 自動化測試 | CI Pipeline 全部通過 | [ ] | |
| 手動測試 | 回歸測試 checklist 全部通過 | [ ] | |
| 測試簽核 | QA 簽核通過 | [ ] | 簽核人: |

---

## 8. 上線執行清單

確認 Go 決策後，按以下步驟執行上線：

### 8.1 上線前 (T-30 分鐘)

- [ ] 通知相關人員即將上線
- [ ] 確認 On-Call 人員就位
- [ ] 確認監控儀表板已開啟
- [ ] 確認 rollback 腳本準備好

### 8.2 上線執行 (T-0)

```bash
# 觸發生產環境部署
# 方法 1: 透過 GitHub Actions
git tag v1.0.0
git push origin v1.0.0

# 方法 2: 手動觸發 Cloud Build
gcloud builds submit --config=cloudbuild.yaml
```

- [ ] 部署指令已執行
- [ ] Cloud Build 構建成功
- [ ] API Service 部署成功
- [ ] Web Service 部署成功

### 8.3 上線後驗證 (T+10 分鐘)

- [ ] https://picklego.tw 可正常存取
- [ ] https://api.picklego.tw/health 返回 200
- [ ] 執行快速冒煙測試
  - [ ] 首頁載入正常
  - [ ] LINE 登入可點擊
  - [ ] 地圖顯示正常
- [ ] Sentry 無新錯誤
- [ ] 服務 latency 正常

### 8.4 上線後監控 (T+1 小時)

- [ ] 持續監控 Error Rate
- [ ] 持續監控 Response Time
- [ ] 確認無用戶反映問題
- [ ] 通知團隊上線完成

---

## 9. Rollback 流程

如果上線後發現嚴重問題，執行回滾：

```bash
# 1. 查看之前的版本
gcloud run revisions list --service=pickle-go-api --region=asia-east1
gcloud run revisions list --service=pickle-go-web --region=asia-east1

# 2. 回滾 API
gcloud run services update-traffic pickle-go-api \
  --to-revisions=<previous-revision>=100 \
  --region=asia-east1

# 3. 回滾 Web
gcloud run services update-traffic pickle-go-web \
  --to-revisions=<previous-revision>=100 \
  --region=asia-east1
```

**Rollback 觸發條件:**
- Error rate > 5%
- P95 latency > 2s
- 發現 P0 級 Bug

---

## 附錄

### A. 緊急聯絡清單

| 角色 | 姓名 | 電話 | LINE |
|------|------|------|------|
| Tech Lead | | | |
| Backend | | | |
| Frontend | | | |
| DevOps | | | |

### B. 相關連結

| 項目 | 連結 |
|------|------|
| GCP Console | https://console.cloud.google.com/run |
| Sentry | https://sentry.io/organizations/YOUR_ORG |
| GitHub Repo | https://github.com/YOUR_ORG/pickle-go |
| CI/CD | https://github.com/YOUR_ORG/pickle-go/actions |

---

*最後更新: 2026-01-21*
