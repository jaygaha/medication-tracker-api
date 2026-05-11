// internal/repository/schedule_repo.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

// ScheduleRepository defines the interface for schedule data access
type ScheduleRepository interface {
	CreateSchedule(ctx context.Context, schedule *models.Schedule) error
	GetActiveByMedicationID(ctx context.Context, medicationID string) (*models.Schedule, error)
	GetSchedulesByMedicationID(ctx context.Context, medicationID string) (*models.Schedule, error)
	GetAllActiveSchedules(ctx context.Context) ([]*models.Schedule, error)
	UpdateSchedule(ctx context.Context, schedule *models.Schedule) error
	DeleteSchedule(ctx context.Context, medicationID, scheduleID string) error
}

// scheduleRepository implements ScheduleRepository
type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

// CreateSchedule inserts a new medication schedule along with its times and days
func (r *scheduleRepository) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("failed to start transaction", err)
	}

	// 1. Insert Core Schedule
	query := `
		INSERT INTO schedules (
			id, medication_id, type, interval_days, start_date, end_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query,
		schedule.ID,
		schedule.MedicationID,
		string(schedule.Type),
		schedule.IntervalDays,
		schedule.StartDate,
		schedule.EndDate,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to insert schedule", err)
	}

	// 2. Insert Schedule Times
	timeQuery := `
		INSERT INTO schedule_times (
			id, schedule_id, time_of_day, dose_amount, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`
	for _, t := range schedule.Times {
		_, err = tx.ExecContext(ctx, timeQuery,
			t.ID,
			schedule.ID,
			t.TimeOfDay,
			t.DoseAmount,
			t.CreatedAt,
		)
		if err != nil {
			tx.Rollback()
			return errors.NewDatabaseError("failed to insert schedule time", err)
		}
	}

	// 3. Insert Schedule Days (only if FrequencySpecificDays)
	if schedule.Type == models.FrequencySpecificDays {
		dayQuery := `INSERT INTO schedule_days (schedule_id, day_of_week) VALUES ($1, $2)`
		for _, day := range schedule.Days {
			_, err = tx.ExecContext(ctx, dayQuery, schedule.ID, day)
			if err != nil {
				tx.Rollback()
				return errors.NewDatabaseError("failed to insert schedule day", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("failed to commit transaction", err)
	}

	return nil
}

// GetActiveByMedicationID retrieves a specific schedule
func (r *scheduleRepository) GetActiveByMedicationID(ctx context.Context, medicationID string) (*models.Schedule, error) {
	query := `
		SELECT id, medication_id, type, interval_days, start_date, end_date
		FROM schedules
		WHERE medication_id = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, medicationID)

	schedule := &models.Schedule{}
	err := row.Scan(
		&schedule.ID,
		&schedule.MedicationID,
		&schedule.Type,
		&schedule.IntervalDays,
		&schedule.StartDate,
		&schedule.EndDate,
	)

	if err == nil {
		return schedule, nil
	}

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("schedule", medicationID)
	}

	return nil, errors.NewDatabaseError("failed to get schedule", err)
}

// GetSchedulesByMedicationID retrieves a schedules and its related data by medication ID
func (r *scheduleRepository) GetSchedulesByMedicationID(ctx context.Context, medicationID string) (*models.Schedule, error) {
	// 1. Get the base schedule
	query := `
		SELECT id, medication_id, type, interval_days, start_date, end_date, created_at, updated_at
		FROM schedules
		WHERE medication_id = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, medicationID)

	schedule := &models.Schedule{}
	err := row.Scan(
		&schedule.ID,
		&schedule.MedicationID,
		&schedule.Type,
		&schedule.IntervalDays,
		&schedule.StartDate,
		&schedule.EndDate,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("schedule", medicationID)
		}
		return nil, errors.NewDatabaseError("failed to get schedule", err)
	}

	// 2. Get schedule times
	timeQuery := `
		SELECT id, time_of_day, dose_amount, created_at
		FROM schedule_times
		WHERE schedule_id = $1
		ORDER BY time_of_day
	`
	rows, err := r.db.QueryContext(ctx, timeQuery, schedule.ID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get schedule times", err)
	}
	defer rows.Close()

	schedule.Times = make([]models.ScheduleTime, 0)
	for rows.Next() {
		t := models.ScheduleTime{}
		if err := rows.Scan(&t.ID, &t.TimeOfDay, &t.DoseAmount, &t.CreatedAt); err != nil {
			return nil, errors.NewDatabaseError("failed to scan schedule time", err)
		}
		t.ScheduleID = schedule.ID
		schedule.Times = append(schedule.Times, t)
	}

	// 3. Get schedule days
	schedule.Days = make([]int, 0)
	if schedule.Type == models.FrequencySpecificDays {
		dayQuery := `SELECT day_of_week FROM schedule_days WHERE schedule_id = $1 ORDER BY day_of_week`
		dayRows, err := r.db.QueryContext(ctx, dayQuery, schedule.ID)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to get schedule days", err)
		}
		defer dayRows.Close()

		for dayRows.Next() {
			var day int
			if err := dayRows.Scan(&day); err != nil {
				return nil, errors.NewDatabaseError("failed to scan schedule day", err)
			}
			schedule.Days = append(schedule.Days, day)
		}
	}

	return schedule, nil
}

// GetAllActiveSchedules retrieves all active schedules with their related data
func (r *scheduleRepository) GetAllActiveSchedules(ctx context.Context) ([]*models.Schedule, error) {
	// 1. Get all active base schedules
	query := `
		SELECT s.id, s.medication_id, s.type, s.interval_days, s.start_date, s.end_date, u.timezone AS user_time_zone, s.created_at, s.updated_at
		FROM schedules AS s
		JOIN medications AS m ON s.medication_id = m.id
		JOIN users AS u ON m.user_id = u.id
		WHERE s.deleted_at IS NULL AND (s.end_date IS NULL OR s.end_date >= CURRENT_DATE)
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get active schedules", err)
	}
	defer rows.Close()

	var schedules []*models.Schedule
	scheduleMap := make(map[string]*models.Schedule)

	for rows.Next() {
		s := &models.Schedule{}
		if err := rows.Scan(
			&s.ID, &s.MedicationID, &s.Type, &s.IntervalDays, &s.StartDate, &s.EndDate, &s.UserTimeZone, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, errors.NewDatabaseError("failed to scan schedule", err)
		}
		s.Times = make([]models.ScheduleTime, 0)
		s.Days = make([]int, 0)
		schedules = append(schedules, s)
		scheduleMap[s.ID] = s
	}
	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating schedules", err)
	}

	if len(schedules) == 0 {
		return schedules, nil
	}

	// 2. Get all times for these schedules
	timeQuery := `
		SELECT id, schedule_id, time_of_day, dose_amount, created_at
		FROM schedule_times
		WHERE schedule_id IN (SELECT id FROM schedules WHERE deleted_at IS NULL AND (end_date IS NULL OR end_date >= CURRENT_DATE))
		ORDER BY time_of_day
	`
	timeRows, err := r.db.QueryContext(ctx, timeQuery)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get schedule times", err)
	}
	defer timeRows.Close()

	for timeRows.Next() {
		var t models.ScheduleTime
		if err := timeRows.Scan(&t.ID, &t.ScheduleID, &t.TimeOfDay, &t.DoseAmount, &t.CreatedAt); err != nil {
			return nil, errors.NewDatabaseError("failed to scan schedule time", err)
		}
		if s, ok := scheduleMap[t.ScheduleID]; ok {
			s.Times = append(s.Times, t)
		}
	}

	// 3. Get all days for these schedules
	dayQuery := `
		SELECT schedule_id, day_of_week
		FROM schedule_days
		WHERE schedule_id IN (SELECT id FROM schedules WHERE deleted_at IS NULL AND (end_date IS NULL OR end_date >= CURRENT_DATE))
	`
	dayRows, err := r.db.QueryContext(ctx, dayQuery)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get schedule days", err)
	}
	defer dayRows.Close()

	for dayRows.Next() {
		var scheduleID string
		var day int
		if err := dayRows.Scan(&scheduleID, &day); err != nil {
			return nil, errors.NewDatabaseError("failed to scan schedule day", err)
		}
		if s, ok := scheduleMap[scheduleID]; ok {
			s.Days = append(s.Days, day)
		}
	}

	return schedules, nil
}

// UpdateSchedule updates a schedule
func (r *scheduleRepository) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("failed to start transaction", err)
	}

	// FLOW
	// Step 1: Delete existing times and days
	// Step 2: Insert new times and days
	// Step 3: Update the main schedule entry with new dates or other fields

	// Step 1: Delete existing times
	deleteTimeQuery := `DELETE FROM schedule_times WHERE schedule_id = $1`
	_, err = tx.ExecContext(ctx, deleteTimeQuery, schedule.ID)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to delete schedule times", err)
	}

	// Delete existing days
	deleteDayQuery := `DELETE FROM schedule_days WHERE schedule_id = $1`
	_, err = tx.ExecContext(ctx, deleteDayQuery, schedule.ID)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to delete schedule days", err)
	}

	// Step 2: Insert new times and days

	// Insert new times
	insertTimeQuery := `INSERT INTO schedule_times (id, schedule_id, time_of_day, dose_amount, created_at) VALUES ($1, $2, $3, $4, $5)`
	for _, t := range schedule.Times {
		_, err = tx.ExecContext(ctx, insertTimeQuery, t.ID, schedule.ID, t.TimeOfDay, t.DoseAmount, t.CreatedAt)
		if err != nil {
			tx.Rollback()
			return errors.NewDatabaseError("failed to insert schedule time", err)
		}
	}

	// Insert new days
	insertDayQuery := `INSERT INTO schedule_days (schedule_id, day_of_week) VALUES ($1, $2)`
	for _, day := range schedule.Days {
		_, err = tx.ExecContext(ctx, insertDayQuery, schedule.ID, day)
		if err != nil {
			tx.Rollback()
			return errors.NewDatabaseError("failed to insert schedule day", err)
		}
	}

	// Step 3: Update the main schedule entry with new dates or other fields
	query := `
		UPDATE schedules 
		SET 
			type = $1,
			interval_days = $2,
			start_date = $3,
			end_date = $4,
			updated_at = $5
		WHERE id = $6
	`

	_, err = tx.ExecContext(ctx, query,
		string(schedule.Type),
		schedule.IntervalDays,
		schedule.StartDate,
		schedule.EndDate,
		schedule.UpdatedAt,
		schedule.ID,
	)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to update schedule", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("failed to commit transaction", err)
	}

	return nil
}

// DeleteSchedule soft delete by medication id & schedule id
func (r *scheduleRepository) DeleteSchedule(ctx context.Context, medicationID, scheduleID string) error {
	query := `UPDATE schedules SET deleted_at = NOW() WHERE medication_id = $1 AND id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, medicationID, scheduleID)
	if err != nil {
		return errors.NewDatabaseError("failed to delete schedule", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to get affected rows", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("schedule", fmt.Sprintf("medication_id = %s and schedule_id = %s", medicationID, scheduleID))
	}

	return nil
}
