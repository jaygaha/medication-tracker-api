-- ==============================================================================
-- NOTIFICATIONS AND DEVICE TOKENS - DATABASE SCHEMA (PostgreSQL)
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- ENUMS (Custom Data Types)
-- ------------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_preference') THEN
        CREATE TYPE notification_preference AS ENUM (
            'none', 'email', 'push', 'all'
        );
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'device_platform') THEN
        CREATE TYPE device_platform AS ENUM (
            'android', 'ios', 'web'
        );
    END IF;
END
$$;

-- 1. Alter table users to add notification preferences
ALTER TABLE users ADD COLUMN IF NOT EXISTS notification_preference notification_preference DEFAULT 'all';

-- ------------------------------------------------------------------------------
-- 2. DEVICE TOKENS
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform device_platform NOT NULL,
    last_used TIMESTAMP WITH TIME ZONE,
    is_active   BOOLEAN DEFAULT TRUE,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens(user_id, is_active);