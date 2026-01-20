# Pickle Go - 第一階段 MVP 產品需求文件 (PRD-V1)

**文件版本：** 1.0
**建立日期：** 2026-01-20
**文件負責人：** Product Team
**狀態：** Draft

---

## 1. 產品概述 (Product Overview)

### 1.1 產品願景

Pickle Go 是一個 **Web-First** 的匹克球揪團平台，致力於解決台灣匹克球社群在「揪團」與「找團」過程中的核心痛點。我們的目標是讓任何人都能在 **30 秒內** 完成「找到附近球局並報名」的完整流程。

### 1.2 核心問題陳述

| 問題 | 現況痛點 | Pickle Go 解法 |
|------|----------|----------------|
| 資訊碎片化 | 球局資訊散落在多個 Line 群組，難以追蹤 | 統一的地圖式瀏覽介面 |
| 報名摩擦高 | 需要加入群組、手動喊 +1、等待回覆 | 一鍵 Line 登入報名 |
| 程度不透明 | 無法預先知道參與者程度，造成期望落差 | 明確的程度標示與篩選 |
| 統計繁瑣 | 團主需手動統計報名、處理候補 | 自動化名單管理與候補遞補 |

### 1.3 產品定位

- **不是**：社群媒體、聊天工具、場地預約系統
- **是**：專注於「活動配對」的輕量工具，作為「公海流量」與「私域群組」之間的橋樑

---

## 2. 目標用戶 (Target Users)

### 2.1 主要用戶類型

#### 用戶 A：團主 (Host)

| 屬性 | 描述 |
|------|------|
| 典型人物 | Wayne，35 歲，球齡半年，週末固定租場地揪團 |
| 核心需求 | 快速湊滿人數、減少行政工作 |
| 痛點 | 在 Line 群組喊人效率低、手動統計耗時 |
| 成功定義 | 從建立活動到滿團的時間縮短 50% |

#### 用戶 B：球友 (Seeker)

| 屬性 | 描述 |
|------|------|
| 典型人物 | Alice，26 歲，零經驗新手，想嘗試匹克球 |
| 核心需求 | 找到程度相符的球局、低門檻參與 |
| 痛點 | 不知道去哪找、怕程度不符被打爆 |
| 成功定義 | 30 秒內完成找局到報名 |

### 2.2 MVP 階段用戶優先級

1. **P0 - 團主 (Host)**：平台冷啟動關鍵，需先有活動才有內容
2. **P1 - 球友 (Seeker)**：透過搜尋/分享連結進入的新用戶

---

## 3. MVP 範圍定義 (MVP Scope)

### 3.1 納入範圍 (In Scope)

| 模組 | 功能項 | 優先級 |
|------|--------|--------|
| 模組 A：探索與報名 | 地圖模式瀏覽、程度篩選、Line 登入報名 | P0 |
| 模組 B：開團與管理 | 快速建立活動、Line 分享預覽、自動候補 | P0 |
| 模組 C：擴散與留存 | SEO 優化活動頁 | P1 |

### 3.2 排除範圍 (Out of Scope - Phase 2+)

- 原生 App (iOS/Android)
- 付款金流整合
- 場地預約功能
- 私訊/聊天功能
- 用戶評價系統
- 數據分析 Dashboard
- 多語系支援

### 3.3 範圍決策原則

> "如果這個功能移除後，用戶仍能完成「建立活動 -> 分享 -> 報名 -> 滿團」的核心流程，則該功能為 P1 或更低優先級。"

---

## 4. 功能需求清單 (Functional Requirements)

### 4.1 模組 A：探索與報名 (Discovery & Join)

#### FR-A01：地圖模式瀏覽

**對應 User Story：** US-01

| 項目 | 規格 |
|------|------|
| **功能描述** | 用戶進入首頁即看到地圖，顯示附近的開團活動 |
| **前置條件** | 用戶允許瀏覽器取得地理位置（非必要，可預設台北市中心） |
| **業務規則** | - 預設顯示範圍：以用戶位置為中心，半徑 10 公里<br>- Pin 顏色邏輯：綠色 = 有缺額、紅色 = 已滿、灰色 = 已結束<br>- 預設僅顯示未來 7 天內的活動 |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 用戶瀏覽地圖上的活動
  Given 用戶開啟 Pickle Go 首頁
  When 頁面載入完成
  Then 顯示以用戶位置為中心的地圖
  And 地圖上顯示附近活動的 Pin
  And 綠色 Pin 代表有缺額的活動
  And 紅色 Pin 代表已滿的活動

Scenario: 用戶點擊地圖上的 Pin
  Given 用戶在地圖頁面
  When 用戶點擊任一 Pin
  Then 顯示活動摘要卡片，包含：時間、地點、費用、目前人數
  And 卡片上有「查看詳情」按鈕
```

**UI 要點：**
- 地圖使用 Google Maps JavaScript API
- Pin 需有 hover/active 狀態
- 摘要卡片需支援手機觸控操作

---

#### FR-A02：程度篩選器

**對應 User Story：** US-02

| 項目 | 規格 |
|------|------|
| **功能描述** | 用戶可依據程度篩選活動 |
| **程度定義** | - 新手友善 (2.0-2.5)<br>- 中階 (2.5-3.5)<br>- 進階 (3.5-4.5)<br>- 高階 (4.5+)<br>- 不限程度 |
| **預設值** | 「全部」（不篩選） |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 用戶篩選程度
  Given 用戶在地圖或列表頁面
  When 用戶點擊程度篩選器並選擇「新手友善」
  Then 僅顯示標示為「新手友善」的活動
  And 其他活動 Pin 從地圖上移除或變淡

Scenario: 活動卡片顯示程度標籤
  Given 用戶瀏覽任一活動卡片
  Then 卡片上明確顯示程度標籤（如「新手友善 2.0-2.5」）
```

---

#### FR-A03：Line 登入與報名

**對應 User Story：** US-03

| 項目 | 規格 |
|------|------|
| **功能描述** | 用戶透過 Line 登入後即可一鍵報名 |
| **登入方式** | Line Login v2.1 (OAuth 2.0) |
| **取得資料** | - Line User ID<br>- 顯示名稱<br>- 大頭貼 URL |
| **報名邏輯** | - 未滿：直接加入正取<br>- 已滿：加入候補 |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 未登入用戶報名
  Given 用戶未登入且瀏覽活動詳情頁
  When 用戶點擊「+1 參加」按鈕
  Then 跳出 Line Login 授權畫面
  When 用戶完成授權
  Then 自動完成報名
  And 顯示「報名成功」確認訊息
  And 頁面顯示「已報名」狀態

Scenario: 已登入用戶報名
  Given 用戶已登入
  When 用戶點擊「+1 參加」按鈕
  Then 直接完成報名（無需再次登入）
  And 更新活動人數顯示
```

**錯誤處理：**
- Line 授權失敗：顯示「登入失敗，請重試」
- 報名時活動已滿：自動轉入候補，顯示「已加入候補（第 N 位）」

---

### 4.2 模組 B：開團與管理 (Host & Manage)

#### FR-B01：建立活動

**對應 User Story：** US-04

| 項目 | 規格 |
|------|------|
| **功能描述** | 團主可快速建立活動頁面 |
| **必填欄位** | - 活動日期與時間<br>- 地點（支援 Google Places API）<br>- 人數上限（4-20 人）<br>- 程度要求 |
| **選填欄位** | - 費用（預設 0）<br>- 備註說明 |
| **產出** | 獨立活動網址 (如 `picklego.tw/g/{unique_id}`) |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 團主建立活動
  Given 團主已登入
  When 團主點擊「發起揪團」
  Then 顯示活動建立表單
  When 團主填寫必填欄位並送出
  Then 系統產生活動頁面
  And 顯示獨立活動網址
  And 提供「複製連結」按鈕

Scenario: 地點自動完成
  Given 團主在建立活動表單
  When 團主在地點欄位輸入「內湖運動」
  Then 顯示 Google Places 建議清單
  When 團主選擇「內湖運動中心」
  Then 自動帶入完整地址與座標
```

**欄位驗證規則：**
- 日期：不可選擇過去日期
- 時間：以 30 分鐘為單位
- 人數：4-20 人（含團主）
- 費用：0-9999 元

---

#### FR-B02：Line 分享預覽

**對應 User Story：** US-05

| 項目 | 規格 |
|------|------|
| **功能描述** | 活動連結分享至 Line 時顯示豐富的預覽卡片 |
| **Open Graph Tags** | - og:title: `{日期} {時間} @ {地點}`<br>- og:description: `{程度} | 還缺 {N} 人`<br>- og:image: 預設品牌圖或地點圖 |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 分享連結到 Line 群組
  Given 活動已建立且缺 3 人
  When 用戶複製連結並貼到 Line 聊天室
  Then Line 顯示預覽卡片
  And 卡片標題為「01/25 (六) 20:00 @ 內湖運動中心」
  And 卡片描述為「新手友善 | 還缺 3 人」
```

**技術要點：**
- OG Tags 需動態產生（依據活動資料）
- 需處理 Line 的 User-Agent 以正確回應爬蟲

---

#### FR-B03：自動候補遞補

**對應 User Story：** US-06

| 項目 | 規格 |
|------|------|
| **功能描述** | 正取取消時，自動將候補第一位轉正 |
| **通知方式** | Phase 1：網頁內通知（登入後可見）<br>Phase 2：Line 推播通知 |
| **候補邏輯** | FIFO (先進先出) |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 活動已滿，用戶排候補
  Given 活動人數已達上限
  When 新用戶點擊「+1 參加」
  Then 按鈕文字變為「排候補」
  When 用戶點擊「排候補」
  Then 加入候補名單
  And 顯示「已加入候補（第 N 位）」

Scenario: 正取取消，候補遞補
  Given 活動已滿且有候補名單
  When 正取用戶點擊「取消報名」
  Then 候補第一位自動轉為正取
  And 被遞補用戶下次登入時看到通知
  And 取消用戶看到確認訊息
```

**邊界情況處理：**
- 候補為空時：正取取消後活動變為「有缺額」
- 團主不可取消自己的報名

---

### 4.3 模組 C：擴散與留存 (Growth & Retention)

#### FR-C01：SEO 優化活動頁

**對應 User Story：** US-07

| 項目 | 規格 |
|------|------|
| **功能描述** | 活動頁面可被搜尋引擎索引，提升自然流量 |
| **渲染方式** | SSR (Server-Side Rendering) 或 SSG (Static Site Generation) |
| **結構化資料** | Schema.org Event 標記 |

**驗收標準 (Acceptance Criteria)：**

```gherkin
Scenario: 搜尋引擎爬取活動頁
  Given 活動頁面已建立
  When Google Bot 爬取頁面
  Then 回傳完整的 HTML 內容（非空白 JavaScript 頁面）
  And 包含正確的 meta tags
  And 包含 Schema.org Event 結構化資料

Scenario: 未登入用戶瀏覽活動
  Given 用戶未登入
  When 用戶透過 Google 搜尋進入活動頁
  Then 可看到完整的活動資訊
  And 報名按鈕可見（點擊後觸發登入流程）
```

**SEO 必備項目：**
- Title Tag：`{地點} 匹克球揪團 {日期} | Pickle Go`
- Meta Description：動態產生，包含時間、地點、程度
- Canonical URL
- Sitemap.xml

---

## 5. 非功能需求 (Non-Functional Requirements)

### 5.1 效能需求 (Performance)

| 指標 | 目標值 | 測量方式 |
|------|--------|----------|
| 首頁載入時間 (LCP) | < 2.5 秒 | Lighthouse |
| 首次互動時間 (FID) | < 100 毫秒 | Lighthouse |
| API 回應時間 (P95) | < 500 毫秒 | 後端監控 |
| 地圖 Pin 載入 | < 1 秒 (100 筆以內) | 實測 |

### 5.2 可用性需求 (Availability)

| 指標 | 目標值 |
|------|--------|
| 服務可用率 | 99.5% (月) |
| 計畫性維護 | 週間凌晨 2-4 點 |

### 5.3 安全性需求 (Security)

| 項目 | 要求 |
|------|------|
| HTTPS | 強制全站 HTTPS |
| 認證 | Line Login OAuth 2.0 |
| Session | JWT Token，過期時間 7 天 |
| 資料保護 | 不儲存 Line 密碼，僅儲存授權 Token |

### 5.4 相容性需求 (Compatibility)

| 平台 | 支援版本 |
|------|----------|
| iOS Safari | iOS 14+ |
| Android Chrome | Android 10+ |
| Desktop Chrome | 最新兩個版本 |
| Desktop Safari | 最新兩個版本 |

### 5.5 可擴展性需求 (Scalability)

| 階段 | 預估量 | 架構支援 |
|------|--------|----------|
| MVP | 100 活動/月，500 用戶 | 單一伺服器 |
| Phase 2 | 1,000 活動/月，5,000 用戶 | 水平擴展 |

---

## 6. 技術架構建議 (Technical Architecture Recommendations)

### 6.1 前端架構

| 項目 | 建議 | 理由 |
|------|------|------|
| 框架 | Next.js 14+ (App Router) | SSR/SSG 支援、SEO 友善 |
| UI 元件 | Tailwind CSS + shadcn/ui | 快速開發、一致性 |
| 地圖 | Google Maps JavaScript API | 台灣地址支援完整 |
| 狀態管理 | React Query (TanStack Query) | Server State 最佳實踐 |

### 6.2 後端架構

| 項目 | 建議 | 理由 |
|------|------|------|
| 語言/框架 | Go + Gin 或 Node.js + Fastify | 高效能、快速開發 |
| API 風格 | REST API (後期可考慮 GraphQL) | 簡單明確 |
| 認證 | Line Login SDK | 官方支援 |

### 6.3 資料庫架構

| 項目 | 建議 | 理由 |
|------|------|------|
| 主資料庫 | PostgreSQL | 關聯式資料、地理查詢 (PostGIS) |
| 快取 | Redis | Session、熱門活動快取 |

### 6.4 部署架構

| 項目 | 建議 | 理由 |
|------|------|------|
| 雲端平台 | GCP (Google Cloud Platform) | 整合 Google Maps API |
| 容器化 | Docker + Cloud Run | 自動擴展、按需計費 |
| CDN | Cloudflare | 免費方案足夠 MVP |

### 6.5 核心資料模型 (ERD 概要)

```
┌──────────────┐       ┌──────────────┐
│    User      │       │    Event     │
├──────────────┤       ├──────────────┤
│ id (PK)      │       │ id (PK)      │
│ line_user_id │       │ host_id (FK) │──────┐
│ display_name │       │ title        │      │
│ avatar_url   │       │ datetime     │      │
│ created_at   │       │ location     │      │
└──────────────┘       │ coordinates  │      │
       │               │ capacity     │      │
       │               │ skill_level  │      │
       │               │ fee          │      │
       │               │ status       │      │
       │               │ created_at   │      │
       │               └──────────────┘      │
       │                      │              │
       │                      │              │
       ▼                      ▼              │
┌────────────────────────────────────────┐  │
│           Registration                  │  │
├────────────────────────────────────────┤  │
│ id (PK)                                │  │
│ event_id (FK)                          │──┘
│ user_id (FK)                           │
│ status (confirmed/waitlist/cancelled)  │
│ waitlist_position                      │
│ created_at                             │
└────────────────────────────────────────┘
```

---

## 7. 成功指標 (Success Metrics/KPIs)

### 7.1 北極星指標 (North Star Metric)

> **每週成功媒合的打球人次**
>
> 定義：該週所有「已舉行」活動的總報名人數

### 7.2 關鍵指標分解

| 類別 | 指標 | MVP 目標 (上線後 3 個月) |
|------|------|--------------------------|
| **獲取 (Acquisition)** | 週活躍用戶 (WAU) | 200 |
| | 自然搜尋流量佔比 | 20% |
| **活化 (Activation)** | 訪客 -> 報名轉化率 | 15% |
| | 首次訪問 -> 報名時間 | < 60 秒 |
| **留存 (Retention)** | 週報名留存率 (W1) | 30% |
| | 團主回訪開團率 | 50% |
| **推薦 (Referral)** | 每活動平均分享次數 | 3 |
| **供給側 (Supply)** | 週新增活動數 | 50 |
| | 活動滿團率 | 60% |

### 7.3 指標監測工具

| 工具 | 用途 |
|------|------|
| Google Analytics 4 | 流量、轉化漏斗 |
| Mixpanel / Amplitude | 用戶行為追蹤 |
| Sentry | 錯誤監控 |
| Uptime Robot | 服務可用性 |

---

## 8. 第一階段里程碑 (Phase 1 Milestones)

### 8.1 里程碑時程表

| 里程碑 | 週數 | 交付項目 | 驗證點 |
|--------|------|----------|--------|
| **M0: 專案啟動** | W0 | PRD 確認、技術選型、環境建置 | 團隊 Kick-off |
| **M1: 核心骨架** | W1-W2 | 資料庫 Schema、API 框架、Line Login 整合 | 可登入/登出 |
| **M2: 團主流程** | W3-W4 | 建立活動、活動詳情頁、OG Tags | 可建立並分享活動 |
| **M3: 球友流程** | W5-W6 | 地圖瀏覽、篩選、報名/取消、候補 | 完整報名流程 |
| **M4: SEO & 優化** | W7-W8 | SSR 渲染、Sitemap、效能調校 | Lighthouse > 80 |
| **M5: Closed Beta** | W9-W10 | 內部測試、Bug 修復 | 10 位真實用戶測試 |
| **M6: Public Launch** | W11-W12 | 正式上線、監控建置 | 首週 10 場活動 |

### 8.2 風險與緩解措施

| 風險 | 可能性 | 影響 | 緩解措施 |
|------|--------|------|----------|
| Line Login API 審核延遲 | 中 | 高 | 提前申請、準備 Email 登入備案 |
| Google Maps API 費用超支 | 低 | 中 | 設定預算上限、實作快取 |
| 冷啟動無活動內容 | 高 | 高 | 種子用戶計畫、團隊自建活動 |
| SEO 效果緩慢 | 高 | 中 | 同步經營社群、付費廣告補充 |

### 8.3 Go/No-Go 標準

上線前必須滿足的最低標準：

- [ ] 完整流程可用：建立活動 -> 分享 -> 報名 -> 滿團
- [ ] Line Login 正常運作
- [ ] 行動裝置體驗順暢
- [ ] 無 P0/P1 等級 Bug
- [ ] 效能指標達標 (LCP < 2.5s)
- [ ] 至少 5 位種子團主承諾開團

---

## 附錄 A：名詞定義

| 名詞 | 定義 |
|------|------|
| 團主 (Host) | 發起活動的用戶 |
| 球友 (Seeker) | 尋找並報名活動的用戶 |
| 正取 | 在人數上限內成功報名的狀態 |
| 候補 | 活動已滿後加入等待名單的狀態 |
| 滿團 | 報名人數達到人數上限 |
| DUPR | 匹克球官方評分系統 (Dynamic Universal Pickleball Rating) |

---

## 附錄 B：User Story 與 FR 對應表

| User Story | Functional Requirement |
|------------|------------------------|
| US-01 | FR-A01 |
| US-02 | FR-A02 |
| US-03 | FR-A03 |
| US-04 | FR-B01 |
| US-05 | FR-B02 |
| US-06 | FR-B03 |
| US-07 | FR-C01 |

---

## 文件變更紀錄

| 版本 | 日期 | 變更內容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-01-20 | 初版建立 | Product Team |

---

*本文件為 Pickle Go Phase 1 MVP 的產品需求規格，後續將依據開發進度與用戶回饋持續迭代。*
