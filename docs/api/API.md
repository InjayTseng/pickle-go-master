# Pickle Go API 文件

## 概述

Pickle Go 提供 RESTful API，用於管理匹克球揪團活動。所有 API 端點都使用 JSON 格式進行資料交換。

**Base URL**: `https://api.picklego.tw/api/v1` (生產環境)
**Base URL**: `http://localhost:8080/api/v1` (開發環境)

## 認證方式

### JWT Token 認證

大部分 API 端點需要使用 JWT Token 進行認證。Token 需要在 HTTP Header 中提供：

```http
Authorization: Bearer {access_token}
```

### Token 類型

- **Access Token**: 有效期 168 小時（7 天）
- **Refresh Token**: 用於更新 Access Token

## 通用回應格式

### 成功回應

```json
{
  "success": true,
  "data": {
    // 回應資料
  }
}
```

### 錯誤回應

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "錯誤訊息",
    "details": {}
  }
}
```

## 常見錯誤代碼

| 錯誤代碼 | HTTP 狀態碼 | 說明 |
|---------|-----------|------|
| `VALIDATION_ERROR` | 400 | 請求參數驗證失敗 |
| `UNAUTHORIZED` | 401 | 未認證或 Token 無效 |
| `FORBIDDEN` | 403 | 無權限執行此操作 |
| `NOT_FOUND` | 404 | 資源不存在 |
| `ALREADY_REGISTERED` | 400 | 已經報名此活動 |
| `EVENT_CANCELLED` | 400 | 活動已取消 |
| `EVENT_COMPLETED` | 400 | 活動已結束 |
| `INTERNAL_ERROR` | 500 | 伺服器內部錯誤 |

---

## API 端點

## 1. 認證相關 (Auth)

### 1.1 Line Login 回呼

處理 Line Login OAuth 回呼並建立使用者登入 Session。

**端點**: `POST /auth/line/callback`
**認證**: 不需要

#### 請求參數

```json
{
  "code": "string (required)",
  "state": "string (optional)",
  "redirect_uri": "string (optional)"
}
```

#### 範例請求

```bash
curl -X POST https://api.picklego.tw/api/v1/auth/line/callback \
  -H "Content-Type: application/json" \
  -d '{
    "code": "abc123xyz",
    "state": "random-state-string"
  }'
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "王小明",
      "avatar_url": "https://profile.line-scdn.net/...",
      "email": null
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

#### 錯誤回應

- `401 LINE_AUTH_FAILED`: Line 認證失敗
- `500 LINE_PROFILE_FAILED`: 無法取得 Line 使用者資料
- `500 USER_CREATION_FAILED`: 建立使用者失敗
- `500 TOKEN_GENERATION_FAILED`: 產生 Token 失敗

---

### 1.2 更新 Token

使用 Refresh Token 取得新的 Access Token。

**端點**: `POST /auth/refresh`
**認證**: 不需要

#### 請求參數

```json
{
  "refresh_token": "string (required)"
}
```

#### 範例請求

```bash
curl -X POST https://api.picklego.tw/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "王小明",
      "avatar_url": "https://profile.line-scdn.net/..."
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

#### 錯誤回應

- `401 TOKEN_EXPIRED`: Refresh Token 已過期
- `401 INVALID_TOKEN`: Refresh Token 無效
- `401 USER_NOT_FOUND`: 使用者不存在

---

### 1.3 登出

登出目前使用者（主要在客戶端清除 Token）。

**端點**: `POST /auth/logout`
**認證**: 需要

#### 範例請求

```bash
curl -X POST https://api.picklego.tw/api/v1/auth/logout \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  }
}
```

---

## 2. 使用者相關 (Users)

### 2.1 取得目前使用者資訊

取得目前登入使用者的個人資訊。

**端點**: `GET /users/me`
**認證**: 需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/users/me \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "display_name": "王小明",
    "avatar_url": "https://profile.line-scdn.net/...",
    "email": null
  }
}
```

---

### 2.2 取得我主辦的活動

取得目前使用者主辦的所有活動。

**端點**: `GET /users/me/events`
**認證**: 需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/users/me/events \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "title": "週末輕鬆打",
        "event_date": "2026-01-25",
        "start_time": "19:00:00",
        "end_time": "21:00:00",
        "location": {
          "name": "大安森林公園網球場",
          "address": "台北市大安區新生南路二段1號",
          "lat": 25.0292,
          "lng": 121.5367
        },
        "capacity": 8,
        "skill_level": "intermediate",
        "skill_level_label": "中階 (2.5-3.5)",
        "fee": 200,
        "status": "open"
      }
    ],
    "total": 1
  }
}
```

---

### 2.3 取得我的報名記錄

取得目前使用者報名的所有活動。

**端點**: `GET /users/me/registrations`
**認證**: 需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/users/me/registrations \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "registrations": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "event_id": "660e8400-e29b-41d4-a716-446655440000",
        "status": "confirmed",
        "waitlist_position": null,
        "registered_at": "2026-01-21T10:30:00Z",
        "event": {
          "id": "660e8400-e29b-41d4-a716-446655440000",
          "title": "週末輕鬆打",
          "event_date": "2026-01-25",
          "start_time": "19:00:00",
          "location": "大安森林公園網球場",
          "skill_level": "intermediate",
          "status": "open"
        }
      }
    ],
    "total": 1
  }
}
```

---

### 2.4 取得我的通知

取得目前使用者的通知列表。

**端點**: `GET /users/me/notifications`
**認證**: 需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/users/me/notifications \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "880e8400-e29b-41d4-a716-446655440000",
        "type": "waitlist_promoted",
        "title": "您已從候補轉為正取",
        "message": "活動「週末輕鬆打」有人取消報名，您已從候補轉為正取！",
        "event_id": "660e8400-e29b-41d4-a716-446655440000",
        "is_read": false,
        "created_at": "2026-01-21T12:00:00Z"
      }
    ],
    "total": 1,
    "unread_count": 1
  }
}
```

---

## 3. 活動相關 (Events)

### 3.1 列出活動

取得活動列表，支援地理位置篩選和技能等級篩選。

**端點**: `GET /events`
**認證**: 不需要

#### 查詢參數

| 參數 | 類型 | 必填 | 說明 |
|-----|------|-----|------|
| `lat` | float | 否 | 緯度 |
| `lng` | float | 否 | 經度 |
| `radius` | int | 否 | 搜尋半徑（公尺），預設 10000，最大 50000 |
| `skill_level` | string | 否 | 技能等級篩選 (beginner/intermediate/advanced/expert/any) |
| `status` | string | 否 | 活動狀態 (open/full/cancelled/completed) |
| `limit` | int | 否 | 回傳數量，預設 20，最大 100 |
| `offset` | int | 否 | 偏移量，用於分頁 |

#### 範例請求

```bash
curl -X GET "https://api.picklego.tw/api/v1/events?lat=25.0330&lng=121.5654&radius=5000&skill_level=intermediate&limit=10" \
  -H "Content-Type: application/json"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "host": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "display_name": "王小明",
          "avatar_url": "https://profile.line-scdn.net/..."
        },
        "title": "週末輕鬆打",
        "description": "歡迎新手參加，請自備球拍",
        "event_date": "2026-01-25",
        "start_time": "19:00:00",
        "end_time": "21:00:00",
        "location": {
          "name": "大安森林公園網球場",
          "address": "台北市大安區新生南路二段1號",
          "lat": 25.0292,
          "lng": 121.5367,
          "google_place_id": "ChIJXxZ..."
        },
        "capacity": 8,
        "confirmed_count": 5,
        "waitlist_count": 2,
        "skill_level": "intermediate",
        "skill_level_label": "中階 (2.5-3.5)",
        "fee": 200,
        "status": "open"
      }
    ],
    "total": 1,
    "has_more": false
  }
}
```

---

### 3.2 取得單一活動

根據活動 ID 取得活動詳細資訊。

**端點**: `GET /events/:id`
**認證**: 不需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "host": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "王小明",
      "avatar_url": "https://profile.line-scdn.net/..."
    },
    "title": "週末輕鬆打",
    "description": "歡迎新手參加，請自備球拍",
    "event_date": "2026-01-25",
    "start_time": "19:00:00",
    "end_time": "21:00:00",
    "location": {
      "name": "大安森林公園網球場",
      "address": "台北市大安區新生南路二段1號",
      "lat": 25.0292,
      "lng": 121.5367,
      "google_place_id": "ChIJXxZ..."
    },
    "capacity": 8,
    "confirmed_count": 5,
    "waitlist_count": 2,
    "skill_level": "intermediate",
    "skill_level_label": "中階 (2.5-3.5)",
    "fee": 200,
    "status": "open"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤
- `404 NOT_FOUND`: 活動不存在

---

### 3.3 透過短網址代碼取得活動

根據短網址代碼（short code）取得活動詳細資訊。

**端點**: `GET /events/by-code/:code`
**認證**: 不需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/events/by-code/abc123
```

#### 成功回應 (200 OK)

回應格式與「取得單一活動」相同。

#### 錯誤回應

- `400 VALIDATION_ERROR`: 短網址代碼為空
- `404 NOT_FOUND`: 活動不存在

---

### 3.4 建立活動

建立新的活動。

**端點**: `POST /events`
**認證**: 需要

#### 請求參數

```json
{
  "title": "string (optional)",
  "description": "string (optional)",
  "event_date": "string (required, format: YYYY-MM-DD)",
  "start_time": "string (required, format: HH:MM)",
  "end_time": "string (optional, format: HH:MM)",
  "location": {
    "name": "string (required)",
    "address": "string (optional)",
    "lat": "float (required)",
    "lng": "float (required)",
    "google_place_id": "string (optional)"
  },
  "capacity": "int (required, min: 4, max: 20)",
  "skill_level": "string (required, enum: beginner|intermediate|advanced|expert|any)",
  "fee": "int (optional, min: 0, max: 9999)"
}
```

#### 範例請求

```bash
curl -X POST https://api.picklego.tw/api/v1/events \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "週末輕鬆打",
    "description": "歡迎新手參加",
    "event_date": "2026-01-25",
    "start_time": "19:00",
    "end_time": "21:00",
    "location": {
      "name": "大安森林公園網球場",
      "address": "台北市大安區新生南路二段1號",
      "lat": 25.0292,
      "lng": 121.5367,
      "google_place_id": "ChIJXxZ..."
    },
    "capacity": 8,
    "skill_level": "intermediate",
    "fee": 200
  }'
```

#### 成功回應 (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "share_url": "https://picklego.tw/g/abc123"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 請求參數驗證失敗
- `401 UNAUTHORIZED`: 未認證
- `500 INTERNAL_ERROR`: 建立活動失敗

---

### 3.5 更新活動

更新現有活動（僅活動主辦人可以更新）。

**端點**: `PUT /events/:id`
**認證**: 需要

#### 請求參數

所有參數皆為選填，只需傳送要更新的欄位。

```json
{
  "title": "string (optional)",
  "description": "string (optional)",
  "event_date": "string (optional, format: YYYY-MM-DD)",
  "start_time": "string (optional, format: HH:MM)",
  "end_time": "string (optional, format: HH:MM)",
  "capacity": "int (optional, min: 4, max: 20)",
  "skill_level": "string (optional, enum: beginner|intermediate|advanced|expert|any)",
  "fee": "int (optional, min: 0, max: 9999)",
  "status": "string (optional, enum: open|full|cancelled)"
}
```

#### 範例請求

```bash
curl -X PUT https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "週末進階訓練",
    "capacity": 10
  }'
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "message": "Event updated successfully"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤或參數驗證失敗
- `401 UNAUTHORIZED`: 未認證
- `403 FORBIDDEN`: 您不是此活動的主辦人
- `404 NOT_FOUND`: 活動不存在
- `500 INTERNAL_ERROR`: 更新失敗

---

### 3.6 取消活動

取消活動（僅活動主辦人可以取消）。此操作會將活動狀態設為 `cancelled` 並取消所有報名。

**端點**: `DELETE /events/:id`
**認證**: 需要

#### 範例請求

```bash
curl -X DELETE https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "message": "Event cancelled successfully"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤
- `401 UNAUTHORIZED`: 未認證
- `403 FORBIDDEN`: 您不是此活動的主辦人
- `500 INTERNAL_ERROR`: 取消失敗

---

## 4. 報名相關 (Registrations)

### 4.1 報名活動

報名參加活動。如果活動已額滿，將自動加入候補名單。

**端點**: `POST /events/:id/register`
**認證**: 需要

#### 範例請求

```bash
curl -X POST https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000/register \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (201 Created)

報名成功（正取）：

```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "event_id": "660e8400-e29b-41d4-a716-446655440000",
    "status": "confirmed",
    "waitlist_position": null,
    "message": "報名成功！"
  }
}
```

加入候補：

```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "event_id": "660e8400-e29b-41d4-a716-446655440000",
    "status": "waitlist",
    "waitlist_position": 3,
    "message": "已加入候補（第 3 位）"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤
- `400 ALREADY_REGISTERED`: 您已經報名此活動
- `400 EVENT_CANCELLED`: 活動已取消
- `400 EVENT_COMPLETED`: 活動已結束
- `400 HOST_CANNOT_REGISTER`: 您不能報名自己主辦的活動
- `401 UNAUTHORIZED`: 未認證
- `404 NOT_FOUND`: 活動不存在
- `500 INTERNAL_ERROR`: 報名失敗

---

### 4.2 取消報名

取消報名活動。如果是正取名單取消，將自動提升第一位候補者為正取。

**端點**: `DELETE /events/:id/register`
**認證**: 需要

#### 範例請求

```bash
curl -X DELETE https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000/register \
  -H "Authorization: Bearer {access_token}"
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "message": "Registration cancelled successfully"
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤
- `400 ALREADY_CANCELLED`: 報名已取消
- `401 UNAUTHORIZED`: 未認證
- `404 NOT_FOUND`: 您未報名此活動
- `500 INTERNAL_ERROR`: 取消失敗

---

### 4.3 取得活動報名名單

取得活動的所有報名者（包含正取和候補）。

**端點**: `GET /events/:id/registrations`
**認證**: 不需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/api/v1/events/660e8400-e29b-41d4-a716-446655440000/registrations
```

#### 成功回應 (200 OK)

```json
{
  "success": true,
  "data": {
    "confirmed": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "user": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "display_name": "王小明",
          "avatar_url": "https://profile.line-scdn.net/..."
        },
        "registered_at": "2026-01-21T10:30:00Z"
      }
    ],
    "waitlist": [
      {
        "id": "880e8400-e29b-41d4-a716-446655440000",
        "user": {
          "id": "990e8400-e29b-41d4-a716-446655440000",
          "display_name": "李小華",
          "avatar_url": "https://profile.line-scdn.net/..."
        },
        "waitlist_position": 1,
        "registered_at": "2026-01-21T11:00:00Z"
      }
    ],
    "confirmed_count": 1,
    "waitlist_count": 1
  }
}
```

#### 錯誤回應

- `400 VALIDATION_ERROR`: 活動 ID 格式錯誤
- `404 NOT_FOUND`: 活動不存在
- `500 INTERNAL_ERROR`: 取得報名名單失敗

---

## 5. 健康檢查

### 5.1 Health Check

用於檢查 API 服務狀態，適合用於監控和負載平衡器健康檢查。

**端點**: `GET /health`
**認證**: 不需要

#### 範例請求

```bash
curl -X GET https://api.picklego.tw/health
```

#### 成功回應 (200 OK)

```json
{
  "status": "ok",
  "service": "pickle-go-api",
  "version": "0.1.0"
}
```

---

## 6. 速率限制

為了保護 API 服務，生產環境會啟用速率限制：

### 一般端點

- 限制: 100 次請求 / 分鐘 / IP
- 超過限制回應: `429 Too Many Requests`

### 認證端點 (嚴格限制)

以下端點有更嚴格的速率限制：
- `POST /auth/line/callback`
- `POST /auth/refresh`
- `POST /auth/logout`

- 限制: 10 次請求 / 分鐘 / IP
- 超過限制回應: `429 Too Many Requests`

---

## 7. 資料型別定義

### SkillLevel (技能等級)

| 值 | 標籤 | 說明 |
|----|------|------|
| `beginner` | 新手友善 (2.0-2.5) | 適合初學者 |
| `intermediate` | 中階 (2.5-3.5) | 有一定經驗 |
| `advanced` | 進階 (3.5-4.5) | 進階球友 |
| `expert` | 高階 (4.5+) | 高階球友 |
| `any` | 不限程度 | 所有程度都歡迎 |

### EventStatus (活動狀態)

| 值 | 說明 |
|----|------|
| `open` | 開放報名 |
| `full` | 已額滿（但仍可加入候補） |
| `cancelled` | 已取消 |
| `completed` | 已結束 |

### RegistrationStatus (報名狀態)

| 值 | 說明 |
|----|------|
| `confirmed` | 正取 |
| `waitlist` | 候補 |
| `cancelled` | 已取消 |

---

## 8. 使用範例

### 完整流程範例：建立活動並報名

#### 1. Line Login 取得 Token

```javascript
// 前端處理 Line Login 回呼
const response = await fetch('https://api.picklego.tw/api/v1/auth/line/callback', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ code: authCode })
});

const data = await response.json();
const accessToken = data.data.access_token;
// 儲存 Token 到 Cookie 或 localStorage
```

#### 2. 建立活動

```javascript
const response = await fetch('https://api.picklego.tw/api/v1/events', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${accessToken}`
  },
  body: JSON.stringify({
    title: '週末輕鬆打',
    event_date: '2026-01-25',
    start_time: '19:00',
    end_time: '21:00',
    location: {
      name: '大安森林公園網球場',
      lat: 25.0292,
      lng: 121.5367
    },
    capacity: 8,
    skill_level: 'intermediate',
    fee: 200
  })
});

const data = await response.json();
const eventId = data.data.id;
const shareUrl = data.data.share_url; // https://picklego.tw/g/abc123
```

#### 3. 報名活動

```javascript
const response = await fetch(`https://api.picklego.tw/api/v1/events/${eventId}/register`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});

const data = await response.json();
console.log(data.data.message); // "報名成功！" or "已加入候補（第 X 位）"
```

#### 4. 取得報名名單

```javascript
const response = await fetch(`https://api.picklego.tw/api/v1/events/${eventId}/registrations`);
const data = await response.json();

console.log(`正取人數: ${data.data.confirmed_count}`);
console.log(`候補人數: ${data.data.waitlist_count}`);
```

---

## 9. 錯誤處理建議

### 前端錯誤處理範例

```javascript
async function apiCall(url, options) {
  try {
    const response = await fetch(url, options);
    const data = await response.json();

    if (!data.success) {
      // 處理 API 錯誤
      switch (data.error.code) {
        case 'UNAUTHORIZED':
          // Token 過期，嘗試 refresh
          await refreshToken();
          // 重試請求
          return apiCall(url, options);

        case 'ALREADY_REGISTERED':
          alert('您已經報名此活動');
          break;

        case 'EVENT_CANCELLED':
          alert('此活動已取消');
          break;

        default:
          alert(data.error.message);
      }
      throw new Error(data.error.message);
    }

    return data.data;
  } catch (error) {
    console.error('API Error:', error);
    throw error;
  }
}
```

---

## 10. 版本資訊

**當前版本**: v1
**最後更新**: 2026-01-21

### 未來規劃

- WebSocket 支援即時通知
- 活動評論功能
- 使用者評分系統
- 活動照片上傳
- 推薦活動演算法

---

## 11. 支援與聯絡

如有 API 相關問題或建議，請透過以下方式聯絡：

- GitHub Issues: https://github.com/anthropics/pickle-go/issues
- Email: support@picklego.tw
