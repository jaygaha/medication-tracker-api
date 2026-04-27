// internal/models/log.go
// This is medication logs

package models

import (
	"errors"
	"time"
)

// Custom type for log actions
type ActionType string

const (
	ActionTypeTaken   ActionType = "taken"
	ActionTypeSkipped ActionType = "skipped"
)

type MedicationLog struct {
	ID                 string     `json:"id" db:"id"`
	MedicationID       string     `json:"medication_id" db:"medication_id"`
	UserID             string     `json:"-" db:"user_id"`
	ScheduleTimeID     *string    `json:"schedule_time_id,omitempty" db:"schedule_time_id"`
	Status             ActionType `json:"status" db:"status"` // db is 'status', json can still be whatever you want
	DoseTaken          *float64   `json:"dose_taken,omitempty" db:"dose_taken"`
	ScheduledTimestamp *time.Time `json:"scheduled_timestamp,omitempty" db:"scheduled_timestamp"`
	ActualTimestamp    time.Time  `json:"actual_timestamp" db:"actual_timestamp"`
	Notes              *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// CreateLogRequest defines the payload for creating a new medication log
type CreateLogRequest struct {
	ScheduleTimeID     *string    `json:"schedule_time_id,omitempty" binding:"omitempty,uuid"` // Must be a valid UUID if provided
	Status             ActionType `json:"status" binding:"required,oneof=taken skipped"`
	DoseTaken          *float64   `json:"dose_taken,omitempty" binding:"omitempty,gt=0"` // Must be > 0 if provided
	ScheduledTimestamp *time.Time `json:"scheduled_timestamp,omitempty" binding:"omitempty"`
	ActualTimestamp    *time.Time `json:"actual_timestamp,omitempty" binding:"omitempty"`
	Notes              *string    `json:"notes,omitempty" binding:"omitempty"`
}

// Validate validates the CreateLogRequest payload
func (r *CreateLogRequest) Validate() error {
	// If an actual timestamp is provided, ensure it's not in the future
	if r.ActualTimestamp != nil && r.ActualTimestamp.After(time.Now()) {
		return errors.New("actual_timestamp cannot be in the future")
	}

	return nil
}
