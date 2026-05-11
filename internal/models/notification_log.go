// internal/models/notification_log.go
package models

import (
	"time"
)

type NotificationLog struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	ScheduleID     string    `json:"schedule_id" db:"schedule_id"`
	ScheduleTimeID string    `json:"schedule_time_id" db:"schedule_time_id"`
	ScheduledDate  time.Time `json:"scheduled_date" db:"scheduled_date"`
	SentAt         time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
