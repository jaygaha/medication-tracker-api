// internal/models/interaction.go

package models

import (
	"errors"
	"time"
)

type InteractionSeverity string

const (
	SeverityMinor    InteractionSeverity = "minor"
	SeverityModerate InteractionSeverity = "moderate"
	SeveritySevere   InteractionSeverity = "severe"
	SeverityCritical InteractionSeverity = "critical"
)

type DrugInteraction struct {
	ID            string              `json:"id" db:"id"`
	UserID        string              `json:"-" db:"user_id"`
	Medication1ID string              `json:"medication_1_id" db:"medication_1_id"`
	Medication2ID string              `json:"medication_2_id" db:"medication_2_id"`
	Severity      InteractionSeverity `json:"severity" db:"severity"`
	Description   *string             `json:"description" db:"description"`
	Acknowledged  bool                `json:"acknowledged" db:"acknowledged"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
}

type CreateDrugInteractionRequest struct {
	Medication1ID string              `json:"medication_1_id" binding:"required"`
	Medication2ID string              `json:"medication_2_id" binding:"required"`
	Severity      InteractionSeverity `json:"severity" binding:"required"`
	Description   *string             `json:"description" binding:"omitempty"`
}

func (r *CreateDrugInteractionRequest) Validate() error {
	switch r.Severity {
	case SeverityMinor, SeverityModerate, SeveritySevere, SeverityCritical:
		// Valid severity level
	default:
		return errors.New("invalid severity level")
	}

	if r.Medication1ID == r.Medication2ID {
		return errors.New("medication_1_id and medication_2_id cannot be the same")
	}
	return nil
}
