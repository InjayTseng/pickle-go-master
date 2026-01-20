# Pickle Go - 技術實作計劃

**文件版本：** 1.0
**建立日期：** 2026-01-20
**對應 PRD：** PRD-V1.md
**狀態：** Draft

---

## 1. 專案概述

### 1.1 專案背景與目標

Pickle Go 是一個 Web-First 的匹克球揪團平台，目標是讓用戶在 30 秒內完成「找到附近球局並報名」的完整流程。

**核心功能：**
- 地圖模式瀏覽附近活動
- Line 登入一鍵報名
- 快速建立活動並分享
- 自動候補遞補機制
- SEO 優化活動頁面

### 1.2 技術約束與假設

| 項目 | 約束/假設 |
|------|----------|
| 目標平台 | Web-First (Mobile Safari/Chrome 優先) |
| 預期規模 | MVP: 100 活動/月, 500 用戶 |
| 開發時程 | 12 週 (M0-M6) |
| 團隊規模 | 假設 2-3 位全端工程師 |
| 預算限制 | 使用免費層或低成本雲端服務 |

---

## 2. 架構決策記錄 (ADR)

### ADR-001: 專案結構選擇 - Monorepo

- **狀態**: 已決定
- **背景**: 需要決定前後端是否放在同一個 Repository
- **決策**: 採用 Monorepo 架構，使用 pnpm workspace
- **備選方案**:
  - 方案 A: Monorepo (選擇)
  - 方案 B: 分離式 Repository
- **理由**:
  - 小團隊更易於協作與程式碼共享
  - 統一版本控制與 CI/CD
  - 共用 TypeScript 型別定義
  - 減少跨 Repo 同步成本
- **影響**: 需要設置 workspace 配置，CI/CD 需支援 monorepo

### ADR-002: 後端語言選擇 - Go + Gin

- **狀態**: 已決定
- **背景**: PRD 建議 Go + Gin 或 Node.js + Fastify
- **決策**: 採用 Go + Gin 框架
- **備選方案**:
  - 方案 A: Go + Gin (選擇)
  - 方案 B: Node.js + Fastify
- **理由**:
  - 高效能，適合地理位置查詢密集場景
  - 編譯型語言，部署產物單一二進位檔
  - 內建並發處理，適合未來水平擴展
  - Cloud Run 原生支援 Go
- **影響**:
  - 團隊需具備 Go 開發經驗
  - 無法與前端共用 TypeScript 型別 (需使用 OpenAPI 生成)

### ADR-003: 前端框架選擇 - Next.js 14+ App Router

- **狀態**: 已決定
- **背景**: PRD 明確指定 Next.js 14+ 搭配 App Router
- **決策**: 採用 Next.js 14 App Router + React Server Components
- **理由**:
  - SSR/SSG 支援 SEO 需求 (FR-C01)
  - App Router 提供更好的伺服器端資料獲取
  - 內建圖片優化與路由
  - Vercel 或 Cloud Run 部署友善
- **影響**:
  - 需要熟悉 React Server Components 新範式
  - 部分 Client Component 需明確標註 "use client"

### ADR-004: 資料庫選擇 - PostgreSQL + PostGIS

- **狀態**: 已決定
- **背景**: 需要支援地理位置查詢 (FR-A01 地圖模式)
- **決策**: PostgreSQL 15+ 搭配 PostGIS 擴展
- **備選方案**:
  - 方案 A: PostgreSQL + PostGIS (選擇)
  - 方案 B: MongoDB + Geospatial Index
- **理由**:
  - PostGIS 是業界最成熟的地理空間解決方案
  - 關聯式資料模型適合 User-Event-Registration 結構
  - GCP Cloud SQL 原生支援
  - 支援 ACID 事務保證報名一致性
- **影響**:
  - 需要額外設置 PostGIS 擴展
  - Cloud SQL 有最低費用門檻

### ADR-005: 認證機制 - Line Login + JWT

- **狀態**: 已決定
- **背景**: PRD 要求 Line Login OAuth 2.0，Session 過期 7 天
- **決策**: Line Login 取得身份後，後端發放 JWT Token
- **理由**:
  - Line Login 符合台灣用戶習慣
  - JWT 無狀態，適合 Cloud Run 無伺服器架構
  - 避免 Session 儲存負擔
- **影響**:
  - 需提前申請 Line Login Channel
  - 需處理 JWT Refresh Token 機制

### ADR-006: 快取策略 - Redis (Cloud Memorystore)

- **狀態**: 已決定
- **背景**: 需要快取熱門活動、減少資料庫負載
- **決策**: 使用 Redis 做為快取層
- **理由**:
  - 地圖 Pin 資料適合快取 (頻繁讀取)
  - Session Token 可儲存於 Redis (備用方案)
  - GCP Memorystore 全代管
- **影響**:
  - MVP 階段可選擇性使用，先以資料庫為主
  - 增加基礎設施成本

### ADR-007: API 設計風格 - RESTful + OpenAPI

- **狀態**: 已決定
- **背景**: 需要定義前後端 API 契約
- **決策**: RESTful API，使用 OpenAPI 3.0 規格文件
- **理由**:
  - REST 風格直觀，學習曲線低
  - OpenAPI 可自動生成 TypeScript Client
  - 便於文件化與測試
- **影響**:
  - 需維護 OpenAPI spec 文件
  - 複雜查詢可能需要多次 API 呼叫

---

## 3. 系統架構

### 3.1 架構圖

```
                                    ┌─────────────────────────────────────────────────────────┐
                                    │                    GCP Cloud Platform                    │
                                    │                                                          │
┌──────────────┐                    │  ┌────────────────┐       ┌────────────────────────┐   │
│              │                    │  │                │       │                        │   │
│   Browser    │◄───── HTTPS ──────┼──┤  Cloud CDN     │       │   Cloud Run            │   │
│   (Mobile)   │                    │  │  (Cloudflare)  │       │                        │   │
│              │                    │  │                │       │  ┌──────────────────┐  │   │
└──────────────┘                    │  └───────┬────────┘       │  │  Frontend        │  │   │
                                    │          │                │  │  (Next.js 14)    │  │   │
                                    │          ▼                │  │  - SSR/SSG       │  │   │
                                    │  ┌────────────────┐       │  │  - React Query   │  │   │
                                    │  │                │       │  └────────┬─────────┘  │   │
                                    │  │  Cloud Load    │       │           │            │   │
                                    │  │  Balancer      │───────┤           │ REST API   │   │
                                    │  │                │       │           ▼            │   │
                                    │  └────────────────┘       │  ┌──────────────────┐  │   │
                                    │                           │  │  Backend         │  │   │
                                    │                           │  │  (Go + Gin)      │  │   │
                                    │                           │  │  - REST API      │  │   │
                                    │                           │  │  - JWT Auth      │  │   │
                                    │                           │  └────────┬─────────┘  │   │
                                    │                           │           │            │   │
                                    │                           └───────────┼────────────┘   │
                                    │                                       │                │
                                    │          ┌────────────────────────────┼───────┐        │
                                    │          │                            │       │        │
                                    │          ▼                            ▼       │        │
                                    │  ┌────────────────┐       ┌──────────────────┐│        │
                                    │  │                │       │                  ││        │
                                    │  │  Cloud SQL     │       │  Memorystore     ││        │
                                    │  │  (PostgreSQL   │       │  (Redis)         ││        │
                                    │  │   + PostGIS)   │       │                  ││        │
                                    │  │                │       │                  ││        │
                                    │  └────────────────┘       └──────────────────┘│        │
                                    │                                               │        │
                                    └───────────────────────────────────────────────┴────────┘

External Services:
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Line Login      │  │  Google Maps     │  │  Google Places   │
│  OAuth 2.0       │  │  JavaScript API  │  │  API             │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

### 3.2 專案目錄結構

```
pickle-go-master/
├── .github/
│   └── workflows/
│       ├── ci.yml                 # CI 流程 (lint, test, build)
│       └── deploy.yml             # CD 流程 (部署到 Cloud Run)
├── apps/
│   ├── web/                       # Next.js 前端應用
│   │   ├── src/
│   │   │   ├── app/               # App Router 頁面
│   │   │   │   ├── (auth)/        # 認證相關頁面群組
│   │   │   │   │   └── login/
│   │   │   │   ├── (main)/        # 主要頁面群組
│   │   │   │   │   ├── page.tsx   # 首頁 (地圖)
│   │   │   │   │   └── events/
│   │   │   │   │       ├── [id]/
│   │   │   │   │       │   └── page.tsx  # 活動詳情頁
│   │   │   │   │       └── new/
│   │   │   │   │           └── page.tsx  # 建立活動頁
│   │   │   │   ├── api/           # API Routes (BFF 層)
│   │   │   │   │   └── auth/
│   │   │   │   │       └── [...nextauth]/
│   │   │   │   ├── layout.tsx
│   │   │   │   └── globals.css
│   │   │   ├── components/        # React 元件
│   │   │   │   ├── ui/            # shadcn/ui 元件
│   │   │   │   ├── map/           # 地圖相關元件
│   │   │   │   │   ├── EventMap.tsx
│   │   │   │   │   ├── EventPin.tsx
│   │   │   │   │   └── EventCard.tsx
│   │   │   │   ├── event/         # 活動相關元件
│   │   │   │   │   ├── EventForm.tsx
│   │   │   │   │   ├── EventDetail.tsx
│   │   │   │   │   └── RegistrationButton.tsx
│   │   │   │   └── layout/        # 版面元件
│   │   │   │       ├── Header.tsx
│   │   │   │       └── MobileNav.tsx
│   │   │   ├── lib/               # 工具函式
│   │   │   │   ├── api-client.ts  # API Client
│   │   │   │   ├── auth.ts        # 認證工具
│   │   │   │   └── utils.ts
│   │   │   ├── hooks/             # Custom Hooks
│   │   │   │   ├── useEvents.ts
│   │   │   │   └── useGeolocation.ts
│   │   │   └── types/             # TypeScript 型別
│   │   │       └── index.ts
│   │   ├── public/
│   │   │   └── images/
│   │   ├── next.config.js
│   │   ├── tailwind.config.ts
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   └── api/                       # Go 後端應用
│       ├── cmd/
│       │   └── server/
│       │       └── main.go        # 程式進入點
│       ├── internal/
│       │   ├── config/            # 配置管理
│       │   │   └── config.go
│       │   ├── handler/           # HTTP Handler
│       │   │   ├── auth.go
│       │   │   ├── event.go
│       │   │   ├── registration.go
│       │   │   └── user.go
│       │   ├── middleware/        # 中介軟體
│       │   │   ├── auth.go
│       │   │   ├── cors.go
│       │   │   └── logger.go
│       │   ├── model/             # 資料模型
│       │   │   ├── user.go
│       │   │   ├── event.go
│       │   │   └── registration.go
│       │   ├── repository/        # 資料存取層
│       │   │   ├── user_repo.go
│       │   │   ├── event_repo.go
│       │   │   └── registration_repo.go
│       │   ├── service/           # 業務邏輯層
│       │   │   ├── auth_service.go
│       │   │   ├── event_service.go
│       │   │   └── registration_service.go
│       │   └── dto/               # Data Transfer Objects
│       │       ├── request.go
│       │       └── response.go
│       ├── pkg/                   # 可重用套件
│       │   ├── jwt/
│       │   ├── line/              # Line Login SDK
│       │   └── geo/               # 地理計算工具
│       ├── migrations/            # 資料庫遷移
│       │   ├── 000001_init_schema.up.sql
│       │   └── 000001_init_schema.down.sql
│       ├── api/                   # OpenAPI 規格
│       │   └── openapi.yaml
│       ├── Dockerfile
│       ├── go.mod
│       ├── go.sum
│       └── Makefile
│
├── packages/                      # 共用套件
│   └── shared-types/              # 共用型別 (如果需要)
│       ├── src/
│       │   └── index.ts
│       ├── tsconfig.json
│       └── package.json
│
├── docker/
│   ├── docker-compose.yml         # 本地開發環境
│   └── docker-compose.prod.yml
│
├── docs/
│   ├── api/                       # API 文件
│   └── architecture/              # 架構文件
│
├── scripts/
│   ├── setup.sh                   # 環境設置腳本
│   └── seed.sh                    # 種子資料腳本
│
├── .env.example
├── .gitignore
├── pnpm-workspace.yaml
├── package.json
├── PRD-V1.md
├── PRODUCT.md
├── MULTI_AGENT_PLAN.md
└── README.md
```

### 3.3 技術棧總覽

| 層級 | 技術選擇 | 版本 |
|------|----------|------|
| **前端框架** | Next.js (App Router) | 14.x |
| **前端 UI** | Tailwind CSS + shadcn/ui | 3.x |
| **狀態管理** | TanStack Query (React Query) | 5.x |
| **地圖** | Google Maps JavaScript API | latest |
| **後端框架** | Go + Gin | Go 1.21+, Gin 1.9+ |
| **ORM** | sqlx (非 ORM，輕量查詢) | 1.3+ |
| **資料庫** | PostgreSQL + PostGIS | 15+ |
| **快取** | Redis | 7.x |
| **認證** | Line Login + JWT | - |
| **部署** | GCP Cloud Run | - |
| **CDN** | Cloudflare | - |
| **監控** | Sentry + Cloud Monitoring | - |

---

## 4. 資料庫 Schema 設計

### 4.1 ERD (Entity-Relationship Diagram)

```sql
-- 啟用 PostGIS 擴展
CREATE EXTENSION IF NOT EXISTS postgis;

-- ============================================
-- 用戶表 (users)
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    line_user_id    VARCHAR(64) UNIQUE NOT NULL,      -- Line 用戶 ID
    display_name    VARCHAR(100) NOT NULL,             -- 顯示名稱
    avatar_url      TEXT,                              -- 大頭貼 URL
    email           VARCHAR(255),                      -- Email (可選)
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_line_user_id ON users(line_user_id);

-- ============================================
-- 活動表 (events)
-- ============================================
CREATE TABLE events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id         UUID NOT NULL REFERENCES users(id),

    -- 基本資訊
    title           VARCHAR(200),                      -- 活動標題 (可選)
    description     TEXT,                              -- 備註說明

    -- 時間
    event_date      DATE NOT NULL,                     -- 活動日期
    start_time      TIME NOT NULL,                     -- 開始時間
    end_time        TIME,                              -- 結束時間 (可選)

    -- 地點
    location_name   VARCHAR(200) NOT NULL,             -- 地點名稱
    location_address VARCHAR(500),                     -- 完整地址
    location_point  GEOGRAPHY(POINT, 4326) NOT NULL,   -- PostGIS 地理座標
    google_place_id VARCHAR(255),                      -- Google Place ID

    -- 活動設定
    capacity        SMALLINT NOT NULL CHECK (capacity >= 4 AND capacity <= 20),
    skill_level     VARCHAR(20) NOT NULL,              -- 程度: beginner, intermediate, advanced, expert, any
    fee             INTEGER DEFAULT 0 CHECK (fee >= 0 AND fee <= 9999),  -- 費用 (元)

    -- 狀態
    status          VARCHAR(20) DEFAULT 'open',        -- open, full, cancelled, completed

    -- 時間戳
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 地理空間索引
CREATE INDEX idx_events_location ON events USING GIST(location_point);
CREATE INDEX idx_events_event_date ON events(event_date);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_skill_level ON events(skill_level);
CREATE INDEX idx_events_host_id ON events(host_id);

-- ============================================
-- 報名表 (registrations)
-- ============================================
CREATE TABLE registrations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id            UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id),

    -- 報名狀態
    status              VARCHAR(20) NOT NULL DEFAULT 'confirmed',  -- confirmed, waitlist, cancelled
    waitlist_position   SMALLINT,                      -- 候補順序 (僅 waitlist 狀態)

    -- 時間戳
    registered_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_at        TIMESTAMP WITH TIME ZONE,      -- 正取確認時間
    cancelled_at        TIMESTAMP WITH TIME ZONE,

    -- 唯一約束：同一用戶不能重複報名同一活動
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_registrations_event_id ON registrations(event_id);
CREATE INDEX idx_registrations_user_id ON registrations(user_id);
CREATE INDEX idx_registrations_status ON registrations(status);

-- ============================================
-- 通知表 (notifications) - Phase 1 簡易版
-- ============================================
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    event_id        UUID REFERENCES events(id),

    type            VARCHAR(50) NOT NULL,              -- waitlist_promoted, event_cancelled, etc.
    title           VARCHAR(200) NOT NULL,
    message         TEXT,

    is_read         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);

-- ============================================
-- 輔助視圖 (Views)
-- ============================================

-- 活動摘要視圖 (包含報名人數)
CREATE VIEW event_summary AS
SELECT
    e.*,
    COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END) AS confirmed_count,
    COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END) AS waitlist_count,
    u.display_name AS host_name,
    u.avatar_url AS host_avatar
FROM events e
LEFT JOIN registrations r ON e.id = r.event_id
LEFT JOIN users u ON e.host_id = u.id
GROUP BY e.id, u.display_name, u.avatar_url;
```

### 4.2 程度標籤對應

```go
// 程度定義
const (
    SkillBeginner     = "beginner"      // 新手友善 (2.0-2.5)
    SkillIntermediate = "intermediate"  // 中階 (2.5-3.5)
    SkillAdvanced     = "advanced"      // 進階 (3.5-4.5)
    SkillExpert       = "expert"        // 高階 (4.5+)
    SkillAny          = "any"           // 不限程度
)
```

---

## 5. API 設計規範

### 5.1 API 端點總覽

| 方法 | 端點 | 說明 | 認證 |
|------|------|------|------|
| **認證** |
| POST | `/api/v1/auth/line/callback` | Line Login 回調 | No |
| POST | `/api/v1/auth/refresh` | 刷新 Token | Yes |
| POST | `/api/v1/auth/logout` | 登出 | Yes |
| **用戶** |
| GET | `/api/v1/users/me` | 取得當前用戶資訊 | Yes |
| GET | `/api/v1/users/me/events` | 取得我的活動列表 | Yes |
| GET | `/api/v1/users/me/registrations` | 取得我的報名列表 | Yes |
| GET | `/api/v1/users/me/notifications` | 取得我的通知 | Yes |
| **活動** |
| GET | `/api/v1/events` | 查詢活動列表 (含地理篩選) | No |
| GET | `/api/v1/events/:id` | 取得活動詳情 | No |
| POST | `/api/v1/events` | 建立活動 | Yes |
| PUT | `/api/v1/events/:id` | 更新活動 | Yes (Host) |
| DELETE | `/api/v1/events/:id` | 取消活動 | Yes (Host) |
| **報名** |
| POST | `/api/v1/events/:id/register` | 報名活動 | Yes |
| DELETE | `/api/v1/events/:id/register` | 取消報名 | Yes |
| GET | `/api/v1/events/:id/registrations` | 取得活動報名列表 | No |

### 5.2 API 請求/回應範例

#### 5.2.1 查詢活動 (地圖模式)

```http
GET /api/v1/events?lat=25.0330&lng=121.5654&radius=10000&skill_level=beginner&status=open
```

**Response:**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "host": {
          "id": "...",
          "display_name": "Wayne",
          "avatar_url": "https://..."
        },
        "event_date": "2026-01-25",
        "start_time": "20:00",
        "location": {
          "name": "內湖運動中心",
          "address": "台北市內湖區...",
          "lat": 25.0830,
          "lng": 121.5890
        },
        "capacity": 4,
        "confirmed_count": 2,
        "waitlist_count": 0,
        "skill_level": "beginner",
        "skill_level_label": "新手友善 (2.0-2.5)",
        "fee": 200,
        "status": "open"
      }
    ],
    "total": 15,
    "has_more": true
  }
}
```

#### 5.2.2 建立活動

```http
POST /api/v1/events
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "event_date": "2026-01-25",
  "start_time": "20:00",
  "end_time": "22:00",
  "location": {
    "name": "內湖運動中心",
    "address": "台北市內湖區...",
    "lat": 25.0830,
    "lng": 121.5890,
    "google_place_id": "ChIJ..."
  },
  "capacity": 4,
  "skill_level": "beginner",
  "fee": 200,
  "description": "歡迎新手加入!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "share_url": "https://picklego.tw/g/abc123"
  }
}
```

#### 5.2.3 報名活動

```http
POST /api/v1/events/550e8400-e29b-41d4-a716-446655440000/register
Authorization: Bearer <jwt_token>
```

**Response (正取):**
```json
{
  "success": true,
  "data": {
    "status": "confirmed",
    "message": "報名成功！"
  }
}
```

**Response (候補):**
```json
{
  "success": true,
  "data": {
    "status": "waitlist",
    "waitlist_position": 2,
    "message": "已加入候補（第 2 位）"
  }
}
```

### 5.3 錯誤回應格式

```json
{
  "success": false,
  "error": {
    "code": "EVENT_FULL",
    "message": "活動已滿，無法報名",
    "details": {}
  }
}
```

---

## 6. 第三方服務整合

### 6.1 Line Login

```go
// Line Login 流程
// 1. 前端導向 Line 授權頁面
// 2. 用戶授權後，Line 回調到前端
// 3. 前端將 authorization_code 傳給後端
// 4. 後端向 Line 換取 access_token 及 ID token
// 5. 後端驗證 ID token 並取得用戶資訊
// 6. 後端建立/更新用戶並發放 JWT

// Line Login Configuration
type LineConfig struct {
    ChannelID     string
    ChannelSecret string
    RedirectURI   string
}
```

**前端整合:**
```typescript
// apps/web/src/lib/auth.ts
const LINE_AUTH_URL = 'https://access.line.me/oauth2/v2.1/authorize';

export function getLineLoginURL(state: string): string {
  const params = new URLSearchParams({
    response_type: 'code',
    client_id: process.env.NEXT_PUBLIC_LINE_CHANNEL_ID!,
    redirect_uri: process.env.NEXT_PUBLIC_LINE_REDIRECT_URI!,
    state,
    scope: 'profile openid',
  });
  return `${LINE_AUTH_URL}?${params}`;
}
```

### 6.2 Google Maps JavaScript API

```typescript
// apps/web/src/components/map/EventMap.tsx
import { GoogleMap, useLoadScript, Marker } from '@react-google-maps/api';

const libraries: Libraries = ['places'];

export function EventMap({ events, center, onMarkerClick }) {
  const { isLoaded } = useLoadScript({
    googleMapsApiKey: process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY!,
    libraries,
  });

  if (!isLoaded) return <MapSkeleton />;

  return (
    <GoogleMap
      center={center}
      zoom={13}
      mapContainerClassName="w-full h-full"
      options={{
        disableDefaultUI: true,
        zoomControl: true,
      }}
    >
      {events.map((event) => (
        <EventPin
          key={event.id}
          event={event}
          onClick={() => onMarkerClick(event)}
        />
      ))}
    </GoogleMap>
  );
}
```

### 6.3 Google Places API

```typescript
// 地點自動完成
import usePlacesAutocomplete, {
  getGeocode,
  getLatLng,
} from 'use-places-autocomplete';

export function LocationInput({ onSelect }) {
  const {
    ready,
    value,
    suggestions: { status, data },
    setValue,
    clearSuggestions,
  } = usePlacesAutocomplete({
    requestOptions: {
      componentRestrictions: { country: 'tw' },  // 限制台灣
    },
  });

  const handleSelect = async (address: string) => {
    setValue(address, false);
    clearSuggestions();

    const results = await getGeocode({ address });
    const { lat, lng } = await getLatLng(results[0]);

    onSelect({
      name: address,
      address: results[0].formatted_address,
      lat,
      lng,
      placeId: results[0].place_id,
    });
  };

  // ... render UI
}
```

---

## 7. 里程碑與任務分解

### 7.1 M0: 專案啟動 (W0)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| M0-T01 | 確認 PRD 並完成技術選型文件 | P0 | - | S | Tech Lead |
| M0-T02 | 申請 Line Login Channel (需 3-5 工作天) | P0 | - | S | PM/DevOps |
| M0-T03 | 申請 Google Cloud Platform 帳號並啟用 API | P0 | - | S | DevOps |
| M0-T04 | 建立 Monorepo 專案結構 (pnpm workspace) | P0 | - | M | Frontend |
| M0-T05 | 建立 Go 後端專案骨架 (Gin + 目錄結構) | P0 | - | M | Backend |
| M0-T06 | 設置本地開發環境 (Docker Compose) | P0 | M0-T04, M0-T05 | M | DevOps |
| M0-T07 | 設置 CI/CD Pipeline (GitHub Actions) | P1 | M0-T04, M0-T05 | M | DevOps |
| M0-T08 | 建立 PostgreSQL + PostGIS 本地與雲端環境 | P0 | M0-T03 | M | DevOps |

**M0 驗收標準：**
- 團隊可在本地執行前後端開發環境
- CI Pipeline 可執行 lint 與 test
- Line Login Channel 申請已送出

---

### 7.2 M1: 核心骨架 (W1-W2)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **資料庫** |
| M1-T01 | 建立資料庫遷移工具與初始 Schema | P0 | M0-T08 | M | Backend |
| M1-T02 | 實作 User Model 與 Repository | P0 | M1-T01 | S | Backend |
| M1-T03 | 實作 Event Model 與 Repository | P0 | M1-T01 | M | Backend |
| M1-T04 | 實作 Registration Model 與 Repository | P0 | M1-T01 | M | Backend |
| **後端 API** |
| M1-T05 | 實作 JWT 認證中介軟體 | P0 | - | M | Backend |
| M1-T06 | 實作 Line Login 後端流程 (OAuth callback) | P0 | M0-T02, M1-T02 | L | Backend |
| M1-T07 | 實作認證 API (login, logout, refresh) | P0 | M1-T05, M1-T06 | M | Backend |
| M1-T08 | 撰寫 OpenAPI 規格文件 | P1 | M1-T07 | M | Backend |
| **前端基礎** |
| M1-T09 | 設置 Next.js 14 + Tailwind + shadcn/ui | P0 | M0-T04 | M | Frontend |
| M1-T10 | 實作基礎 Layout 與導航元件 | P0 | M1-T09 | S | Frontend |
| M1-T11 | 實作 Line Login 前端流程 | P0 | M1-T07 | M | Frontend |
| M1-T12 | 設置 TanStack Query 與 API Client | P0 | M1-T07 | M | Frontend |
| M1-T13 | 實作認證狀態管理 (Context/Hook) | P0 | M1-T11, M1-T12 | M | Frontend |

**M1 驗收標準：**
- 用戶可完成 Line Login 登入/登出
- JWT Token 正確發放與驗證
- 前端可顯示登入狀態與用戶資訊

---

### 7.3 M2: 團主流程 (W3-W4)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **後端 API** |
| M2-T01 | 實作建立活動 API (POST /events) | P0 | M1-T03 | M | Backend |
| M2-T02 | 實作活動詳情 API (GET /events/:id) | P0 | M1-T03 | S | Backend |
| M2-T03 | 實作更新/取消活動 API | P0 | M2-T01 | M | Backend |
| M2-T04 | 實作活動短網址生成邏輯 | P1 | M2-T01 | S | Backend |
| **前端頁面** |
| M2-T05 | 整合 Google Places API 地點自動完成 | P0 | M1-T09 | M | Frontend |
| M2-T06 | 實作建立活動表單頁面 | P0 | M2-T01, M2-T05 | L | Frontend |
| M2-T07 | 實作活動詳情頁面 | P0 | M2-T02 | M | Frontend |
| M2-T08 | 實作表單驗證 (日期、人數、費用) | P0 | M2-T06 | M | Frontend |
| M2-T09 | 實作活動建立成功頁面 (含複製連結) | P0 | M2-T06 | S | Frontend |
| **SEO 與分享** |
| M2-T10 | 實作活動頁 OG Tags (動態生成) | P0 | M2-T07 | M | Frontend |
| M2-T11 | 驗證 Line 分享預覽卡片顯示 | P0 | M2-T10 | S | QA |

**M2 驗收標準：**
- 團主可建立活動並取得獨立網址
- 活動詳情頁可正確顯示
- 分享到 Line 群組有正確的預覽卡片

---

### 7.4 M3: 球友流程 (W5-W6)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **後端 API** |
| M3-T01 | 實作地理查詢 API (GET /events with geo filter) | P0 | M1-T03 | L | Backend |
| M3-T02 | 實作程度篩選邏輯 | P0 | M3-T01 | S | Backend |
| M3-T03 | 實作報名 API (POST /events/:id/register) | P0 | M1-T04 | M | Backend |
| M3-T04 | 實作取消報名 API | P0 | M3-T03 | M | Backend |
| M3-T05 | 實作候補邏輯 (滿團自動轉候補) | P0 | M3-T03 | L | Backend |
| M3-T06 | 實作候補遞補邏輯 (正取取消時) | P0 | M3-T04, M3-T05 | L | Backend |
| M3-T07 | 實作簡易通知記錄 | P1 | M3-T06 | M | Backend |
| **前端頁面** |
| M3-T08 | 整合 Google Maps JavaScript API | P0 | M1-T09 | M | Frontend |
| M3-T09 | 實作地圖模式首頁 | P0 | M3-T01, M3-T08 | L | Frontend |
| M3-T10 | 實作 Event Pin 元件 (顏色邏輯) | P0 | M3-T09 | M | Frontend |
| M3-T11 | 實作活動摘要卡片 (點擊 Pin 顯示) | P0 | M3-T09 | M | Frontend |
| M3-T12 | 實作程度篩選器 UI | P0 | M3-T02, M3-T09 | M | Frontend |
| M3-T13 | 實作報名按鈕元件 (含狀態切換) | P0 | M3-T03 | M | Frontend |
| M3-T14 | 實作報名/候補成功回饋 UI | P0 | M3-T13 | S | Frontend |
| M3-T15 | 實作「我的報名」頁面 | P1 | M3-T03 | M | Frontend |
| M3-T16 | 取得瀏覽器地理位置 (Geolocation API) | P0 | M3-T09 | S | Frontend |

**M3 驗收標準：**
- 用戶可在地圖上瀏覽附近活動
- 可依程度篩選活動
- 可完成報名/取消報名流程
- 滿團後可排候補，正取取消時候補自動遞補

---

### 7.5 M4: SEO 與優化 (W7-W8)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **SEO** |
| M4-T01 | 確保活動頁 SSR 正確渲染 | P0 | M2-T07 | M | Frontend |
| M4-T02 | 實作 Schema.org Event 結構化資料 | P0 | M4-T01 | M | Frontend |
| M4-T03 | 實作動態 Sitemap.xml | P1 | M4-T01 | M | Frontend |
| M4-T04 | 設置 robots.txt | P1 | - | S | Frontend |
| M4-T05 | 優化 Meta Tags (Title, Description) | P0 | M4-T01 | S | Frontend |
| **效能優化** |
| M4-T06 | 實作地圖 Pin 資料快取 (React Query) | P1 | M3-T09 | M | Frontend |
| M4-T07 | 實作圖片優化 (Next.js Image) | P1 | - | S | Frontend |
| M4-T08 | 實作 API 回應快取 (Redis) | P2 | M3-T01 | M | Backend |
| M4-T09 | 執行 Lighthouse 效能測試 | P0 | M4-T01 ~ M4-T07 | S | QA |
| M4-T10 | 修復 Lighthouse 標示的問題 | P0 | M4-T09 | M | All |
| **安全性** |
| M4-T11 | 設置 CORS 策略 | P0 | - | S | Backend |
| M4-T12 | 設置 Rate Limiting | P1 | - | M | Backend |
| M4-T13 | 設置 HTTPS (Cloudflare) | P0 | - | S | DevOps |

**M4 驗收標準：**
- Lighthouse Performance Score > 80
- 活動頁面可被搜尋引擎正確索引
- API 回應時間 P95 < 500ms

---

### 7.6 M5: Closed Beta (W9-W10)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **部署準備** |
| M5-T01 | 設置 GCP Cloud Run 環境 | P0 | - | M | DevOps |
| M5-T02 | 設置 Cloud SQL (PostgreSQL) 生產環境 | P0 | - | M | DevOps |
| M5-T03 | 設置環境變數與 Secrets 管理 | P0 | M5-T01 | S | DevOps |
| M5-T04 | 部署前端到 Cloud Run | P0 | M5-T01 | M | DevOps |
| M5-T05 | 部署後端到 Cloud Run | P0 | M5-T01 | M | DevOps |
| M5-T06 | 設置自訂網域 (picklego.tw) | P0 | M5-T04 | M | DevOps |
| **測試** |
| M5-T07 | 執行端對端測試 (E2E) | P0 | M5-T04, M5-T05 | L | QA |
| M5-T08 | 建立種子資料 (測試活動) | P0 | M5-T05 | S | Dev |
| M5-T09 | 邀請 10 位內部測試用戶 | P0 | M5-T08 | S | PM |
| M5-T10 | 收集並分類 Bug 回報 | P0 | M5-T09 | M | PM |
| M5-T11 | 修復 P0/P1 Bug | P0 | M5-T10 | L | All |
| **監控** |
| M5-T12 | 設置 Sentry 錯誤監控 | P0 | M5-T04, M5-T05 | M | DevOps |
| M5-T13 | 設置 Uptime 監控 | P1 | M5-T06 | S | DevOps |
| M5-T14 | 設置 Google Analytics 4 | P1 | M5-T04 | S | Frontend |

**M5 驗收標準：**
- 系統部署到生產環境並可存取
- 10 位測試用戶完成至少一次完整流程
- 無 P0/P1 等級 Bug

---

### 7.7 M6: Public Launch (W11-W12)

| 任務 ID | 任務描述 | 優先級 | 依賴 | 複雜度 | 負責角色 |
|---------|----------|--------|------|--------|----------|
| **最終準備** |
| M6-T01 | 完成所有 P1 Bug 修復 | P0 | M5-T11 | M | All |
| M6-T02 | 執行最終回歸測試 | P0 | M6-T01 | M | QA |
| M6-T03 | 準備種子團主計畫 (5 位) | P0 | - | S | PM |
| M6-T04 | 建立初始活動內容 | P0 | M6-T03 | S | PM |
| **上線** |
| M6-T05 | 執行 Go/No-Go 檢查清單 | P0 | M6-T01 ~ M6-T04 | S | Tech Lead |
| M6-T06 | 正式上線發布 | P0 | M6-T05 | S | DevOps |
| M6-T07 | 監控上線後系統狀態 | P0 | M6-T06 | M | DevOps |
| **觀察期** |
| M6-T08 | 追蹤首週 KPI (活動數、用戶數) | P1 | M6-T06 | S | PM |
| M6-T09 | 收集用戶回饋 | P1 | M6-T06 | S | PM |
| M6-T10 | 修復上線後發現的問題 | P0 | M6-T06 | M | All |

**M6 驗收標準 (Go/No-Go)：**
- [ ] 完整流程可用：建立活動 -> 分享 -> 報名 -> 滿團
- [ ] Line Login 正常運作
- [ ] 行動裝置體驗順暢
- [ ] 無 P0/P1 等級 Bug
- [ ] 效能指標達標 (LCP < 2.5s)
- [ ] 至少 5 位種子團主承諾開團
- [ ] 首週目標：10 場活動

---

## 8. 任務複雜度與工時估算

### 8.1 複雜度定義

| 複雜度 | 定義 | 預估工時 |
|--------|------|----------|
| S (Small) | 直觀實作，無技術挑戰 | 0.5-1 天 |
| M (Medium) | 需要一些研究或整合工作 | 1-2 天 |
| L (Large) | 複雜邏輯或多項整合 | 2-4 天 |

### 8.2 各里程碑工時估算

| 里程碑 | 任務數 | S | M | L | 預估工時 (人天) |
|--------|--------|---|---|---|-----------------|
| M0 | 8 | 3 | 5 | 0 | 8-10 |
| M1 | 13 | 2 | 10 | 1 | 14-18 |
| M2 | 11 | 3 | 6 | 2 | 12-16 |
| M3 | 16 | 3 | 9 | 4 | 18-24 |
| M4 | 13 | 5 | 7 | 1 | 10-14 |
| M5 | 14 | 4 | 7 | 3 | 14-18 |
| M6 | 10 | 5 | 4 | 1 | 8-12 |
| **總計** | 85 | 25 | 48 | 12 | **84-112** |

**備註：** 以 2 位全端工程師計算，12 週內可完成 (含緩衝)。

---

## 9. 風險與緩解措施

| 風險 | 可能性 | 影響 | 緩解措施 |
|------|--------|------|----------|
| Line Login Channel 審核延遲 | 中 | 高 | 1. 第一天即送出申請<br>2. 準備 Email/密碼登入作為備案 |
| Google Maps API 費用超支 | 低 | 中 | 1. 設定每月預算上限 ($200)<br>2. 實作客戶端快取減少請求<br>3. 考慮免費替代方案 (Leaflet + OSM) |
| PostGIS 地理查詢效能問題 | 低 | 中 | 1. 預先建立空間索引<br>2. 限制查詢範圍<br>3. 使用 Redis 快取熱門區域 |
| 冷啟動無活動內容 | 高 | 高 | 1. 種子用戶計畫 (5 位團主)<br>2. 團隊自建測試活動<br>3. 與既有匹克球社群合作 |
| SEO 效果緩慢 | 高 | 中 | 1. 同步經營社群媒體<br>2. 預留付費廣告預算<br>3. 優化 Core Web Vitals |
| 手機瀏覽器相容性問題 | 中 | 中 | 1. 使用 BrowserStack 測試<br>2. 優先支援 iOS Safari / Android Chrome |

---

## 10. 關鍵技術決策摘要

| 決策項目 | 選擇 | 主要理由 |
|----------|------|----------|
| 專案結構 | Monorepo (pnpm) | 小團隊協作效率、共用程式碼 |
| 後端語言 | Go + Gin | 高效能、單一部署產物、Cloud Run 友善 |
| 前端框架 | Next.js 14 App Router | SSR/SSG、SEO、React 生態系 |
| 資料庫 | PostgreSQL + PostGIS | 地理查詢、關聯資料、事務一致性 |
| 認證方式 | Line Login + JWT | 台灣用戶習慣、無狀態擴展性 |
| 雲端平台 | GCP Cloud Run | 按需計費、整合 Google Maps API |
| API 風格 | REST + OpenAPI | 簡單直觀、可自動生成客戶端 |

---

## 附錄 A：環境變數清單

```bash
# === 前端 (Next.js) ===
NEXT_PUBLIC_API_URL=https://api.picklego.tw
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIza...
NEXT_PUBLIC_LINE_CHANNEL_ID=12345678
NEXT_PUBLIC_LINE_REDIRECT_URI=https://picklego.tw/auth/callback

# === 後端 (Go) ===
PORT=8080
DATABASE_URL=postgres://user:pass@host:5432/picklego?sslmode=require
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=168h  # 7 days
LINE_CHANNEL_ID=12345678
LINE_CHANNEL_SECRET=abc123...
CORS_ALLOWED_ORIGINS=https://picklego.tw
```

---

## 附錄 B：開發環境設置指南

```bash
# 1. Clone Repository
git clone https://github.com/org/pickle-go.git
cd pickle-go

# 2. 安裝依賴
pnpm install

# 3. 啟動本地服務
docker compose up -d  # PostgreSQL + Redis

# 4. 執行資料庫遷移
cd apps/api && make migrate-up

# 5. 啟動開發伺服器
pnpm dev  # 同時啟動前後端
```

---

## 文件變更紀錄

| 版本 | 日期 | 變更內容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-01-20 | 初版建立 | Tech Architect |

---

*本文件為 Pickle Go Phase 1 MVP 的技術實作計劃，後續將依據開發進度與技術挑戰持續迭代。*
