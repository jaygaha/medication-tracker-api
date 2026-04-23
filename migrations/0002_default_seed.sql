-- ==============================================================================
-- MEDICATIONS TRACKER - DEFAULT SEED DATA
-- ==============================================================================

-- Insert a Sample User
INSERT INTO users (id, first_name, last_name, email, timezone)
VALUES ('11111111-1111-1111-1111-111111111111', 'Jay', 'Gaha', 'jaygaha@gmail.com', 'Asia/Tokyo')
ON CONFLICT DO NOTHING;

-- Insert Medications (Ibuprofen & Amoxicillin)
INSERT INTO medications (id, user_id, name, form, strength_value, strength_unit, notes, status)
VALUES
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Ibuprofen', 'tablet', 200.00, 'mg', 'Take with food to prevent upset stomach.', 'active'),
('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Amoxicillin', 'capsule', 500.00, 'mg', 'Finish entire prescription course.', 'active')
ON CONFLICT DO NOTHING;

-- Insert Medication Visuals
INSERT INTO medication_visuals (medication_id, shape, primary_color, secondary_color, background_color)
VALUES
('22222222-2222-2222-2222-222222222222', 'round', '#FFFFFF', NULL, '#E5E5EA'), -- White pill on gray background
('33333333-3333-3333-3333-333333333333', 'capsule', '#FF3B30', '#FFCC00', '#FFD6D6') -- Red/Yellow capsule
ON CONFLICT DO NOTHING;

-- Insert Schedules
-- Ibuprofen: As needed (No end date)
-- Amoxicillin: Every day for 10 days
INSERT INTO schedules (id, medication_id, type, interval_days, start_date, end_date)
VALUES
('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 'as_needed', NULL, CURRENT_DATE, NULL),
('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', 'every_day', NULL, CURRENT_DATE, CURRENT_DATE + INTERVAL '10 days')
ON CONFLICT DO NOTHING;

-- Insert Schedule Times (Only applies to Amoxicillin since Ibuprofen is 'as_needed')
-- Taking Amoxicillin at 8:00 AM and 8:00 PM
INSERT INTO schedule_times (id, schedule_id, time_of_day, dose_amount)
VALUES
('55555555-5555-5555-5555-555555555555', '66666666-6666-6666-6666-666666666666', '08:00:00', 1.00),
('77777777-7777-7777-7777-777777777777', '66666666-6666-6666-6666-666666666666', '20:00:00', 1.00)
ON CONFLICT DO NOTHING;

-- Insert Medication Logs (Simulating taking a dose today)
INSERT INTO medication_logs (user_id, medication_id, schedule_time_id, status, dose_taken, scheduled_timestamp, actual_timestamp, notes)
VALUES
('11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555', 'taken', 1.00, CURRENT_DATE + TIME '08:00:00', CURRENT_TIMESTAMP, 'Took with breakfast.')
ON CONFLICT DO NOTHING;

-- Insert Drug Interactions
INSERT INTO drug_interactions (user_id, medication_1_id, medication_2_id, severity, description, acknowledged)
VALUES
('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', 'minor', 'No significant interaction between Ibuprofen and Amoxicillin.', true)
ON CONFLICT DO NOTHING;