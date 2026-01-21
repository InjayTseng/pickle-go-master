# Pickle Go 問題修復實作計劃

## 1. 項目概述

### 項目背景
Pickle Go 是一個匹克球活動報名平台，使用 Go + Gin 後端和 Next.js 前端。目前系統存在多個競態條件和安全性問題，需要緊急修復以確保資料一致性和使用者安全。

### 核心問題清單
| 優先級 | 問題類型 | 問題描述 | 影響 |
|--------|----------|----------|------|
| P0 | 競態條件 | 報名系統 Check-Then-Act 非原子操作 | 超額報名 |
| P0 | 競態條件 | 候補遞補無事務保護 | 重複遞補 |
| P0 | 競態條件 | 候補位置分配無鎖定 | 位置重複 |
| P0 | 競態條件 | PromoteFromWaitlist 多步驟無事務 | 資料不一致 |
| P0 | 邏輯錯誤 | 重新報名會違反 UNIQUE 約束 | 使用者體驗問題 |
| P0 | 安全性 | Cookie 缺少安全標記 | XSS Token 竊取 |
| P0 | 安全性 | CSRF State 未驗證 | CSRF 攻擊 |
| P1 | 競態條件 | Token Refresh 多次觸發 | 效能問題 |
| P1 | 測試覆蓋 | 後端 0% 測試覆蓋率 | 品質風險 |

### 技術約束
- 後端使用 `jmoiron/sqlx` 進行資料庫操作
- PostgreSQL 15 支援 `SELECT FOR UPDATE` 和 `SERIALIZABLE` 隔離級別
- 前端使用 `js-cookie` 處理 Cookie
- 必須向後相容現有 API 介面

---

## 2. 架構決策記錄 (ADR)

### ADR-001: 報名系統競態條件解決方案

- **狀態**: 已決定
- **背景**:
  報名系統存在典型的 Check-Then-Act 競態條件。當前流程：
  1. 查詢 `confirmedCount`
  2. 判斷是否有名額
  3. 建立報名記錄

  在高併發場景下，多個請求可能同時通過步驟 1-2，導致超額報名。

- **決策**: 採用 **SELECT FOR UPDATE + 事務** 方案

- **備選方案**:
  | 方案 | 優點 | 缺點 |
  |------|------|------|
  | 樂觀鎖 (version column) | 無鎖等待 | 需要重試邏輯，schema 變更 |
  | SERIALIZABLE 隔離級別 | 最強一致性 | 效能影響大，需處理序列化失敗 |
  | **SELECT FOR UPDATE** | 精確鎖定，效能可控 | 需要事務管理 |
  | 應用層分布式鎖 (Redis) | 可跨實例 | 增加依賴，複雜度高 |

- **理由**:
  1. `SELECT FOR UPDATE` 可以精確鎖定相關的 event 記錄
  2. 報名操作時間短，鎖競爭影響小
  3. 不需要額外的 schema 變更
  4. 不需要額外的基礎設施 (Redis)
  5. PostgreSQL 原生支援，實作簡單可靠

- **影響**:
  - 需要重構 Repository 層，支援事務傳遞
  - 報名 API 響應時間可能略微增加 (毫秒級)
  - 高併發時會有短暫的鎖等待

### ADR-002: 候補遞補原子性保證

- **狀態**: 已決定
- **背景**:
  取消報名時的候補遞補涉及多個步驟：
  1. 取得第一位候補
  2. 更新該候補狀態為 confirmed
  3. 重排其他候補的位置

  這些步驟目前未包在事務中，可能導致部分失敗。

- **決策**: 將整個遞補流程包在單一事務中，使用 `SELECT FOR UPDATE SKIP LOCKED`

- **理由**:
  1. `SKIP LOCKED` 可以避免多個取消同時鎖定同一個候補
  2. 事務確保原子性，要嘛全部成功，要嘛全部失敗
  3. 搭配 Advisory Lock 可以確保同一活動的遞補是序列化的

- **影響**:
  - Repository 需要支援事務注入
  - 需要新增資料庫 migration 來優化索引

### ADR-003: Cookie 安全性強化

- **狀態**: 已決定
- **背景**:
  目前 Cookie 設定缺少安全標記，Token 可能被 XSS 攻擊竊取。

- **決策**:
  - Access Token: 使用 `httpOnly=false`（需要在 API 請求中使用），但設定 `secure=true`, `sameSite=strict`
  - Refresh Token: 使用 `httpOnly=true`, `secure=true`, `sameSite=strict`

- **備選方案**:
  | 方案 | 優點 | 缺點 |
  |------|------|------|
  | 全部 httpOnly | 最安全 | 需要後端處理所有請求的 token 注入 |
  | **Access Token 可讀** | 相容現有架構 | XSS 風險略高，但 sameSite 可緩解 |
  | 改用 Session Cookie | 更安全 | 需要後端 session 儲存 |

- **理由**:
  1. 目前架構 API client 需要讀取 access_token
  2. 加上 `sameSite=strict` 可以防止 CSRF
  3. `secure=true` 確保只在 HTTPS 傳輸

- **影響**:
  - 開發環境需要特殊處理 (localhost 例外)
  - 需要更新 AuthContext 中的 Cookie 設定

### ADR-004: CSRF State 驗證機制

- **狀態**: 已決定
- **背景**:
  Line Login 回呼時，前端傳送的 state 參數未在後端驗證，可能遭受 CSRF 攻擊。

- **決策**: 實作基於 Redis/Memory 的 State 驗證機制

- **實作細節**:
  1. 前端生成 state 時，同時發送請求到後端儲存
  2. 後端在 `/auth/line/callback` 時驗證 state
  3. State 設定 5 分鐘過期時間

- **簡化替代方案**（適用於 MVP）:
  - 使用 HMAC 簽名的 state，後端驗證簽名而非儲存 state
  - 格式：`timestamp:random:hmac(timestamp:random)`

- **影響**:
  - 需要新增 state 驗證邏輯
  - 如果用 HMAC 方案，需要新增環境變數

---

## 3. 系統架構

### 3.1 報名系統事務流程

```
┌──────────────────────────────────────────────────────────────────┐
│                        報名請求處理流程                           │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────┐     ┌──────────────────────────────────────────┐   │
│  │ Handler │────>│              BEGIN TRANSACTION            │   │
│  └─────────┘     └──────────────────────────────────────────┘   │
│                                     │                            │
│                                     v                            │
│              ┌──────────────────────────────────────────┐       │
│              │ SELECT * FROM events WHERE id = $1       │       │
│              │ FOR UPDATE                                │       │
│              │ (鎖定 event 記錄，防止併發修改)            │       │
│              └──────────────────────────────────────────┘       │
│                                     │                            │
│                                     v                            │
│              ┌──────────────────────────────────────────┐       │
│              │ SELECT COUNT(*) FROM registrations       │       │
│              │ WHERE event_id = $1 AND status = 'confirmed'     │
│              └──────────────────────────────────────────┘       │
│                                     │                            │
│                     ┌───────────────┴───────────────┐            │
│                     v                               v            │
│              ┌─────────────┐              ┌─────────────┐       │
│              │ count < cap │              │ count >= cap│       │
│              │ => CONFIRMED│              │ => WAITLIST │       │
│              └─────────────┘              └─────────────┘       │
│                     │                               │            │
│                     v                               v            │
│              ┌──────────────────────────────────────────┐       │
│              │ INSERT/UPDATE registration               │       │
│              │ (使用 ON CONFLICT 處理重新報名)           │       │
│              └──────────────────────────────────────────┘       │
│                                     │                            │
│                                     v                            │
│              ┌──────────────────────────────────────────┐       │
│              │              COMMIT                       │       │
│              └──────────────────────────────────────────┘       │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 3.2 候補遞補事務流程

```
┌──────────────────────────────────────────────────────────────────┐
│                        取消報名 + 遞補流程                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────┐     ┌──────────────────────────────────────────┐   │
│  │ Handler │────>│              BEGIN TRANSACTION            │   │
│  └─────────┘     └──────────────────────────────────────────┘   │
│                                     │                            │
│                                     v                            │
│              ┌──────────────────────────────────────────┐       │
│              │ UPDATE registrations                      │       │
│              │ SET status = 'cancelled'                  │       │
│              │ WHERE id = $1                             │       │
│              └──────────────────────────────────────────┘       │
│                                     │                            │
│                     ┌───────────────┴───────────────┐            │
│                     │ wasConfirmed?                  │            │
│                     v                               v            │
│              ┌─────────────┐              ┌─────────────┐       │
│              │    YES      │              │     NO      │       │
│              └─────────────┘              └─────────────┘       │
│                     │                               │            │
│                     v                               │            │
│              ┌──────────────────────────────────────────┐       │
│              │ SELECT * FROM registrations              │       │
│              │ WHERE event_id = $1 AND status = 'waitlist'      │
│              │ ORDER BY waitlist_position ASC           │       │
│              │ LIMIT 1                                  │       │
│              │ FOR UPDATE SKIP LOCKED                   │       │
│              └──────────────────────────────────────────┘       │
│                     │                               │            │
│                     v                               │            │
│              ┌──────────────────────────────────────────┐       │
│              │ UPDATE registrations                      │       │
│              │ SET status = 'confirmed',                 │       │
│              │     waitlist_position = NULL              │       │
│              │ WHERE id = $promoted_id                   │       │
│              └──────────────────────────────────────────┘       │
│                     │                               │            │
│                     v                               │            │
│              ┌──────────────────────────────────────────┐       │
│              │ UPDATE registrations                      │       │
│              │ SET waitlist_position = waitlist_position - 1    │
│              │ WHERE event_id = $1 AND status = 'waitlist'      │
│              └──────────────────────────────────────────┘       │
│                     │                               │            │
│                     └───────────────┬───────────────┘            │
│                                     v                            │
│              ┌──────────────────────────────────────────┐       │
│              │              COMMIT                       │       │
│              └──────────────────────────────────────────┘       │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 3.3 技術棧更新

| 層級 | 現有技術 | 變更 |
|------|----------|------|
| Handler | Gin + 直接調用 Repo | + 事務管理 |
| Repository | sqlx | + 事務支援接口 |
| Database | PostgreSQL | + 優化索引 |
| Frontend Auth | js-cookie | + 安全標記 |

---

## 4. 任務分解

### 階段 1: 後端競態條件修復 (P0)

| 任務ID | 任務描述 | 優先級 | 依賴 | 預估複雜度 | 負責智能體 | 驗收標準 |
|--------|----------|--------|------|------------|------------|----------|
| T1.1 | 建立事務管理基礎設施 | P0 | - | 中 | Backend Agent | 新增 `TxManager` 接口和實作 |
| T1.2 | 重構 RegistrationRepository 支援事務 | P0 | T1.1 | 中 | Backend Agent | 所有方法支援 `*sqlx.Tx` 參數 |
| T1.3 | 實作 RegisterEventWithLock | P0 | T1.2 | 高 | Backend Agent | 使用 FOR UPDATE 的原子報名 |
| T1.4 | 實作 CancelWithPromotion 事務 | P0 | T1.2 | 高 | Backend Agent | 取消+遞補在同一事務 |
| T1.5 | 處理重新報名邏輯 (UPDATE vs INSERT) | P0 | T1.3 | 中 | Backend Agent | 使用 UPSERT 避免約束違反 |
| T1.6 | 修復 GetNextWaitlistPosition 競態 | P0 | T1.2 | 低 | Backend Agent | 在事務中使用 MAX() + 1 |

### 階段 2: 前端安全性修復 (P0)

| 任務ID | 任務描述 | 優先級 | 依賴 | 預估複雜度 | 負責智能體 | 驗收標準 |
|--------|----------|--------|------|------------|------------|----------|
| T2.1 | 修復 Cookie 安全設定 | P0 | - | 低 | Frontend Agent | 加入 secure, sameSite 標記 |
| T2.2 | 實作 CSRF State 驗證 (後端) | P0 | - | 中 | Backend Agent | 驗證 OAuth state 參數 |
| T2.3 | 更新前端 State 生成邏輯 | P0 | T2.2 | 低 | Frontend Agent | 配合後端驗證機制 |

### 階段 3: Token Refresh 優化 (P1)

| 任務ID | 任務描述 | 優先級 | 依賴 | 預估複雜度 | 負責智能體 | 驗收標準 |
|--------|----------|--------|------|------------|------------|----------|
| T3.1 | 實作 Token Refresh 佇列機制 | P1 | - | 中 | Frontend Agent | 多請求共用一次 refresh |
| T3.2 | 加入 Refresh 請求去重 | P1 | T3.1 | 低 | Frontend Agent | 使用 Promise 共享 |

### 階段 4: 後端單元測試 (P1)

| 任務ID | 任務描述 | 優先級 | 依賴 | 預估複雜度 | 負責智能體 | 驗收標準 |
|--------|----------|--------|------|------------|------------|----------|
| T4.1 | 建立測試基礎設施 | P1 | - | 低 | Backend Agent | testify, mock 設定 |
| T4.2 | pkg/jwt 單元測試 | P1 | T4.1 | 低 | Backend Agent | 覆蓋率 > 80% |
| T4.3 | pkg/geo 單元測試 | P1 | T4.1 | 低 | Backend Agent | 覆蓋率 > 80% |
| T4.4 | pkg/shortcode 單元測試 | P1 | T4.1 | 低 | Backend Agent | 覆蓋率 > 80% |
| T4.5 | Registration Handler 整合測試 | P1 | T1.6 | 高 | Backend Agent | 競態條件測試通過 |

### 階段 5: 資料庫優化

| 任務ID | 任務描述 | 優先級 | 依賴 | 預估複雜度 | 負責智能體 | 驗收標準 |
|--------|----------|--------|------|------------|------------|----------|
| T5.1 | 新增 migration: 優化索引 | P1 | - | 低 | Backend Agent | 加入複合索引 |

---

## 5. 詳細實作規格

### T1.1 建立事務管理基礎設施

**檔案**: `/apps/api/internal/database/tx.go` (新建)

```go
package database

import (
    "context"
    "github.com/jmoiron/sqlx"
)

// DBTX 是 *sqlx.DB 和 *sqlx.Tx 的共同接口
type DBTX interface {
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
    QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

// TxManager 管理事務
type TxManager struct {
    db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
    return &TxManager{db: db}
}

// WithTx 在事務中執行函數
func (m *TxManager) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
    tx, err := m.db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}
```

### T1.3 實作 RegisterEventWithLock

**檔案**: `/apps/api/internal/repository/registration_repo.go` (修改)

```go
// RegisterWithLock 使用鎖定的報名方法
func (r *RegistrationRepository) RegisterWithLock(
    ctx context.Context,
    tx *sqlx.Tx,
    eventID, userID uuid.UUID,
) (*model.Registration, error) {
    // 1. 鎖定 event 記錄
    var event struct {
        Capacity int    `db:"capacity"`
        Status   string `db:"status"`
        HostID   uuid.UUID `db:"host_id"`
    }
    err := tx.GetContext(ctx, &event,
        `SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`,
        eventID)
    if err != nil {
        return nil, err
    }

    // 2. 檢查活動狀態
    if event.Status == "cancelled" || event.Status == "completed" {
        return nil, ErrEventNotOpen
    }
    if event.HostID == userID {
        return nil, ErrHostCannotRegister
    }

    // 3. 檢查是否已有報名記錄 (包含 cancelled)
    var existingReg model.Registration
    err = tx.GetContext(ctx, &existingReg,
        `SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`,
        eventID, userID)

    hasExisting := err == nil
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }

    if hasExisting && existingReg.Status != model.RegistrationCancelled {
        return nil, ErrAlreadyRegistered
    }

    // 4. 計算確認人數
    var confirmedCount int
    err = tx.GetContext(ctx, &confirmedCount,
        `SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`,
        eventID)
    if err != nil {
        return nil, err
    }

    // 5. 決定狀態
    var status model.RegistrationStatus
    var waitlistPos *int

    if confirmedCount < event.Capacity {
        status = model.RegistrationConfirmed
    } else {
        status = model.RegistrationWaitlist
        // 取得候補位置
        var maxPos *int
        tx.GetContext(ctx, &maxPos,
            `SELECT MAX(waitlist_position) FROM registrations
             WHERE event_id = $1 AND status = 'waitlist'`,
            eventID)
        pos := 1
        if maxPos != nil {
            pos = *maxPos + 1
        }
        waitlistPos = &pos
    }

    // 6. INSERT 或 UPDATE
    reg := &model.Registration{
        EventID:          eventID,
        UserID:           userID,
        Status:           status,
        WaitlistPosition: waitlistPos,
    }

    if hasExisting {
        // UPDATE 已取消的報名
        reg.ID = existingReg.ID
        _, err = tx.ExecContext(ctx, `
            UPDATE registrations
            SET status = $2, waitlist_position = $3,
                registered_at = NOW(),
                confirmed_at = CASE WHEN $2 = 'confirmed' THEN NOW() ELSE NULL END,
                cancelled_at = NULL
            WHERE id = $1`,
            reg.ID, status, waitlistPos)
    } else {
        // INSERT 新報名
        reg.ID = uuid.New()
        _, err = tx.ExecContext(ctx, `
            INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at, confirmed_at)
            VALUES ($1, $2, $3, $4, $5, NOW(), CASE WHEN $4 = 'confirmed' THEN NOW() ELSE NULL END)`,
            reg.ID, eventID, userID, status, waitlistPos)
    }

    if err != nil {
        return nil, err
    }

    return reg, nil
}
```

### T1.4 實作 CancelWithPromotion 事務

**檔案**: `/apps/api/internal/repository/registration_repo.go` (修改)

```go
// CancelAndPromote 取消報名並遞補
func (r *RegistrationRepository) CancelAndPromote(
    ctx context.Context,
    tx *sqlx.Tx,
    registrationID, eventID uuid.UUID,
) (*model.Registration, error) {
    // 1. 取得並鎖定報名記錄
    var reg model.Registration
    err := tx.GetContext(ctx, &reg,
        `SELECT * FROM registrations WHERE id = $1 FOR UPDATE`,
        registrationID)
    if err != nil {
        return nil, err
    }

    if reg.Status == model.RegistrationCancelled {
        return nil, ErrAlreadyCancelled
    }

    wasConfirmed := reg.Status == model.RegistrationConfirmed

    // 2. 更新為取消
    _, err = tx.ExecContext(ctx,
        `UPDATE registrations SET status = 'cancelled', cancelled_at = NOW() WHERE id = $1`,
        registrationID)
    if err != nil {
        return nil, err
    }

    // 3. 如果是確認狀態，進行遞補
    var promoted *model.Registration
    if wasConfirmed {
        // 取得第一位候補 (使用 SKIP LOCKED 避免死鎖)
        var waitlistReg model.Registration
        err = tx.GetContext(ctx, &waitlistReg, `
            SELECT * FROM registrations
            WHERE event_id = $1 AND status = 'waitlist'
            ORDER BY waitlist_position ASC
            LIMIT 1
            FOR UPDATE SKIP LOCKED`,
            eventID)

        if err == nil {
            // 遞補
            _, err = tx.ExecContext(ctx, `
                UPDATE registrations
                SET status = 'confirmed', confirmed_at = NOW(), waitlist_position = NULL
                WHERE id = $1`,
                waitlistReg.ID)
            if err != nil {
                return nil, err
            }

            // 重排候補位置
            _, err = tx.ExecContext(ctx, `
                UPDATE registrations
                SET waitlist_position = waitlist_position - 1
                WHERE event_id = $1 AND status = 'waitlist'`,
                eventID)
            if err != nil {
                return nil, err
            }

            promoted = &waitlistReg
            promoted.Status = model.RegistrationConfirmed
        }
    }

    return promoted, nil
}
```

### T2.1 修復 Cookie 安全設定

**檔案**: `/apps/web/src/contexts/AuthContext.tsx` (修改)

```typescript
// Cookie 設定輔助函數
const getCookieOptions = (days: number) => {
  const isProduction = process.env.NODE_ENV === 'production';
  return {
    expires: days,
    secure: isProduction, // 只在 HTTPS 傳輸
    sameSite: 'strict' as const, // 防止 CSRF
    path: '/',
  };
};

// 在 login 函數中
Cookies.set('access_token', response.access_token, getCookieOptions(7));
if (response.refresh_token) {
  Cookies.set('refresh_token', response.refresh_token, getCookieOptions(30));
}
```

### T2.2 實作 CSRF State 驗證

**檔案**: `/apps/api/internal/handler/auth.go` (修改)

```go
// 使用 HMAC 簽名驗證 state
func (h *AuthHandler) validateState(state string) bool {
    if state == "" {
        return false
    }

    parts := strings.Split(state, ":")
    if len(parts) != 3 {
        return false
    }

    timestamp, random, providedHmac := parts[0], parts[1], parts[2]

    // 檢查時間戳 (5分鐘有效)
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return false
    }
    if time.Now().Unix()-ts > 300 {
        return false
    }

    // 驗證 HMAC
    expectedHmac := h.computeStateHmac(timestamp, random)
    return hmac.Equal([]byte(providedHmac), []byte(expectedHmac))
}

func (h *AuthHandler) computeStateHmac(timestamp, random string) string {
    secret := os.Getenv("STATE_SECRET")
    if secret == "" {
        secret = os.Getenv("JWT_SECRET") // 共用 secret
    }
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(timestamp + ":" + random))
    return base64.URLEncoding.EncodeToString(mac.Sum(nil))[:16]
}
```

### T3.1 實作 Token Refresh 佇列機制

**檔案**: `/apps/web/src/lib/api-client.ts` (修改)

```typescript
class ApiClient {
  private refreshPromise: Promise<boolean> | null = null;

  private async refreshToken(): Promise<boolean> {
    // 如果已經有進行中的 refresh，等待它完成
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    const refreshToken = Cookies.get('refresh_token');
    if (!refreshToken) {
      return false;
    }

    // 建立新的 refresh promise
    this.refreshPromise = (async () => {
      try {
        const response = await fetch(`${this.baseUrl}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (!response.ok) {
          return false;
        }

        const data: ApiResponse<AuthResponse> = await response.json();
        if (data.success && data.data) {
          Cookies.set('access_token', data.data.access_token, getCookieOptions(7));
          return true;
        }
        return false;
      } catch {
        return false;
      } finally {
        // 清除 promise，允許下次 refresh
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }
}
```

---

## 6. 資料庫變更

### Migration: 000002_add_registration_indexes.up.sql

```sql
-- 優化報名查詢的複合索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registrations_event_status
ON registrations(event_id, status)
WHERE status != 'cancelled';

-- 優化候補排序查詢
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registrations_waitlist_order
ON registrations(event_id, waitlist_position)
WHERE status = 'waitlist';

-- 優化使用者報名查詢
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registrations_user_active
ON registrations(user_id, event_id)
WHERE status != 'cancelled';
```

### Migration: 000002_add_registration_indexes.down.sql

```sql
DROP INDEX CONCURRENTLY IF EXISTS idx_registrations_event_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_registrations_waitlist_order;
DROP INDEX CONCURRENTLY IF EXISTS idx_registrations_user_active;
```

---

## 7. 測試策略

### 7.1 競態條件測試

**檔案**: `/apps/api/internal/handler/registration_test.go`

```go
func TestRegisterEvent_ConcurrentRequests(t *testing.T) {
    // 建立容量為 2 的活動
    event := createTestEvent(t, capacity: 2)

    // 建立 5 個使用者
    users := createTestUsers(t, 5)

    // 同時發送 5 個報名請求
    var wg sync.WaitGroup
    results := make(chan *RegistrationResult, 5)

    for _, user := range users {
        wg.Add(1)
        go func(u *User) {
            defer wg.Done()
            result := registerForEvent(event.ID, u.Token)
            results <- result
        }(user)
    }

    wg.Wait()
    close(results)

    // 驗證結果
    var confirmed, waitlisted int
    for result := range results {
        if result.Status == "confirmed" {
            confirmed++
        } else if result.Status == "waitlist" {
            waitlisted++
        }
    }

    assert.Equal(t, 2, confirmed, "應該只有 2 人確認報名")
    assert.Equal(t, 3, waitlisted, "應該有 3 人進入候補")

    // 驗證資料庫狀態
    confirmedCount := countConfirmedRegistrations(t, event.ID)
    assert.Equal(t, 2, confirmedCount, "資料庫應該只有 2 筆確認報名")
}

func TestCancelRegistration_ConcurrentPromotion(t *testing.T) {
    // 建立容量為 2 的活動，2 人確認，3 人候補
    event := createTestEvent(t, capacity: 2)
    confirmed := createConfirmedRegistrations(t, event.ID, 2)
    waitlisted := createWaitlistRegistrations(t, event.ID, 3)

    // 同時取消 2 個確認報名
    var wg sync.WaitGroup
    for _, reg := range confirmed {
        wg.Add(1)
        go func(r *Registration) {
            defer wg.Done()
            cancelRegistration(event.ID, r.UserToken)
        }(reg)
    }

    wg.Wait()

    // 驗證: 應該正好遞補 2 人
    newConfirmed := countConfirmedRegistrations(t, event.ID)
    assert.Equal(t, 2, newConfirmed, "應該遞補 2 人")

    newWaitlisted := countWaitlistRegistrations(t, event.ID)
    assert.Equal(t, 1, newWaitlisted, "應該剩餘 1 人候補")
}
```

### 7.2 單元測試覆蓋率目標

| Package | 目標覆蓋率 | 關鍵測試案例 |
|---------|-----------|-------------|
| pkg/jwt | 80% | Token 生成、驗證、過期處理 |
| pkg/geo | 90% | 距離計算、邊界情況 |
| pkg/shortcode | 90% | 生成、驗證、唯一性 |
| handler/registration | 70% | 競態條件、邊界情況 |

---

## 8. 風險與緩解措施

| 風險 | 可能性 | 影響 | 緩解措施 |
|------|--------|------|----------|
| 事務鎖定導致效能下降 | 中 | 中 | 監控鎖等待時間，必要時加入超時 |
| Migration 在生產環境失敗 | 低 | 高 | 使用 CONCURRENTLY，先在 staging 測試 |
| Cookie 安全設定影響開發環境 | 高 | 低 | 開發環境條件判斷 |
| 遞補通知失敗 | 中 | 低 | 事務外異步處理，失敗重試 |

---

## 9. 里程碑

| 里程碑 | 完成標準 | 包含任務 | 預估時間 |
|--------|----------|----------|----------|
| M1: 競態條件修復完成 | 所有 P0 競態測試通過 | T1.1-T1.6 | 2-3 天 |
| M2: 安全性修復完成 | Cookie 和 CSRF 驗證生效 | T2.1-T2.3 | 1 天 |
| M3: P1 優化完成 | Token refresh 正常，測試覆蓋率達標 | T3.1-T3.2, T4.1-T4.5 | 2-3 天 |
| M4: 資料庫優化完成 | Migration 成功執行 | T5.1 | 0.5 天 |

---

## 10. 附錄：關鍵檔案變更清單

### 後端

| 檔案路徑 | 變更類型 | 說明 |
|----------|----------|------|
| `internal/database/tx.go` | 新增 | 事務管理器 |
| `internal/repository/registration_repo.go` | 修改 | 新增事務支援方法 |
| `internal/handler/registration.go` | 修改 | 使用事務包裝報名邏輯 |
| `internal/handler/auth.go` | 修改 | 新增 state 驗證 |
| `pkg/jwt/jwt_test.go` | 新增 | JWT 單元測試 |
| `pkg/geo/geo_test.go` | 新增 | Geo 單元測試 |
| `pkg/shortcode/shortcode_test.go` | 新增 | Shortcode 單元測試 |
| `migrations/000002_*.sql` | 新增 | 索引優化 migration |

### 前端

| 檔案路徑 | 變更類型 | 說明 |
|----------|----------|------|
| `src/contexts/AuthContext.tsx` | 修改 | Cookie 安全設定 |
| `src/lib/api-client.ts` | 修改 | Token refresh 佇列 |
