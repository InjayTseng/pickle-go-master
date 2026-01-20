# Pickle Go 資料庫 Schema 文件

## 概述

Pickle Go 使用 PostgreSQL 資料庫搭配 PostGIS 擴展，用於儲存使用者、活動、報名和通知資料。PostGIS 提供地理空間資料支援，實現基於位置的活動搜尋功能。

**資料庫**: PostgreSQL 14+
**擴展**: PostGIS (地理空間資料)

---

## 資料庫連線資訊

### 開發環境

```
Host: localhost
Port: 5432
Database: picklego
User: postgres
Password: postgres
```

### 生產環境 (Cloud SQL)

```
Connection: /cloudsql/[PROJECT_ID]:[REGION]:[INSTANCE_NAME]
Database: picklego
```

---

## 資料表結構

## 1. users (使用者表)

儲存使用者基本資訊。

### 欄位說明

| 欄位名稱 | 資料型別 | 約束 | 說明 |
|---------|---------|------|------|
| `id` | UUID | PRIMARY KEY | 使用者唯一識別碼 |
| `line_user_id` | VARCHAR(64) | UNIQUE, NOT NULL | Line User ID |
| `display_name` | VARCHAR(100) | NOT NULL | 顯示名稱 |
| `avatar_url` | TEXT | NULLABLE | 頭像 URL |
| `email` | VARCHAR(255) | NULLABLE | Email（保留欄位） |
| `created_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 建立時間 |
| `updated_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 更新時間 |

### 索引

```sql
-- Line User ID 查詢索引
CREATE INDEX idx_users_line_user_id ON users(line_user_id);
```

### 觸發器

```sql
-- 自動更新 updated_at
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### 範例資料

```sql
INSERT INTO users (id, line_user_id, display_name, avatar_url) VALUES
(
  '550e8400-e29b-41d4-a716-446655440000',
  'U1234567890abcdef',
  '王小明',
  'https://profile.line-scdn.net/...'
);
```

---

## 2. events (活動表)

儲存活動資訊。

### 欄位說明

| 欄位名稱 | 資料型別 | 約束 | 說明 |
|---------|---------|------|------|
| `id` | UUID | PRIMARY KEY | 活動唯一識別碼 |
| `host_id` | UUID | FOREIGN KEY (users.id), NOT NULL | 主辦人 ID |
| `title` | VARCHAR(200) | NULLABLE | 活動標題 |
| `description` | TEXT | NULLABLE | 活動說明 |
| `event_date` | DATE | NOT NULL | 活動日期 |
| `start_time` | TIME | NOT NULL | 開始時間 |
| `end_time` | TIME | NULLABLE | 結束時間 |
| `location_name` | VARCHAR(200) | NOT NULL | 地點名稱 |
| `location_address` | VARCHAR(500) | NULLABLE | 地點地址 |
| `location_point` | GEOGRAPHY(POINT, 4326) | NOT NULL | 地理座標（PostGIS） |
| `google_place_id` | VARCHAR(255) | NULLABLE | Google Places ID |
| `capacity` | SMALLINT | NOT NULL, CHECK (4-20) | 人數上限 |
| `skill_level` | VARCHAR(20) | NOT NULL | 技能等級 |
| `fee` | INTEGER | DEFAULT 0, CHECK (0-9999) | 費用（元） |
| `status` | VARCHAR(20) | DEFAULT 'open' | 活動狀態 |
| `short_code` | VARCHAR(10) | UNIQUE | 短網址代碼 |
| `created_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 建立時間 |
| `updated_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 更新時間 |

### 枚舉值

#### skill_level (技能等級)

- `beginner`: 新手友善 (2.0-2.5)
- `intermediate`: 中階 (2.5-3.5)
- `advanced`: 進階 (3.5-4.5)
- `expert`: 高階 (4.5+)
- `any`: 不限程度

#### status (活動狀態)

- `open`: 開放報名
- `full`: 已額滿
- `cancelled`: 已取消
- `completed`: 已結束

### 索引

```sql
-- 地理空間索引（PostGIS GIST）
CREATE INDEX idx_events_location ON events USING GIST(location_point);

-- 其他索引
CREATE INDEX idx_events_event_date ON events(event_date);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_skill_level ON events(skill_level);
CREATE INDEX idx_events_host_id ON events(host_id);
CREATE INDEX idx_events_short_code ON events(short_code);
```

### 觸發器

```sql
-- 自動生成 short_code
CREATE TRIGGER trigger_set_event_short_code
    BEFORE INSERT ON events
    FOR EACH ROW
    EXECUTE FUNCTION set_event_short_code();

-- 自動更新 updated_at
CREATE TRIGGER trigger_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### 約束

```sql
-- 人數上限 4-20 人
CHECK (capacity >= 4 AND capacity <= 20)

-- 費用 0-9999 元
CHECK (fee >= 0 AND fee <= 9999)

-- 技能等級枚舉
CHECK (skill_level IN ('beginner', 'intermediate', 'advanced', 'expert', 'any'))

-- 狀態枚舉
CHECK (status IN ('open', 'full', 'cancelled', 'completed'))
```

### 範例資料

```sql
INSERT INTO events (
  id, host_id, title, description, event_date, start_time, end_time,
  location_name, location_address, location_point, capacity, skill_level, fee, status
) VALUES (
  '660e8400-e29b-41d4-a716-446655440000',
  '550e8400-e29b-41d4-a716-446655440000',
  '週末輕鬆打',
  '歡迎新手參加，請自備球拍',
  '2026-01-25',
  '19:00:00',
  '21:00:00',
  '大安森林公園網球場',
  '台北市大安區新生南路二段1號',
  ST_SetSRID(ST_MakePoint(121.5367, 25.0292), 4326)::geography,
  8,
  'intermediate',
  200,
  'open'
);
```

---

## 3. registrations (報名表)

儲存使用者活動報名記錄。

### 欄位說明

| 欄位名稱 | 資料型別 | 約束 | 說明 |
|---------|---------|------|------|
| `id` | UUID | PRIMARY KEY | 報名記錄唯一識別碼 |
| `event_id` | UUID | FOREIGN KEY (events.id), NOT NULL | 活動 ID |
| `user_id` | UUID | FOREIGN KEY (users.id), NOT NULL | 使用者 ID |
| `status` | VARCHAR(20) | NOT NULL, DEFAULT 'confirmed' | 報名狀態 |
| `waitlist_position` | SMALLINT | NULLABLE | 候補順位 |
| `registered_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 報名時間 |
| `confirmed_at` | TIMESTAMP WITH TIME ZONE | NULLABLE | 確認時間 |
| `cancelled_at` | TIMESTAMP WITH TIME ZONE | NULLABLE | 取消時間 |

### 枚舉值

#### status (報名狀態)

- `confirmed`: 正取
- `waitlist`: 候補
- `cancelled`: 已取消

### 索引

```sql
CREATE INDEX idx_registrations_event_id ON registrations(event_id);
CREATE INDEX idx_registrations_user_id ON registrations(user_id);
CREATE INDEX idx_registrations_status ON registrations(status);
```

### 唯一約束

```sql
-- 防止重複報名
UNIQUE(event_id, user_id)
```

### 級聯刪除

```sql
-- 當活動被刪除時，相關報名也會被刪除
FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE

-- 當使用者被刪除時，相關報名也會被刪除
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

### 範例資料

```sql
-- 正取報名
INSERT INTO registrations (id, event_id, user_id, status) VALUES
(
  '770e8400-e29b-41d4-a716-446655440000',
  '660e8400-e29b-41d4-a716-446655440000',
  '550e8400-e29b-41d4-a716-446655440000',
  'confirmed'
);

-- 候補報名
INSERT INTO registrations (id, event_id, user_id, status, waitlist_position) VALUES
(
  '880e8400-e29b-41d4-a716-446655440000',
  '660e8400-e29b-41d4-a716-446655440000',
  '990e8400-e29b-41d4-a716-446655440000',
  'waitlist',
  1
);
```

---

## 4. notifications (通知表)

儲存使用者通知。

### 欄位說明

| 欄位名稱 | 資料型別 | 約束 | 說明 |
|---------|---------|------|------|
| `id` | UUID | PRIMARY KEY | 通知唯一識別碼 |
| `user_id` | UUID | FOREIGN KEY (users.id), NOT NULL | 使用者 ID |
| `event_id` | UUID | FOREIGN KEY (events.id), NULLABLE | 關聯活動 ID |
| `type` | VARCHAR(50) | NOT NULL | 通知類型 |
| `title` | VARCHAR(200) | NOT NULL | 通知標題 |
| `message` | TEXT | NULLABLE | 通知訊息 |
| `is_read` | BOOLEAN | DEFAULT FALSE | 是否已讀 |
| `created_at` | TIMESTAMP WITH TIME ZONE | DEFAULT NOW() | 建立時間 |

### 通知類型

| 類型 | 說明 |
|-----|------|
| `waitlist_promoted` | 從候補轉為正取 |
| `event_cancelled` | 活動取消 |
| `event_updated` | 活動更新 |
| `registration_confirmed` | 報名確認 |

### 索引

```sql
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
```

### 級聯行為

```sql
-- 當使用者被刪除時，通知也會被刪除
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

-- 當活動被刪除時，event_id 設為 NULL
FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE SET NULL
```

### 範例資料

```sql
INSERT INTO notifications (
  id, user_id, event_id, type, title, message, is_read
) VALUES (
  'aa0e8400-e29b-41d4-a716-446655440000',
  '550e8400-e29b-41d4-a716-446655440000',
  '660e8400-e29b-41d4-a716-446655440000',
  'waitlist_promoted',
  '您已從候補轉為正取',
  '活動「週末輕鬆打」有人取消報名，您已從候補轉為正取！',
  false
);
```

---

## 5. Views (視圖)

### event_summary

活動摘要視圖，包含報名人數統計。

```sql
CREATE OR REPLACE VIEW event_summary AS
SELECT
    e.id,
    e.host_id,
    e.title,
    e.description,
    e.event_date,
    e.start_time,
    e.end_time,
    e.location_name,
    e.location_address,
    ST_Y(e.location_point::geometry) AS latitude,
    ST_X(e.location_point::geometry) AS longitude,
    e.google_place_id,
    e.capacity,
    e.skill_level,
    e.fee,
    e.status,
    e.short_code,
    e.created_at,
    e.updated_at,
    COALESCE(COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END), 0) AS confirmed_count,
    COALESCE(COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END), 0) AS waitlist_count,
    u.display_name AS host_name,
    u.avatar_url AS host_avatar
FROM events e
LEFT JOIN registrations r ON e.id = r.event_id AND r.status != 'cancelled'
LEFT JOIN users u ON e.host_id = u.id
GROUP BY e.id, u.display_name, u.avatar_url;
```

#### 使用範例

```sql
-- 查詢活動摘要
SELECT * FROM event_summary
WHERE status = 'open'
AND event_date >= CURRENT_DATE
ORDER BY event_date ASC;

-- 查詢特定主辦人的活動
SELECT * FROM event_summary
WHERE host_id = '550e8400-e29b-41d4-a716-446655440000';
```

---

## 6. Functions (函式)

### generate_short_code

生成隨機短網址代碼。

```sql
CREATE OR REPLACE FUNCTION generate_short_code(length INTEGER DEFAULT 6)
RETURNS VARCHAR AS $$
DECLARE
    chars VARCHAR := 'abcdefghijkmnpqrstuvwxyz23456789';
    result VARCHAR := '';
    i INTEGER;
BEGIN
    FOR i IN 1..length LOOP
        result := result || substr(chars, floor(random() * length(chars) + 1)::INTEGER, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;
```

#### 使用範例

```sql
SELECT generate_short_code();     -- 'a3k9m2'
SELECT generate_short_code(8);    -- 'x4p8n6v2'
```

---

### set_event_short_code

自動為活動設定唯一的短網址代碼（觸發器函式）。

```sql
CREATE OR REPLACE FUNCTION set_event_short_code()
RETURNS TRIGGER AS $$
DECLARE
    new_code VARCHAR;
    code_exists BOOLEAN;
BEGIN
    IF NEW.short_code IS NULL THEN
        LOOP
            new_code := generate_short_code(6);
            SELECT EXISTS(SELECT 1 FROM events WHERE short_code = new_code) INTO code_exists;
            EXIT WHEN NOT code_exists;
        END LOOP;
        NEW.short_code := new_code;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

---

### update_updated_at_column

自動更新 updated_at 欄位（觸發器函式）。

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

---

## 7. 常用查詢範例

### 7.1 地理位置搜尋

查詢特定位置附近的活動（使用 PostGIS）。

```sql
-- 查詢台北 101 周圍 5 公里的活動
SELECT
    e.*,
    ST_Distance(
        e.location_point,
        ST_SetSRID(ST_MakePoint(121.5645, 25.0330), 4326)::geography
    ) / 1000 AS distance_km
FROM events e
WHERE
    e.status = 'open'
    AND e.event_date >= CURRENT_DATE
    AND ST_DWithin(
        e.location_point,
        ST_SetSRID(ST_MakePoint(121.5645, 25.0330), 4326)::geography,
        5000  -- 5000 公尺 = 5 公里
    )
ORDER BY distance_km ASC
LIMIT 20;
```

---

### 7.2 報名統計

查詢活動的報名統計。

```sql
-- 查詢特定活動的報名統計
SELECT
    e.id,
    e.title,
    e.capacity,
    COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END) as confirmed_count,
    COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END) as waitlist_count,
    COUNT(CASE WHEN r.status = 'cancelled' THEN 1 END) as cancelled_count
FROM events e
LEFT JOIN registrations r ON e.id = r.event_id
WHERE e.id = '660e8400-e29b-41d4-a716-446655440000'
GROUP BY e.id;
```

---

### 7.3 使用者報名記錄

查詢使用者的所有報名記錄。

```sql
-- 查詢使用者的報名記錄（包含活動資訊）
SELECT
    r.id,
    r.status,
    r.waitlist_position,
    r.registered_at,
    e.id as event_id,
    e.title,
    e.event_date,
    e.start_time,
    e.location_name,
    e.skill_level
FROM registrations r
JOIN events e ON r.event_id = e.id
WHERE
    r.user_id = '550e8400-e29b-41d4-a716-446655440000'
    AND r.status != 'cancelled'
    AND e.event_date >= CURRENT_DATE
ORDER BY e.event_date ASC;
```

---

### 7.4 候補自動提升

當有人取消報名時，提升第一位候補者。

```sql
-- 1. 取消報名
UPDATE registrations
SET status = 'cancelled', cancelled_at = NOW()
WHERE id = '770e8400-e29b-41d4-a716-446655440000';

-- 2. 提升候補
UPDATE registrations
SET
    status = 'confirmed',
    confirmed_at = NOW(),
    waitlist_position = NULL
WHERE id = (
    SELECT id
    FROM registrations
    WHERE event_id = '660e8400-e29b-41d4-a716-446655440000'
        AND status = 'waitlist'
    ORDER BY waitlist_position ASC
    LIMIT 1
)
RETURNING *;
```

---

### 7.5 活動搜尋（多條件）

複合條件搜尋活動。

```sql
SELECT * FROM event_summary
WHERE
    status = 'open'
    AND event_date >= CURRENT_DATE
    AND skill_level IN ('beginner', 'any')  -- 技能等級篩選
    AND fee <= 500                           -- 費用篩選
    AND ST_DWithin(
        ST_SetSRID(ST_MakePoint(121.5645, 25.0330), 4326)::geography,
        location_point,
        10000  -- 10 公里範圍
    )
ORDER BY event_date ASC, created_at DESC
LIMIT 20
OFFSET 0;
```

---

## 8. 資料庫維護

### 8.1 備份

```bash
# 完整備份
pg_dump -h localhost -U postgres picklego > backup_$(date +%Y%m%d).sql

# 僅備份 schema
pg_dump -h localhost -U postgres -s picklego > schema_backup.sql

# 僅備份資料
pg_dump -h localhost -U postgres -a picklego > data_backup.sql
```

### 8.2 還原

```bash
# 還原資料庫
psql -h localhost -U postgres picklego < backup_20260121.sql
```

### 8.3 效能監控

```sql
-- 查看資料表大小
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 查看索引使用情況
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- 查看慢查詢
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

---

## 9. Migration 管理

### 9.1 建立 Migration

```bash
# 使用 golang-migrate
migrate create -ext sql -dir apps/api/migrations -seq add_user_email
```

### 9.2 執行 Migration

```bash
# Up migration
migrate -path apps/api/migrations -database "${DATABASE_URL}" up

# Down migration
migrate -path apps/api/migrations -database "${DATABASE_URL}" down 1

# 查看狀態
migrate -path apps/api/migrations -database "${DATABASE_URL}" version
```

### 9.3 Migration 範例

**000002_add_user_email.up.sql**

```sql
-- 新增 email 驗證欄位
ALTER TABLE users
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- 新增索引
CREATE INDEX idx_users_email_verified ON users(email_verified);
```

**000002_add_user_email.down.sql**

```sql
-- 移除索引
DROP INDEX IF EXISTS idx_users_email_verified;

-- 移除欄位
ALTER TABLE users
DROP COLUMN IF EXISTS email_verified;
```

---

## 10. 安全性考量

### 10.1 敏感資料加密

```sql
-- 使用 pgcrypto 擴展加密敏感資料
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 加密範例（如需儲存敏感資料）
INSERT INTO sensitive_data (user_id, encrypted_value)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    pgp_sym_encrypt('sensitive_data', 'encryption_key')
);

-- 解密範例
SELECT
    user_id,
    pgp_sym_decrypt(encrypted_value::bytea, 'encryption_key') as decrypted_value
FROM sensitive_data;
```

### 10.2 Row Level Security (RLS)

```sql
-- 啟用 RLS（如需實作）
ALTER TABLE events ENABLE ROW LEVEL SECURITY;

-- 建立政策：使用者只能修改自己主辦的活動
CREATE POLICY events_update_policy ON events
    FOR UPDATE
    USING (host_id = current_setting('app.current_user_id')::uuid);
```

### 10.3 SQL Injection 防護

應用層使用 Prepared Statements：

```go
// ✅ 正確：使用 Prepared Statement
db.QueryRow("SELECT * FROM users WHERE id = $1", userID)

// ❌ 錯誤：直接串接 SQL
db.QueryRow(fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID))
```

---

## 11. 效能優化建議

### 11.1 索引優化

```sql
-- 複合索引（針對常用查詢）
CREATE INDEX idx_events_date_status ON events(event_date, status);

-- 部分索引（僅索引需要的資料）
CREATE INDEX idx_events_open ON events(event_date)
WHERE status = 'open';
```

### 11.2 查詢優化

```sql
-- 使用 EXPLAIN ANALYZE 分析查詢計劃
EXPLAIN ANALYZE
SELECT * FROM events
WHERE status = 'open'
AND event_date >= CURRENT_DATE
ORDER BY event_date ASC
LIMIT 20;

-- 使用物化視圖（Materialized View）快取常用查詢
CREATE MATERIALIZED VIEW popular_events AS
SELECT e.*, COUNT(r.id) as registration_count
FROM events e
LEFT JOIN registrations r ON e.id = r.event_id
WHERE e.status = 'open'
GROUP BY e.id
ORDER BY registration_count DESC;

-- 定期刷新
REFRESH MATERIALIZED VIEW popular_events;
```

### 11.3 連線池設定

```go
// 應用層設定連線池
db.SetMaxOpenConns(25)          // 最大開啟連線數
db.SetMaxIdleConns(5)           // 最大閒置連線數
db.SetConnMaxLifetime(5 * time.Minute)  // 連線最大生命週期
```

---

## 12. 監控與告警

### 12.1 重要指標

```sql
-- 連線數監控
SELECT count(*) as connection_count
FROM pg_stat_activity
WHERE datname = 'picklego';

-- 長時間執行的查詢
SELECT
    pid,
    now() - query_start as duration,
    query
FROM pg_stat_activity
WHERE state = 'active'
AND now() - query_start > interval '5 seconds'
ORDER BY duration DESC;

-- 資料表膨脹檢查
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    n_dead_tup,
    n_live_tup,
    round((n_dead_tup::float / NULLIF(n_live_tup, 0)) * 100, 2) as dead_tuple_percent
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

### 12.2 定期維護任務

```sql
-- VACUUM 清理
VACUUM ANALYZE events;
VACUUM ANALYZE registrations;

-- REINDEX 重建索引
REINDEX TABLE events;
```

---

## 13. 疑難排解

### 13.1 常見問題

#### Q1: PostGIS 查詢很慢

```sql
-- 檢查是否有 GIST 索引
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'events'
AND indexdef LIKE '%GIST%';

-- 如果沒有，建立索引
CREATE INDEX idx_events_location ON events USING GIST(location_point);

-- 更新統計資訊
ANALYZE events;
```

#### Q2: 連線數過多

```sql
-- 查看目前連線
SELECT
    client_addr,
    count(*) as connection_count
FROM pg_stat_activity
WHERE datname = 'picklego'
GROUP BY client_addr
ORDER BY connection_count DESC;

-- 中斷閒置連線
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'picklego'
AND state = 'idle'
AND state_change < now() - interval '10 minutes';
```

#### Q3: 查詢效能問題

```sql
-- 啟用慢查詢日誌
ALTER DATABASE picklego SET log_min_duration_statement = 1000; -- 1秒

-- 查看查詢統計（需要 pg_stat_statements 擴展）
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

SELECT
    query,
    calls,
    total_time,
    mean_time,
    min_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 20;
```

---

## 14. ER Diagram (實體關係圖)

```
┌──────────────┐
│    users     │
├──────────────┤
│ id (PK)      │
│ line_user_id │
│ display_name │
│ avatar_url   │
│ email        │
│ created_at   │
│ updated_at   │
└──────┬───────┘
       │
       │ 1:N (主辦人)
       ├─────────────────────────────┐
       │                             │
       │ 1:N (報名者)                │
┌──────▼───────┐              ┌─────▼──────────┐
│    events    │              │ registrations  │
├──────────────┤              ├────────────────┤
│ id (PK)      │ 1:N          │ id (PK)        │
│ host_id (FK) ├──────────────┤ event_id (FK)  │
│ title        │              │ user_id (FK)   │
│ description  │              │ status         │
│ event_date   │              │ waitlist_pos   │
│ start_time   │              │ registered_at  │
│ location_*   │              │ confirmed_at   │
│ capacity     │              │ cancelled_at   │
│ skill_level  │              └────────────────┘
│ fee          │
│ status       │
│ short_code   │         ┌─────────────────┐
│ created_at   │         │ notifications   │
│ updated_at   │  1:N    ├─────────────────┤
└──────┬───────┘◄────────┤ id (PK)         │
       │                 │ user_id (FK)    │
       │                 │ event_id (FK)   │
       │                 │ type            │
       │                 │ title           │
       └─────────────────┤ message         │
                         │ is_read         │
                         │ created_at      │
                         └─────────────────┘
```

---

## 15. 資源與參考

- [PostgreSQL 官方文件](https://www.postgresql.org/docs/)
- [PostGIS 文件](https://postgis.net/documentation/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [pgcrypto 擴展](https://www.postgresql.org/docs/current/pgcrypto.html)

---

**版本**: 1.0.0
**最後更新**: 2026-01-21
**Migration 版本**: 000001
