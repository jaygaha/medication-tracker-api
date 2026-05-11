// internal/repository/notification_log_repo.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type NotificationLogRepository interface {
	HasSentNotification(ctx context.Context, scheduleTimeID string, scheduledDate time.Time) (bool, error)
	RecordNotification(ctx context.Context, log *models.NotificationLog) error
}

type notificationLogRepository struct {
	db *sql.DB
}

func NewNotificationLogRepository(db *sql.DB) NotificationLogRepository {
	return &notificationLogRepository{db: db}
}

func (r *notificationLogRepository) HasSentNotification(ctx context.Context, scheduleTimeID string, scheduledDate time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM notification_logs
			WHERE schedule_time_id = $1 AND scheduled_date = $2
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, scheduleTimeID, scheduledDate).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseError("failed to check notification log", err)
	}
	return exists, nil
}

func (r *notificationLogRepository) RecordNotification(ctx context.Context, log *models.NotificationLog) error {
	query := `
		INSERT INTO notification_logs (user_id, schedule_id, schedule_time_id, scheduled_date, sent_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.UserID,
		log.ScheduleID,
		log.ScheduleTimeID,
		log.ScheduledDate,
		log.SentAt,
	)
	if err != nil {
		return errors.NewDatabaseError("failed to record notification log", err)
	}
	return nil
}
