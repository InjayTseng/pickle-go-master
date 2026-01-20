-- Pickle Go Initial Schema Migration
-- Version: 000001
-- Description: Create initial tables for users, events, registrations, and notifications

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Users Table
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    line_user_id    VARCHAR(64) UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    avatar_url      TEXT,
    email           VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for Line user ID lookup
CREATE INDEX IF NOT EXISTS idx_users_line_user_id ON users(line_user_id);

-- ============================================
-- Events Table
-- ============================================
CREATE TABLE IF NOT EXISTS events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Basic information
    title               VARCHAR(200),
    description         TEXT,

    -- Time
    event_date          DATE NOT NULL,
    start_time          TIME NOT NULL,
    end_time            TIME,

    -- Location (PostGIS)
    location_name       VARCHAR(200) NOT NULL,
    location_address    VARCHAR(500),
    location_point      GEOGRAPHY(POINT, 4326) NOT NULL,
    google_place_id     VARCHAR(255),

    -- Event settings
    capacity            SMALLINT NOT NULL CHECK (capacity >= 4 AND capacity <= 20),
    skill_level         VARCHAR(20) NOT NULL CHECK (skill_level IN ('beginner', 'intermediate', 'advanced', 'expert', 'any')),
    fee                 INTEGER DEFAULT 0 CHECK (fee >= 0 AND fee <= 9999),

    -- Status
    status              VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'full', 'cancelled', 'completed')),

    -- Short URL code for sharing
    short_code          VARCHAR(10) UNIQUE,

    -- Timestamps
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Spatial index for geographic queries
CREATE INDEX IF NOT EXISTS idx_events_location ON events USING GIST(location_point);

-- Other indexes
CREATE INDEX IF NOT EXISTS idx_events_event_date ON events(event_date);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_skill_level ON events(skill_level);
CREATE INDEX IF NOT EXISTS idx_events_host_id ON events(host_id);
CREATE INDEX IF NOT EXISTS idx_events_short_code ON events(short_code);

-- ============================================
-- Registrations Table
-- ============================================
CREATE TABLE IF NOT EXISTS registrations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id            UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Registration status
    status              VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'waitlist', 'cancelled')),
    waitlist_position   SMALLINT,

    -- Timestamps
    registered_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_at        TIMESTAMP WITH TIME ZONE,
    cancelled_at        TIMESTAMP WITH TIME ZONE,

    -- Unique constraint: prevent duplicate registrations
    UNIQUE(event_id, user_id)
);

-- Indexes for registrations
CREATE INDEX IF NOT EXISTS idx_registrations_event_id ON registrations(event_id);
CREATE INDEX IF NOT EXISTS idx_registrations_user_id ON registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_registrations_status ON registrations(status);

-- ============================================
-- Notifications Table
-- ============================================
CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id        UUID REFERENCES events(id) ON DELETE SET NULL,

    type            VARCHAR(50) NOT NULL,
    title           VARCHAR(200) NOT NULL,
    message         TEXT,

    is_read         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- ============================================
-- Views
-- ============================================

-- Event summary view with registration counts
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

-- ============================================
-- Functions
-- ============================================

-- Function to generate a short code for events
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

-- Trigger to auto-generate short code on event creation
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

CREATE TRIGGER trigger_set_event_short_code
    BEFORE INSERT ON events
    FOR EACH ROW
    EXECUTE FUNCTION set_event_short_code();

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
