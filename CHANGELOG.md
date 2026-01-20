# Changelog

本專案的所有重大變更將記錄於此文件。

格式基於 [Keep a Changelog](https://keepachangelog.com/zh-TW/1.0.0/)，
版本號遵循 [Semantic Versioning](https://semver.org/lang/zh-TW/)。

---

## [Unreleased]

### 待發布功能
- 候補遞補通知
- 活動評價系統
- 用戶程度認證

---

## [1.0.0] - 2026-01-21

### 新增功能 (Added)

#### 認證系統
- LINE Login OAuth 2.0 整合
- JWT Token 認證機制
- 自動登入狀態保持 (7 天)

#### 團主功能 (建立活動)
- 建立匹克球揪團活動
- 地點搜尋整合 Google Places API
- 活動短網址自動生成
- 分享連結至 LINE 群組
- 活動管理 (編輯、取消)

#### 球友功能 (瀏覽與報名)
- 地圖模式瀏覽附近活動
- 依程度篩選活動 (新手/中階/進階/高階)
- 一鍵報名功能
- 候補機制 (滿團自動轉候補)
- 候補遞補 (正取取消時自動遞補)
- 取消報名

#### 頁面與 UI
- 響應式設計 (Mobile-first)
- 活動詳情頁面
- 我的報名列表
- 我的活動列表

#### SEO 與分享
- 活動頁面 SSR 支援
- Open Graph meta tags
- LINE 分享預覽卡片
- Schema.org 結構化資料

#### 基礎設施
- GCP Cloud Run 部署
- Cloud SQL (PostgreSQL + PostGIS)
- GitHub Actions CI/CD
- Sentry 錯誤監控

### 技術規格

#### 前端
- Next.js 14 (App Router)
- React 18
- TanStack Query 5
- Tailwind CSS 3
- shadcn/ui
- Playwright E2E 測試

#### 後端
- Go 1.21+
- Gin Web Framework
- PostgreSQL 15 + PostGIS
- JWT 認證

#### 部署
- GCP Cloud Run (asia-east1)
- Cloud SQL
- Cloudflare (CDN)

---

## 版本歷史

### 里程碑

| 版本 | 日期 | 說明 |
|------|------|------|
| v1.0.0 | 2026-01-21 | 正式上線 (Public Launch) |
| v0.9.0 | 2026-01-14 | Closed Beta |
| v0.5.0 | 2026-01-07 | 核心功能完成 |
| v0.1.0 | 2025-12-20 | 專案初始化 |

---

## 貢獻者

感謝所有參與開發的貢獻者。

---

[Unreleased]: https://github.com/YOUR_ORG/pickle-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/YOUR_ORG/pickle-go/releases/tag/v1.0.0
