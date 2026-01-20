# Pickle Go 架構審查報告

**審查日期:** 2026-01-21
**審查者:** Architect-Planner Agent
**專案版本:** M6 (Public Launch)

---

## 執行摘要

經過對 Pickle Go 專案的全面架構審查，本報告的結論是：

> **決策: STABLE (架構穩定)**

專案架構設計合理、程式碼品質良好、文件完整，已具備 Public Launch 的條件。雖有若干可改進項目，但均屬於優化性質，不影響系統穩定性和核心功能。

---

## 1. 系統架構審查

### 1.1 前後端分離架構

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 架構模式 | PASS | 採用清晰的前後端分離架構 |
| 前端技術棧 | PASS | Next.js 14 (App Router) + TanStack Query |
| 後端技術棧 | PASS | Go + Gin + sqlx |
| API 通訊 | PASS | RESTful JSON API |
| 狀態管理 | PASS | 前端使用 Context + TanStack Query |

**架構圖:**
```
                    ┌─────────────────┐
                    │   Client        │
                    │   (Browser)     │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Next.js 14    │
                    │   (Vercel/GCP)  │
                    │   - SSR/SSG     │
                    │   - App Router  │
                    └────────┬────────┘
                             │ REST API
                    ┌────────▼────────┐
                    │   Go API        │
                    │   (Cloud Run)   │
                    │   - Gin         │
                    │   - JWT Auth    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │   + PostGIS     │
                    │   (Cloud SQL)   │
                    └─────────────────┘
```

**優點:**
- 前後端完全解耦，可獨立部署和擴展
- Next.js App Router 提供良好的 SEO 支援
- Go 後端性能優異，適合高並發場景
- 使用 PostGIS 實現高效的地理空間查詢

**建議 (非必要):**
- 可考慮引入 API Gateway 層進行流量管理
- 未來規模擴大時可考慮引入 GraphQL

---

### 1.2 API 設計 (RESTful 規範)

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| URL 結構 | PASS | 遵循資源導向設計 `/api/v1/events`, `/api/v1/users` |
| HTTP 方法 | PASS | 正確使用 GET/POST/PUT/DELETE |
| 狀態碼 | PASS | 200/201/400/401/403/404/500 使用正確 |
| 錯誤格式 | PASS | 統一的錯誤回應格式 `{success, error: {code, message}}` |
| 版本控制 | PASS | URL 路徑版本 `/api/v1` |
| 分頁 | PASS | 支援 limit/offset 分頁 |
| 過濾 | PASS | 支援 lat/lng/radius/skill_level 等查詢參數 |

**API 端點清單:**
```
POST   /api/v1/auth/line/callback     - LINE 登入回調
POST   /api/v1/auth/refresh           - 刷新 Token
POST   /api/v1/auth/logout            - 登出
GET    /api/v1/users/me               - 取得當前用戶
GET    /api/v1/users/me/events        - 取得我的活動
GET    /api/v1/users/me/registrations - 取得我的報名
GET    /api/v1/events                 - 列出活動 (支援地理查詢)
GET    /api/v1/events/:id             - 取得活動詳情
POST   /api/v1/events                 - 建立活動
PUT    /api/v1/events/:id             - 更新活動
DELETE /api/v1/events/:id             - 取消活動
POST   /api/v1/events/:id/register    - 報名活動
DELETE /api/v1/events/:id/register    - 取消報名
GET    /api/v1/events/:id/registrations - 取得報名清單
```

**優點:**
- OpenAPI 3.0 規格文件完整 (`apps/api/api/openapi.yaml`)
- 統一的回應格式便於前端處理
- 良好的錯誤碼設計

---

### 1.3 資料庫 Schema 設計

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 正規化 | PASS | 適當的第三正規化 |
| 主鍵設計 | PASS | 使用 UUID，避免自增 ID 暴露 |
| 外鍵約束 | PASS | 正確設置級聯刪除 |
| 索引設計 | PASS | 合理的索引覆蓋查詢需求 |
| 地理空間 | PASS | PostGIS GIST 索引 |
| 資料完整性 | PASS | CHECK 約束確保資料有效性 |

**ER Diagram:**
```
users (1) ───< (N) events
  │                  │
  │                  │ (1)
  │                  │
  └──(N) registrations >───(N)
  │
  └──(N) notifications
```

**Schema 亮點:**
- `events.location_point` 使用 PostGIS GEOGRAPHY 類型支援精確距離計算
- `registrations` 表的 `UNIQUE(event_id, user_id)` 防止重複報名
- 自動生成 `short_code` 的觸發器設計良好
- `event_summary` 視圖預先計算報名人數

---

### 1.4 認證機制

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 身份驗證 | PASS | LINE Login OAuth 2.0 |
| Token 類型 | PASS | JWT (Access + Refresh Token) |
| Token 過期 | PASS | Access: 7天, Refresh: 30天 |
| CSRF 保護 | PASS | OAuth state 參數驗證 |
| 中間件 | PASS | `AuthRequired()` 統一驗證 |

**認證流程:**
```
1. 用戶點擊 LINE 登入
2. 前端生成 state 並存入 sessionStorage (CSRF 保護)
3. 重導向到 LINE 授權頁面
4. 用戶授權後帶 code 回調
5. 前端驗證 state
6. 後端用 code 換取 LINE access token
7. 後端取得 LINE 用戶資料
8. 後端 upsert 用戶並簽發 JWT
9. 前端儲存 token 到 Cookie
```

**安全措施:**
- JWT 簽名使用 HMAC-SHA256
- Token 驗證檢查簽名方法防止算法混淆攻擊
- 認證端點有嚴格的速率限制 (10 req/min)

**建議 (非必要):**
- 考慮將 JWT Secret 移至 Secret Manager
- 可考慮實現 Token 黑名單機制

---

## 2. 程式碼品質審查

### 2.1 程式碼結構

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 專案結構 | PASS | 遵循標準 Go/Next.js 專案結構 |
| 分層架構 | PASS | Handler -> Service -> Repository |
| 關注點分離 | PASS | DTO/Model/Handler 職責清晰 |
| 命名規範 | PASS | 遵循 Go/TypeScript 命名慣例 |

**後端架構分層:**
```
apps/api/
├── cmd/server/main.go       # 應用程式入口
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 處理器 (Controller)
│   ├── service/             # 業務邏輯層
│   ├── repository/          # 資料存取層
│   ├── model/               # 領域模型
│   ├── dto/                 # 資料傳輸物件
│   ├── middleware/          # 中間件
│   └── database/            # 資料庫連線
└── pkg/                     # 可重用套件
    ├── jwt/                 # JWT 工具
    ├── line/                # LINE SDK
    ├── geo/                 # 地理工具
    └── shortcode/           # 短碼生成
```

**前端架構:**
```
apps/web/src/
├── app/                     # Next.js App Router 頁面
│   ├── (auth)/              # 認證相關路由群組
│   ├── (main)/              # 主要功能路由群組
│   └── api/og/              # OG Image API Route
├── components/              # React 元件
│   ├── ui/                  # 基礎 UI 元件 (shadcn/ui)
│   ├── map/                 # 地圖相關元件
│   ├── event/               # 活動相關元件
│   └── seo/                 # SEO 元件
├── hooks/                   # Custom Hooks
├── lib/                     # 工具函式
└── contexts/                # React Contexts
```

---

### 2.2 重複程式碼檢查

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| API 回應格式 | PASS | 使用 `dto.SuccessResponse/ErrorResponse` |
| 錯誤處理 | PASS | 統一的錯誤類型和處理 |
| 驗證邏輯 | PASS | 中間件統一處理 |
| 前端 API 調用 | PASS | `apiClient` 統一封裝 |

**發現的輕微重複:**
1. Handler 中有一些 Legacy 函式保留向後相容，可在未來版本移除
2. Repository 層的 SQL 查詢有部分重複的 SELECT 欄位清單

---

### 2.3 錯誤處理

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 後端錯誤處理 | PASS | 統一錯誤碼和訊息 |
| 前端錯誤處理 | PASS | ApiError 類別封裝 |
| 用戶友好訊息 | PASS | 提供中文錯誤訊息 |
| 錯誤日誌 | PASS | Sentry 整合 |

**錯誤碼設計範例:**
```go
// 認證相關
UNAUTHORIZED, INVALID_TOKEN, TOKEN_EXPIRED

// 資料相關
NOT_FOUND, VALIDATION_ERROR

// 業務邏輯
ALREADY_REGISTERED, HOST_CANNOT_REGISTER,
EVENT_CANCELLED, EVENT_COMPLETED
```

---

### 2.4 安全漏洞檢查

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| SQL Injection | PASS | 使用參數化查詢 ($1, $2...) |
| XSS | PASS | React 自動轉義 |
| CSRF | PASS | OAuth state 驗證 |
| 敏感資料暴露 | PASS | JWT 不包含敏感資訊 |
| 速率限制 | PASS | 多層次速率限制 |
| CORS | PASS | 白名單機制 |
| HTTPS | PASS | 生產環境強制 HTTPS |

**速率限制配置:**
- 一般 API: 100 req/min/IP
- 認證端點: 10 req/min/IP
- API 端點: 60 req/min/IP

**建議 (非必要):**
- 可考慮實施 CSP (Content Security Policy)
- 可考慮加入 HTTP Security Headers

---

## 3. 文件完整性審查

### 3.1 README.md

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| 專案簡介 | PASS | 清楚說明專案目的 |
| 技術棧 | PASS | 完整列出使用技術 |
| 快速開始 | PASS | 提供詳細設置步驟 |
| 開發指令 | PASS | 列出所有可用指令 |
| 環境變數 | PASS | 表格式文件清晰 |
| 專案結構 | PASS | 提供目錄樹說明 |

---

### 3.2 API 文件

| 評估項目 | 狀態 | 說明 |
|---------|------|------|
| OpenAPI 規格 | PASS | `apps/api/api/openapi.yaml` 完整 |
| 端點描述 | PASS | 所有端點有描述和範例 |
| Schema 定義 | PASS | Request/Response Schema 完整 |
| 錯誤回應 | PASS | 文件化所有錯誤狀況 |

---

### 3.3 部署與運維文件

| 文件 | 狀態 | 說明 |
|------|------|------|
| `docs/deployment.md` | PASS | GCP Cloud Run 部署指南 |
| `docs/launch-checklist.md` | PASS | Go/No-Go 上線檢查清單 |
| `docs/monitoring-runbook.md` | PASS | 監控與故障排除手冊 |
| `docs/database-schema.md` | PASS | 資料庫 Schema 詳細文件 |
| `docs/regression-test-checklist.md` | PASS | 回歸測試清單 |
| `docs/cloud-sql-setup.md` | PASS | Cloud SQL 設置指南 |
| `.env.example` | PASS | 環境變數範本 |
| `cloudbuild.yaml` | PASS | Cloud Build CI/CD 配置 |

---

## 4. 技術債務評估

### 4.1 已識別的技術債務

| 優先級 | 項目 | 說明 | 建議處理時機 |
|--------|------|------|--------------|
| P3 | Legacy Handlers | Handler 檔案中保留了舊的函式簽名 | v1.1 |
| P3 | Redis 整合 | Config 中定義了 RedisURL 但未使用 | v1.2 (快取層) |
| P3 | 通知系統 | TODO: Send notification 註解 | v1.1 (推播功能) |
| P4 | 型別安全 | 前端部分 `unknown` 類型可強化 | 持續改進 |
| P4 | 測試覆蓋 | E2E 測試可擴充覆蓋更多邊界情況 | 持續改進 |

---

### 4.2 可優化項目 (非必要)

| 項目 | 說明 | 預期效益 |
|------|------|----------|
| 連線池優化 | 調整 DB 連線池參數 | 高負載下效能提升 |
| 快取層 | 引入 Redis 快取熱門活動 | 降低 DB 負載 |
| 圖片優化 | 頭像圖片 CDN + 裁切 | 加快載入速度 |
| 批次通知 | 通知系統批次處理 | 提升效率 |

---

## 5. 審查結論

### 5.1 整體評分

| 審查類別 | 評分 | 說明 |
|---------|------|------|
| 系統架構 | A | 設計合理，技術選型適當 |
| API 設計 | A | 遵循 RESTful 規範，文件完整 |
| 資料庫設計 | A | 正規化適當，索引合理 |
| 認證安全 | A | OAuth + JWT 實作正確 |
| 程式碼品質 | A- | 結構清晰，有少量可優化項目 |
| 文件完整性 | A | 文件全面且維護良好 |
| 測試覆蓋 | B+ | 有 E2E 測試，可增加單元測試 |

**綜合評分: A-**

---

### 5.2 決策結果

## **STABLE (架構穩定)**

Pickle Go 專案已完成 M0-M6 所有里程碑的開發，架構設計成熟穩定，具備以下特點：

1. **前後端分離架構清晰** - Next.js + Go 的組合提供良好的效能和開發體驗
2. **API 設計規範** - 遵循 RESTful 最佳實踐，有完整的 OpenAPI 文件
3. **資料庫設計合理** - 正規化適當，PostGIS 支援地理查詢
4. **安全機制完善** - OAuth 認證、JWT、速率限制、CORS 等保護措施
5. **文件齊全** - 從開發到部署的完整文件覆蓋
6. **運維準備就緒** - 監控、告警、rollback 流程已建立

---

### 5.3 後續改進建議

以下為建議的後續改進方向，可在產品穩定運營後逐步實施：

#### 短期 (v1.1)
- [ ] 移除 Legacy Handler 函式
- [ ] 實作推播通知功能
- [ ] 增加單元測試覆蓋率

#### 中期 (v1.2)
- [ ] 引入 Redis 快取層
- [ ] 實作活動搜尋功能
- [ ] 增加用戶個人資料編輯

#### 長期 (v2.0)
- [ ] 考慮微服務拆分 (如需要)
- [ ] 引入訊息佇列處理非同步任務
- [ ] 國際化 (i18n) 支援

---

## 附錄

### A. 審查涵蓋的檔案清單

**後端 (apps/api):**
- `cmd/server/main.go`
- `internal/config/config.go`
- `internal/handler/*.go`
- `internal/middleware/*.go`
- `internal/model/*.go`
- `internal/repository/*.go`
- `internal/service/*.go`
- `pkg/jwt/jwt.go`
- `pkg/line/client.go`
- `migrations/*.sql`
- `api/openapi.yaml`

**前端 (apps/web):**
- `src/app/**/*.tsx`
- `src/components/**/*.tsx`
- `src/contexts/*.tsx`
- `src/hooks/*.ts`
- `src/lib/*.ts`

**文件:**
- `README.md`
- `docs/*.md`
- `.env.example`
- `cloudbuild.yaml`

---

### B. 審查方法論

本次架構審查採用以下方法：

1. **靜態程式碼分析** - 檢視原始碼結構和品質
2. **架構檢視** - 評估系統設計和技術選型
3. **安全評估** - 檢查常見安全漏洞
4. **文件審查** - 確認文件完整性和可用性
5. **最佳實踐比對** - 與業界標準比較

---

*報告結束*

*Generated by Architect-Planner Agent on 2026-01-21*
