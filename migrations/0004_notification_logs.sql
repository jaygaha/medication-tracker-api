-- ==============================================================================
-- NOTIFICATION LOGS - DATABASE SCHEMA (PostgreSQL)
-- ==============================================================================

-- 1. Create the notification_logs table
CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID REFERENCES schedules(id) ON DELETE SET NULL,
    schedule_time_id UUID REFERENCES schedule_times(id) ON DELETE SET NULL,
    scheduled_date DATE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_schedule_time_date UNIQUE(schedule_time_id, scheduled_date)
);

-- 2. Index for faster logging and querying
CREATE INDEX IF NOT EXISTS idx_notification_logs_user_date ON notification_logs(user_id, sent_at DESC);
