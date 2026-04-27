// internal/repository/log_repo.go
// This is log repository

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type LogRepository interface {
	CreateLog(ctx context.Context, log *models.MedicationLog) error
	CheckExists(ctx context.Context, medicationID string, scheduleTimeID string, date time.Time) (bool, error)
}

type logRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) CreateLog(ctx context.Context, log *models.MedicationLog) error {
	query := `
		INSERT INTO medication_logs (
			id, user_id, medication_id, schedule_time_id, status, dose_taken, scheduled_timestamp, actual_timestamp, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.MedicationID,
		log.ScheduleTimeID,
		string(log.Status),
		log.DoseTaken,
		log.ScheduledTimestamp,
		log.ActualTimestamp,
		log.Notes,
		log.CreatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("failed to create medication log", err)
	}
	return nil
}
func (r *logRepository) CheckExists(ctx context.Context, medicationID string, scheduleTimeID string, date time.Time) (bool, error) {
	// Cast the actual_timestamp to a DATE type to compare against the given date
	query := `
		SELECT EXISTS(
			SELECT 1 FROM medication_logs 
			WHERE medication_id = $1 
			  AND schedule_time_id = $2 
			  AND DATE(actual_timestamp) = DATE($3)
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, medicationID, scheduleTimeID, date).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseError("failed to check log existence", err)
	}
	return exists, nil
}
