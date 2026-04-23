-- ==============================================================================
-- MEDICATIONS TRACKER - DATABASE SCHEMA (PostgreSQL)
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- ENUMS (Custom Data Types)
-- ------------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'medication_form') THEN
        CREATE TYPE medication_form AS ENUM (
            'tablet', 'capsule', 'liquid', 'topical', 'injection', 
            'drops', 'inhaler', 'powder', 'device', 'other'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'medication_status') THEN
        CREATE TYPE medication_status AS ENUM (
            'active', 'archived', 'discontinued'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'frequency_type') THEN
        CREATE TYPE frequency_type AS ENUM (
            'every_day', 'regular_intervals', 'specific_days', 'as_needed'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'log_status') THEN
        CREATE TYPE log_status AS ENUM (
            'taken', 'skipped'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'interaction_severity') THEN
        CREATE TYPE interaction_severity AS ENUM (
            'minor', 'moderate', 'severe', 'critical'
        );
    END IF;
END
$$;


-- ------------------------------------------------------------------------------
-- 1. USERS
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255),
    password_hash VARCHAR(255),
    timezone VARCHAR(50) DEFAULT 'UTC', -- Crucial for calculating medication schedules
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- ------------------------------------------------------------------------------
-- 2. MEDICATIONS
-- Core table storing the clinical and general details of the drug.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS medications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,            -- e.g., "Ibuprofen", "Amoxicillin"
    form medication_form NOT NULL,         -- e.g., 'tablet', 'liquid'
    strength_value DECIMAL(10,2),          -- e.g., 200, 500
    strength_unit VARCHAR(20),             -- e.g., 'mg', 'ml', 'mcg'
    rx_number VARCHAR(100),                -- Prescription number (optional)
    notes TEXT,                            -- User-added context ("Take with food")
    status medication_status DEFAULT 'active',
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_medications_user_status ON medications(user_id, status);


-- ------------------------------------------------------------------------------
-- 3. MEDICATION VISUALS
-- Distinct flow for picking shape, primary color, and background.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS medication_visuals (
    medication_id UUID PRIMARY KEY REFERENCES medications(id) ON DELETE CASCADE,
    shape VARCHAR(50),
    primary_color VARCHAR(50),
    secondary_color VARCHAR(50),
    background_color VARCHAR(50),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN medication_visuals.shape IS 'Shape of the medication';
COMMENT ON COLUMN medication_visuals.primary_color IS 'Primary color of the medication';
COMMENT ON COLUMN medication_visuals.secondary_color IS 'Secondary color of the medication';
COMMENT ON COLUMN medication_visuals.background_color IS 'Background color of the medication';


-- ------------------------------------------------------------------------------
-- 4. SCHEDULES
-- Defines *how often* the medication should be taken.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medication_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    type frequency_type NOT NULL,
    interval_days INTEGER,                 -- e.g., 3 (for "every 3 days"). Null if daily/as-needed
    start_date DATE NOT NULL,
    end_date DATE,                         -- Null if ongoing indefinitely
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- ------------------------------------------------------------------------------
-- 5. SCHEDULE DAYS (For "Specific Days of the Week" schedules)
-- Used when frequency_type is 'specific_days' (e.g., Mondays and Wednesdays).
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schedule_days (
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    day_of_week INTEGER CHECK (day_of_week BETWEEN 1 AND 7), -- 1=Monday, 7=Sunday
    PRIMARY KEY (schedule_id, day_of_week)
);


-- ------------------------------------------------------------------------------
-- 6. SCHEDULE TIMES
-- Defines *what time of day* and *how much* to take. 
-- A single schedule can have multiple times (e.g., 8:00 AM and 8:00 PM).
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schedule_times (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    time_of_day TIME NOT NULL,             -- e.g., '08:00:00'
    dose_amount DECIMAL(10,2) NOT NULL,    -- e.g., 1 (pill), 15 (ml)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- ------------------------------------------------------------------------------
-- 7. MEDICATION LOGS
-- The actual history of the user taking or skipping their medication.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS medication_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    medication_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    schedule_time_id UUID REFERENCES schedule_times(id) ON DELETE SET NULL, -- Null if "As Needed"
    
    status log_status NOT NULL,            -- 'taken' or 'skipped'
    dose_taken DECIMAL(10,2),              -- How much was actually taken
    
    scheduled_timestamp TIMESTAMP WITH TIME ZONE, -- When it was supposed to be taken
    actual_timestamp TIMESTAMP WITH TIME ZONE NOT NULL, -- When the user hit "Log"
    
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Crucial indexes for querying a user's log history efficiently (e.g., for charts/calendars)
CREATE INDEX IF NOT EXISTS idx_logs_user_date ON medication_logs(user_id, actual_timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_medication ON medication_logs(medication_id);


-- ------------------------------------------------------------------------------
-- 8. DRUG INTERACTIONS (Reference Table)
-- In a real app, this data usually comes from an external clinical API (like RxNorm),
-- but we store identified interactions locally for offline warnings.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS drug_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    medication_1_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    medication_2_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    severity interaction_severity NOT NULL,
    description TEXT,                      -- e.g., "May cause severe drowsiness"
    acknowledged BOOLEAN DEFAULT FALSE,    -- Whether the user dismissed the warning
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (medication_1_id, medication_2_id)
);