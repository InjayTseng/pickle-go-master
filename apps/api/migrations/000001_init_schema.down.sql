-- Pickle Go Initial Schema Rollback
-- Version: 000001
-- Description: Drop all tables, views, functions, and triggers

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_events_updated_at ON events;
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP TRIGGER IF EXISTS trigger_set_event_short_code ON events;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS set_event_short_code();
DROP FUNCTION IF EXISTS generate_short_code(INTEGER);

-- Drop views
DROP VIEW IF EXISTS event_summary;

-- Drop tables (in order of dependencies)
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS registrations;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;

-- Note: We don't drop the PostGIS extension as it might be used by other databases
-- DROP EXTENSION IF EXISTS postgis;
-- DROP EXTENSION IF EXISTS "uuid-ossp";
