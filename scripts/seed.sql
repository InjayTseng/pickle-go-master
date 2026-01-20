-- ===========================================
-- Pickle Go - 種子資料 (Seed Data)
-- ===========================================
-- 用於 Closed Beta 測試的初始資料
-- 執行方式: psql -d picklego -f scripts/seed.sql

-- 清空現有測試資料（小心使用！）
-- DELETE FROM notifications;
-- DELETE FROM registrations;
-- DELETE FROM events;
-- DELETE FROM users WHERE display_name LIKE '測試%';

-- ===========================================
-- 測試用戶
-- ===========================================
INSERT INTO users (id, line_user_id, display_name, avatar_url, email, created_at) VALUES
  -- 主辦人帳號
  ('11111111-1111-1111-1111-111111111111', 'U001_test_host_alan', '阿倫教練', 'https://api.dicebear.com/7.x/avataaars/svg?seed=alan', 'alan@test.com', NOW() - INTERVAL '30 days'),
  ('22222222-2222-2222-2222-222222222222', 'U002_test_host_betty', '貝蒂', 'https://api.dicebear.com/7.x/avataaars/svg?seed=betty', 'betty@test.com', NOW() - INTERVAL '25 days'),
  ('33333333-3333-3333-3333-333333333333', 'U003_test_host_charlie', '查理', 'https://api.dicebear.com/7.x/avataaars/svg?seed=charlie', 'charlie@test.com', NOW() - INTERVAL '20 days'),

  -- 一般參與者帳號
  ('44444444-4444-4444-4444-444444444444', 'U004_test_player_david', '大衛', 'https://api.dicebear.com/7.x/avataaars/svg?seed=david', 'david@test.com', NOW() - INTERVAL '15 days'),
  ('55555555-5555-5555-5555-555555555555', 'U005_test_player_emma', '艾瑪', 'https://api.dicebear.com/7.x/avataaars/svg?seed=emma', 'emma@test.com', NOW() - INTERVAL '14 days'),
  ('66666666-6666-6666-6666-666666666666', 'U006_test_player_frank', '小法', 'https://api.dicebear.com/7.x/avataaars/svg?seed=frank', 'frank@test.com', NOW() - INTERVAL '13 days'),
  ('77777777-7777-7777-7777-777777777777', 'U007_test_player_grace', '葛瑞絲', 'https://api.dicebear.com/7.x/avataaars/svg?seed=grace', 'grace@test.com', NOW() - INTERVAL '12 days'),
  ('88888888-8888-8888-8888-888888888888', 'U008_test_player_henry', '小亨利', 'https://api.dicebear.com/7.x/avataaars/svg?seed=henry', 'henry@test.com', NOW() - INTERVAL '11 days'),
  ('99999999-9999-9999-9999-999999999999', 'U009_test_player_ivy', '小艾', 'https://api.dicebear.com/7.x/avataaars/svg?seed=ivy', 'ivy@test.com', NOW() - INTERVAL '10 days'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'U010_test_player_jack', '傑克', 'https://api.dicebear.com/7.x/avataaars/svg?seed=jack', 'jack@test.com', NOW() - INTERVAL '9 days')
ON CONFLICT (line_user_id) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  avatar_url = EXCLUDED.avatar_url;

-- ===========================================
-- 測試活動 - 台北地區
-- ===========================================

-- 活動 1: 大安運動中心 - 今天
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e1111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111',
   '週末輕鬆打', '歡迎新手一起來玩！會有人教基本動作。',
   CURRENT_DATE, '14:00', '16:00',
   '大安運動中心', '台北市大安區辛亥路三段55號',
   ST_SetSRID(ST_MakePoint(121.5434, 25.0181), 4326)::geography,
   'ChIJN1t_tDeuEmsRUsoyG83frY4', 8, 'beginner', 100, 'open', 'daan01',
   NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  event_date = EXCLUDED.event_date;

-- 活動 2: 南港運動中心 - 明天
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e2222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222',
   '進階切磋場', '中高程度以上，歡迎來挑戰！',
   CURRENT_DATE + INTERVAL '1 day', '10:00', '12:00',
   '南港運動中心', '台北市南港區玉成街69號',
   ST_SetSRID(ST_MakePoint(121.5857, 25.0454), 4326)::geography,
   'ChIJy6EmpT2rQjQRIXm2ZVfb0bo', 8, 'advanced', 150, 'open', 'nkang1',
   NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  event_date = EXCLUDED.event_date;

-- 活動 3: 中山運動中心 - 後天
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e3333333-3333-3333-3333-333333333333', '33333333-3333-3333-3333-333333333333',
   '混合雙打練習', '不限程度，以練習雙打配合為主',
   CURRENT_DATE + INTERVAL '2 days', '19:00', '21:00',
   '中山運動中心', '台北市中山區中山北路二段44巷2號',
   ST_SetSRID(ST_MakePoint(121.5229, 25.0561), 4326)::geography,
   'ChIJk2_F3PupQjQR7r3jSfX4S5M', 12, 'any', 80, 'open', 'zsport',
   NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  event_date = EXCLUDED.event_date;

-- 活動 4: 士林運動中心 - 本週六
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e4444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111',
   '假日特訓班', '阿倫教練親自指導，名額有限！',
   CURRENT_DATE + INTERVAL '3 days', '09:00', '12:00',
   '士林運動中心', '台北市士林區士商路1號',
   ST_SetSRID(ST_MakePoint(121.5245, 25.0887), 4326)::geography,
   'ChIJgxyBX1GpQjQRGQxJh2Ixxvw', 6, 'intermediate', 300, 'open', 'sltran',
   NOW() - INTERVAL '6 hours')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  event_date = EXCLUDED.event_date;

-- 活動 5: 內湖運動中心 - 本週日
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e5555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222',
   '新手歡樂場', '完全不會也沒關係，會從頭教起',
   CURRENT_DATE + INTERVAL '4 days', '15:00', '17:00',
   '內湖運動中心', '台北市內湖區洲子街12號',
   ST_SetSRID(ST_MakePoint(121.5742, 25.0778), 4326)::geography,
   'ChIJV4jeFBWsQjQRAQTFhW6hzME', 10, 'beginner', 50, 'open', 'nhfun1',
   NOW() - INTERVAL '3 hours')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  event_date = EXCLUDED.event_date;

-- 活動 6: 已額滿的活動
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e6666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333',
   '熱門場次', '這場已經額滿了！',
   CURRENT_DATE + INTERVAL '5 days', '18:00', '20:00',
   '信義運動中心', '台北市信義區松勤街100號',
   ST_SetSRID(ST_MakePoint(121.5654, 25.0304), 4326)::geography,
   'ChIJBwn8v8CrQjQRPYvEn6WFiZM', 4, 'intermediate', 100, 'full', 'xyfull',
   NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  status = EXCLUDED.status;

-- 活動 7: 已取消的活動
INSERT INTO events (id, host_id, title, description, event_date, start_time, end_time, location_name, location_address, location_point, google_place_id, capacity, skill_level, fee, status, short_code, created_at) VALUES
  ('e7777777-7777-7777-7777-777777777777', '11111111-1111-1111-1111-111111111111',
   '已取消的活動', '因故取消',
   CURRENT_DATE + INTERVAL '6 days', '20:00', '22:00',
   '文山運動中心', '台北市文山區興隆路三段222號',
   ST_SetSRID(ST_MakePoint(121.5526, 24.9913), 4326)::geography,
   'ChIJe7gk2FCqQjQRnfFlG8Lnzs0', 8, 'any', 0, 'cancelled', 'canxxl',
   NOW() - INTERVAL '7 days')
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  status = EXCLUDED.status;

-- ===========================================
-- 報名記錄
-- ===========================================

-- 活動 1 的報名 (5 人已報名)
INSERT INTO registrations (id, event_id, user_id, status, registered_at, confirmed_at) VALUES
  ('r1111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', '44444444-4444-4444-4444-444444444444', 'confirmed', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
  ('r1111111-1111-1111-1111-222222222222', 'e1111111-1111-1111-1111-111111111111', '55555555-5555-5555-5555-555555555555', 'confirmed', NOW() - INTERVAL '20 hours', NOW() - INTERVAL '20 hours'),
  ('r1111111-1111-1111-1111-333333333333', 'e1111111-1111-1111-1111-111111111111', '66666666-6666-6666-6666-666666666666', 'confirmed', NOW() - INTERVAL '18 hours', NOW() - INTERVAL '18 hours'),
  ('r1111111-1111-1111-1111-444444444444', 'e1111111-1111-1111-1111-111111111111', '77777777-7777-7777-7777-777777777777', 'confirmed', NOW() - INTERVAL '15 hours', NOW() - INTERVAL '15 hours'),
  ('r1111111-1111-1111-1111-555555555555', 'e1111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888888', 'confirmed', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours')
ON CONFLICT (event_id, user_id) DO NOTHING;

-- 活動 2 的報名 (3 人已報名)
INSERT INTO registrations (id, event_id, user_id, status, registered_at, confirmed_at) VALUES
  ('r2222222-2222-2222-2222-111111111111', 'e2222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444', 'confirmed', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '10 hours'),
  ('r2222222-2222-2222-2222-222222222222', 'e2222222-2222-2222-2222-222222222222', '77777777-7777-7777-7777-777777777777', 'confirmed', NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours'),
  ('r2222222-2222-2222-2222-333333333333', 'e2222222-2222-2222-2222-222222222222', '99999999-9999-9999-9999-999999999999', 'confirmed', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours')
ON CONFLICT (event_id, user_id) DO NOTHING;

-- 活動 3 的報名 (6 人已報名)
INSERT INTO registrations (id, event_id, user_id, status, registered_at, confirmed_at) VALUES
  ('r3333333-3333-3333-3333-111111111111', 'e3333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444', 'confirmed', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours'),
  ('r3333333-3333-3333-3333-222222222222', 'e3333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555', 'confirmed', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours'),
  ('r3333333-3333-3333-3333-333333333333', 'e3333333-3333-3333-3333-333333333333', '66666666-6666-6666-6666-666666666666', 'confirmed', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),
  ('r3333333-3333-3333-3333-444444444444', 'e3333333-3333-3333-3333-333333333333', '77777777-7777-7777-7777-777777777777', 'confirmed', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
  ('r3333333-3333-3333-3333-555555555555', 'e3333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888888888', 'confirmed', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour'),
  ('r3333333-3333-3333-3333-666666666666', 'e3333333-3333-3333-3333-333333333333', '99999999-9999-9999-9999-999999999999', 'confirmed', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes')
ON CONFLICT (event_id, user_id) DO NOTHING;

-- 活動 6 的報名 (已額滿 - 4 人 + 2 候補)
INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at, confirmed_at) VALUES
  ('r6666666-6666-6666-6666-111111111111', 'e6666666-6666-6666-6666-666666666666', '44444444-4444-4444-4444-444444444444', 'confirmed', NULL, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
  ('r6666666-6666-6666-6666-222222222222', 'e6666666-6666-6666-6666-666666666666', '55555555-5555-5555-5555-555555555555', 'confirmed', NULL, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
  ('r6666666-6666-6666-6666-333333333333', 'e6666666-6666-6666-6666-666666666666', '66666666-6666-6666-6666-666666666666', 'confirmed', NULL, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('r6666666-6666-6666-6666-444444444444', 'e6666666-6666-6666-6666-666666666666', '77777777-7777-7777-7777-777777777777', 'confirmed', NULL, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('r6666666-6666-6666-6666-555555555555', 'e6666666-6666-6666-6666-666666666666', '88888888-8888-8888-8888-888888888888', 'waitlist', 1, NOW() - INTERVAL '2 days', NULL),
  ('r6666666-6666-6666-6666-666666666666', 'e6666666-6666-6666-6666-666666666666', '99999999-9999-9999-9999-999999999999', 'waitlist', 2, NOW() - INTERVAL '1 day', NULL)
ON CONFLICT (event_id, user_id) DO NOTHING;

-- ===========================================
-- 通知記錄
-- ===========================================
INSERT INTO notifications (id, user_id, event_id, type, title, message, is_read, created_at) VALUES
  -- 報名成功通知
  ('n1111111-1111-1111-1111-111111111111', '44444444-4444-4444-4444-444444444444', 'e1111111-1111-1111-1111-111111111111',
   'registration_confirmed', '報名成功', '您已成功報名「週末輕鬆打」活動！', false, NOW() - INTERVAL '1 day'),

  -- 活動提醒
  ('n2222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444', 'e1111111-1111-1111-1111-111111111111',
   'event_reminder', '活動提醒', '您報名的「週末輕鬆打」將於今天 14:00 開始！', false, NOW() - INTERVAL '2 hours'),

  -- 候補通知
  ('n3333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888888888', 'e6666666-6666-6666-6666-666666666666',
   'waitlist_joined', '加入候補', '您已加入「熱門場次」的候補名單，目前排第 1 位', true, NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- ===========================================
-- 驗證資料
-- ===========================================
DO $$
BEGIN
  RAISE NOTICE '====== Seed Data Summary ======';
  RAISE NOTICE 'Users: %', (SELECT COUNT(*) FROM users WHERE display_name LIKE '測試%' OR display_name IN ('阿倫教練', '貝蒂', '查理', '大衛', '艾瑪', '小法', '葛瑞絲', '小亨利', '小艾', '傑克'));
  RAISE NOTICE 'Events: %', (SELECT COUNT(*) FROM events WHERE short_code IN ('daan01', 'nkang1', 'zsport', 'sltran', 'nhfun1', 'xyfull', 'canxxl'));
  RAISE NOTICE 'Registrations: %', (SELECT COUNT(*) FROM registrations WHERE event_id IN (
    'e1111111-1111-1111-1111-111111111111', 'e2222222-2222-2222-2222-222222222222',
    'e3333333-3333-3333-3333-333333333333', 'e6666666-6666-6666-6666-666666666666'
  ));
  RAISE NOTICE 'Notifications: %', (SELECT COUNT(*) FROM notifications WHERE id IN (
    'n1111111-1111-1111-1111-111111111111', 'n2222222-2222-2222-2222-222222222222', 'n3333333-3333-3333-3333-333333333333'
  ));
  RAISE NOTICE '================================';
END $$;
