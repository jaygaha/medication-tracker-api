// internal/models/schedule.go
package models

import (
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
)

// FrequencyType represents the 'frequency_type' enum in the database.
type FrequencyType string

const (
	FrequencyEveryDay         FrequencyType = "every_day"
	FrequencyRegularIntervals FrequencyType = "regular_intervals"
	FrequencySpecificDays     FrequencyType = "specific_days"
	FrequencyAsNeeded         FrequencyType = "as_needed"
)

// Schedule represents the 'schedules' table.
// It defines the high-level pattern of when a medication should be taken.
type Schedule struct {
	ID           string        `json:"id" db:"id"`
	MedicationID string        `json:"medication_id" db:"medication_id"`
	Type         FrequencyType `json:"type" db:"type"`
	IntervalDays *int          `json:"interval_days,omitempty" db:"interval_days"` // Used for 'regular_intervals'
	StartDate    time.Time     `json:"start_date" db:"start_date"`
	EndDate      *time.Time    `json:"end_date,omitempty" db:"end_date"` // NULL if ongoing indefinitely
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time    `json:"-" db:"deleted_at"` // Soft delete

	// Nested data (populated by service/repository via joins or separate queries)
	Times []ScheduleTime `json:"times,omitempty" db:"-"`
	Days  []int          `json:"days,omitempty" db:"-"` // 1=Monday, 7=Sunday

	// Relation
	Medication *Medication `json:"medication,omitempty" db:"-"`
}

// ScheduleTime represents the 'schedule_times' table.
// Defines what time of day and how much to take.
type ScheduleTime struct {
	ID         string    `json:"id" db:"id"`
	ScheduleID string    `json:"-" db:"schedule_id"`
	TimeOfDay  string    `json:"time_of_day" db:"time_of_day"` // Format: "HH:MM:SS"
	DoseAmount float64   `json:"dose_amount" db:"dose_amount"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ScheduleDay represents a row in the 'schedule_days' table.
type ScheduleDay struct {
	ScheduleID string `db:"schedule_id"`
	DayOfWeek  int    `db:"day_of_week"` // 1=Monday, 7=Sunday
}

// Validate validates the core schedule data.
func (s *Schedule) Validate() error {
	if s.MedicationID == "" {
		return errors.NewValidationError("medication_id is required")
	}
	if s.Type == "" {
		return errors.NewValidationError("frequency type is required")
	}
	if !isValidFrequency(s.Type) {
		return errors.NewValidationError("invalid frequency type")
	}
	if s.Type == FrequencyRegularIntervals && (s.IntervalDays == nil || *s.IntervalDays <= 0) {
		return errors.NewValidationError("interval_days must be greater than 0 for regular_intervals")
	}
	if s.StartDate.IsZero() {
		return errors.NewValidationError("start_date is required")
	}
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return errors.NewValidationError("end_date cannot be before start_date")
	}

	// Validate nested times if provided
	for _, t := range s.Times {
		if err := t.Validate(); err != nil {
			return err
		}
	}

	// Validate nested days if provided
	if s.Type == FrequencySpecificDays && len(s.Days) == 0 {
		return errors.NewValidationError("at least one day of the week is required for specific_days")
	}
	for _, day := range s.Days {
		if day < 1 || day > 7 {
			return errors.NewValidationError("day_of_week must be between 1 (Monday) and 7 (Sunday)")
		}
	}

	return nil
}

// Validate validates a specific dose time.
func (t *ScheduleTime) Validate() error {
	if t.TimeOfDay == "" {
		return errors.NewValidationError("time_of_day is required")
	}
	if !isValidTimeFormat(t.TimeOfDay) {
		return errors.NewValidationError("invalid time format, use HH:MM or HH:MM:SS")
	}
	if t.DoseAmount <= 0 {
		return errors.NewValidationError("dose_amount must be greater than 0")
	}
	return nil
}

// Helper: Check if a schedule is active on a given date.
func (s *Schedule) IsActiveOn(date time.Time) bool {
	// Truncate to day for comparison
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	start := time.Date(s.StartDate.Year(), s.StartDate.Month(), s.StartDate.Day(), 0, 0, 0, 0, time.UTC)

	if d.Before(start) {
		return false
	}
	if s.EndDate != nil {
		end := time.Date(s.EndDate.Year(), s.EndDate.Month(), s.EndDate.Day(), 0, 0, 0, 0, time.UTC)
		if d.After(end) {
			return false
		}
	}
	return true
}

// Helper functions
func isValidTimeFormat(timeStr string) bool {
	// Supports HH:MM and HH:MM:SS
	_, err := time.Parse("15:04", timeStr)
	if err == nil {
		return true
	}
	_, err = time.Parse("15:04:05", timeStr)
	return err == nil
}

func isValidFrequency(freq FrequencyType) bool {
	validFrequencies := map[FrequencyType]bool{
		FrequencyEveryDay:         true,
		FrequencyRegularIntervals: true,
		FrequencySpecificDays:     true,
		FrequencyAsNeeded:         true,
	}
	return validFrequencies[freq]
}
