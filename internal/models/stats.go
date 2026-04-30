// internal/models/stats.go

package models

type AdherenceStats struct {
	MedicationID   string  `json:"medication_id"`
	From           string  `json:"from"`            // ISO date
	To             string  `json:"to"`              // ISO date
	TotalScheduled int     `json:"total_scheduled"` // taken + skipped
	TotalTaken     int     `json:"total_taken"`
	TotalSkipped   int     `json:"total_skipped"`
	AdherenceRate  float64 `json:"adherence_rate"` // taken / scheduled * 100
	CurrentStreak  int     `json:"current_streak"` // consecutive days taken (most recent)
	LongestStreak  int     `json:"longest_streak"` // all-time best consecutive days
}
