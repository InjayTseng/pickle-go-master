-- Pickle Go Index Optimization Migration
-- Version: 000002
-- Description: Add composite and partial indexes for optimal query performance
--
-- This migration adds indexes based on analysis of repository query patterns:
-- - event_repo.go: geo-spatial queries, status/date filtering
-- - registration_repo.go: event-status lookups, user registration queries
-- - user_repo.go: line_user_id lookups (already indexed)
--
-- Note: For production deployments with large tables, consider running these
-- CREATE INDEX statements with CONCURRENTLY outside of a transaction to avoid
-- locking issues. The standard CREATE INDEX is used here for migration compatibility.

-- ============================================
-- Registrations Table Indexes
-- ============================================

-- Composite index for counting confirmed/waitlist by event
-- Used in: CountConfirmed, GetRegistrationStats, FindByEventID, FindWithUsersByEventID
-- Queries like: SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'
CREATE INDEX IF NOT EXISTS idx_registrations_event_status
    ON registrations(event_id, status);

-- Composite index for user registration lookups by status
-- Used in: FindByUserID, FindEventsByUserID
-- Queries like: SELECT * FROM registrations WHERE user_id = $1 AND status != 'cancelled'
CREATE INDEX IF NOT EXISTS idx_registrations_user_status
    ON registrations(user_id, status);

-- Index for waitlist ordering within an event
-- Used in: GetFirstWaitlist, GetNextWaitlistPosition, PromoteFromWaitlist
-- Queries like: SELECT * FROM registrations WHERE event_id = $1 AND status = 'waitlist' ORDER BY waitlist_position
CREATE INDEX IF NOT EXISTS idx_registrations_event_waitlist
    ON registrations(event_id, waitlist_position)
    WHERE status = 'waitlist';

-- Covering index for registration lookup by event and user
-- Used in: FindByEventAndUser, HasUserRegistered, RegisterWithLock
-- Note: The UNIQUE constraint creates an implicit index, but this partial index
-- excludes cancelled registrations for faster active registration checks
CREATE INDEX IF NOT EXISTS idx_registrations_event_user_active
    ON registrations(event_id, user_id)
    WHERE status != 'cancelled';

-- ============================================
-- Events Table Indexes
-- ============================================

-- Composite index for date-based queries with status filtering
-- Used in: FindNearby, FindUpcoming
-- Queries filter by event_date >= CURRENT_DATE AND status IN ('open', 'full')
CREATE INDEX IF NOT EXISTS idx_events_date_status
    ON events(event_date, status);

-- Partial index for active events (open or full, future dates)
-- This significantly speeds up the most common query pattern: finding bookable events
-- Used in: FindNearby, FindUpcoming
-- Note: The condition event_date >= CURRENT_DATE is evaluated at index creation time,
-- so this index may need to be rebuilt periodically for optimal performance on date filtering.
CREATE INDEX IF NOT EXISTS idx_events_active
    ON events(event_date, start_time)
    WHERE status IN ('open', 'full');

-- Composite index for geo-spatial queries with status filter
-- The GIST index on location_point handles spatial queries, but filtering by status
-- is done post-index scan. This B-tree index helps when status selectivity is high.
-- Used in: FindNearby
CREATE INDEX IF NOT EXISTS idx_events_status_date
    ON events(status, event_date)
    WHERE status IN ('open', 'full');

-- ============================================
-- Notifications Table Indexes
-- ============================================

-- Composite index for fetching unread notifications for a user
-- Used in: GetUnreadNotifications (typical notification query pattern)
-- Queries like: SELECT * FROM notifications WHERE user_id = $1 AND is_read = false ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id, created_at DESC)
    WHERE is_read = false;

-- Composite index for notification queries with event association
-- Used when looking up notifications related to specific events
CREATE INDEX IF NOT EXISTS idx_notifications_event_user
    ON notifications(event_id, user_id)
    WHERE event_id IS NOT NULL;

-- ============================================
-- Analysis Statistics Update
-- ============================================

-- Update statistics for the query planner after adding indexes
ANALYZE registrations;
ANALYZE events;
ANALYZE notifications;
