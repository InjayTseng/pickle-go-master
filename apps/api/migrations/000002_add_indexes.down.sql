-- Pickle Go Index Optimization Rollback
-- Version: 000002
-- Description: Remove composite and partial indexes added for query optimization

-- ============================================
-- Drop Notifications Indexes
-- ============================================
DROP INDEX IF EXISTS idx_notifications_event_user;
DROP INDEX IF EXISTS idx_notifications_user_unread;

-- ============================================
-- Drop Events Indexes
-- ============================================
DROP INDEX IF EXISTS idx_events_status_date;
DROP INDEX IF EXISTS idx_events_active;
DROP INDEX IF EXISTS idx_events_date_status;

-- ============================================
-- Drop Registrations Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registrations_event_user_active;
DROP INDEX IF EXISTS idx_registrations_event_waitlist;
DROP INDEX IF EXISTS idx_registrations_user_status;
DROP INDEX IF EXISTS idx_registrations_event_status;

-- ============================================
-- Update Statistics After Removing Indexes
-- ============================================
ANALYZE registrations;
ANALYZE events;
ANALYZE notifications;
