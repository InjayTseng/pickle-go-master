# Pickle Go

> Web-First 匹克球揪團平台 - 讓用戶在 30 秒內找到附近球局並報名

[![CI](https://github.com/YOUR_ORG/pickle-go/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_ORG/pickle-go/actions/workflows/ci.yml)
[![Deploy](https://github.com/YOUR_ORG/pickle-go/actions/workflows/deploy-production.yml/badge.svg)](https://github.com/YOUR_ORG/pickle-go/actions/workflows/deploy-production.yml)

---

## 專案簡介

Pickle Go 是一個專為台灣匹克球愛好者打造的揪團平台。團主可以快速建立活動並分享至 LINE 群組，球友則可以在地圖上瀏覽附近球局並一鍵報名。

### 主要功能

- **地圖瀏覽** - 在地圖上直覺地瀏覽附近的匹克球活動
- **LINE 登入** - 使用 LINE 帳號一鍵登入，無需額外註冊
- **快速揪團** - 30 秒內建立活動並取得分享連結
- **智能候補** - 活動滿團自動轉為候補，正取取消時自動遞補
- **程度篩選** - 依照自己的程度找到適合的球局

---

## 技術架構

```
├── apps/
│   ├── web/          # Next.js 14 前端 (App Router)
│   └── api/          # Go + Gin 後端 API
├── packages/         # 共用套件
├── docs/             # 專案文件
└── docker/           # Docker 配置
```

### 技術棧

| 類別 | 技術 |
|------|------|
| 前端框架 | Next.js 14 (App Router) |
| UI 元件 | Tailwind CSS + shadcn/ui |
| 狀態管理 | TanStack Query 5 |
| 地圖服務 | Google Maps JavaScript API |
| 後端框架 | Go 1.21+ + Gin |
| 資料庫 | PostgreSQL 15 + PostGIS |
| 認證 | LINE Login + JWT |
| 部署 | GCP Cloud Run |
| 監控 | Sentry |

---

## 快速開始

### 系統需求

- Node.js >= 20.0.0
- pnpm >= 9.0.0
- Go >= 1.21
- Docker & Docker Compose
- PostgreSQL 15+ (含 PostGIS)

### 環境設置

1. **Clone 專案**

   ```bash
   git clone https://github.com/YOUR_ORG/pickle-go.git
   cd pickle-go
   ```

2. **安裝依賴**

   ```bash
   pnpm install
   ```

3. **設置環境變數**

   ```bash
   cp .env.example .env
   # 編輯 .env 填入必要的環境變數
   ```

4. **啟動本地服務**

   ```bash
   # 啟動資料庫
   cd docker/
   docker compose up -d

   # 執行資料庫遷移
   cd apps/api && make migrate-up && cd ../..

   # 啟動開發伺服器
   pnpm dev
   ```

5. **存取應用程式**
   - 前端: http://localhost:3000
   - API: http://localhost:8080
   - API 文件: http://localhost:8080/swagger

---

## 開發指令

### 根目錄指令

```bash
# 啟動所有服務 (開發模式)
pnpm dev

# 只啟動前端
pnpm dev:web

# 只啟動後端
pnpm dev:api

# 建置所有專案
pnpm build

# 執行所有測試
pnpm test

# 執行 lint 檢查
pnpm lint

# 執行 TypeScript 類型檢查
pnpm typecheck
```

### 前端指令 (apps/web)

```bash
cd apps/web

# 開發伺服器
pnpm dev

# 建置
pnpm build

# 執行 E2E 測試
pnpm test:e2e

# E2E 測試 (UI 模式)
pnpm test:e2e:ui
```

### 後端指令 (apps/api)

```bash
cd apps/api

# 開發伺服器 (需要 air)
make dev

# 建置
make build

# 執行測試
make test

# 執行測試並生成覆蓋率報告
make test-coverage

# 程式碼檢查
make lint

# 資料庫遷移
make migrate-up
make migrate-down
```

---

## 環境變數

### 前端 (Next.js)

| 變數 | 說明 | 範例 |
|------|------|------|
| `NEXT_PUBLIC_API_URL` | API 伺服器 URL | `http://localhost:8080` |
| `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` | Google Maps API Key | `AIza...` |
| `NEXT_PUBLIC_LINE_CHANNEL_ID` | LINE Login Channel ID | `12345678` |
| `NEXT_PUBLIC_LINE_REDIRECT_URI` | LINE OAuth 回調 URL | `http://localhost:3000/auth/callback` |
| `NEXT_PUBLIC_SENTRY_DSN` | Sentry DSN (前端) | `https://...@sentry.io/...` |

### 後端 (Go)

| 變數 | 說明 | 範例 |
|------|------|------|
| `PORT` | 伺服器埠號 | `8080` |
| `DATABASE_URL` | PostgreSQL 連線字串 | `postgres://user:pass@localhost:5432/picklego` |
| `JWT_SECRET` | JWT 簽名密鑰 | `your-secret-key` |
| `JWT_EXPIRY` | JWT 過期時間 | `168h` |
| `LINE_CHANNEL_ID` | LINE Channel ID | `12345678` |
| `LINE_CHANNEL_SECRET` | LINE Channel Secret | `abc123...` |
| `CORS_ALLOWED_ORIGINS` | CORS 允許來源 | `http://localhost:3000` |
| `SENTRY_DSN` | Sentry DSN (後端) | `https://...@sentry.io/...` |

---

## 專案結構

### 前端 (apps/web)

```
apps/web/
├── src/
│   ├── app/                 # App Router 頁面
│   │   ├── (auth)/          # 認證相關頁面
│   │   ├── (main)/          # 主要頁面
│   │   │   ├── events/      # 活動頁面
│   │   │   └── page.tsx     # 首頁 (地圖)
│   │   └── layout.tsx       # 根 Layout
│   ├── components/          # React 元件
│   │   ├── ui/              # shadcn/ui 元件
│   │   ├── map/             # 地圖相關元件
│   │   └── event/           # 活動相關元件
│   ├── hooks/               # Custom Hooks
│   ├── lib/                 # 工具函式
│   └── contexts/            # React Contexts
├── e2e/                     # E2E 測試
└── public/                  # 靜態資源
```

### 後端 (apps/api)

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go          # 程式進入點
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP Handler
│   ├── middleware/          # 中介軟體
│   ├── model/               # 資料模型
│   ├── repository/          # 資料存取層
│   ├── service/             # 業務邏輯層
│   └── dto/                 # Data Transfer Objects
├── pkg/                     # 可重用套件
│   ├── jwt/                 # JWT 工具
│   ├── line/                # LINE Login SDK
│   └── geo/                 # 地理計算工具
├── migrations/              # 資料庫遷移
└── api/                     # OpenAPI 規格
```

---

## 部署

### 自動部署

推送標籤到 main 分支會自動觸發 CI/CD 流程：

```bash
# 建立版本標籤
git tag v1.0.0
git push origin v1.0.0
```

### 手動部署

透過 GitHub Actions 手動觸發：
1. 前往 Actions 頁面
2. 選擇 "Deploy Production" workflow
3. 點擊 "Run workflow"

詳細部署文件請參考 [docs/launch-checklist.md](docs/launch-checklist.md)

---

## 文件

| 文件 | 說明 |
|------|------|
| [CHANGELOG.md](CHANGELOG.md) | 版本變更紀錄 |
| [planning/MULTI_AGENT_PLAN-V0.md](planning/MULTI_AGENT_PLAN-V0.md) | 技術實作計劃 |
| [planning/PRD-V0.md](planning/PRD-V0.md) | 產品需求文件 |
| [docs/launch-checklist.md](docs/launch-checklist.md) | 上線檢查清單 |
| [docs/monitoring-runbook.md](docs/monitoring-runbook.md) | 監控運維手冊 |
| [docs/regression-test-checklist.md](docs/regression-test-checklist.md) | 回歸測試清單 |

---

## 貢獻指南

1. Fork 專案
2. 建立功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交變更 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

### Commit 訊息規範

```
<type>(<scope>): <subject>

<body>

<footer>
```

類型 (type):
- `feat`: 新功能
- `fix`: Bug 修復
- `docs`: 文件更新
- `style`: 程式碼格式調整
- `refactor`: 重構
- `test`: 測試相關
- `chore`: 雜項

---

## 授權

本專案採用 MIT 授權 - 詳見 [LICENSE](LICENSE) 檔案

---

## 聯絡方式

- 專案網站: https://picklego.tw
- Email: team@picklego.tw

---

*Built with love for the pickleball community in Taiwan*
