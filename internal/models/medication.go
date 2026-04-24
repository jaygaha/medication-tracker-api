// internal/models/medication.go

package models

import (
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusRefilled  Status = "refilled"
	StatusCompleted Status = "completed"
	StatusSuspended Status = "suspended"
)

type Form string

const (
	FormTablet    Form = "tablet"
	FormCapsule   Form = "capsule"
	FormLiquid    Form = "liquid"
	FormInjection Form = "injection"
	FormOther     Form = "other"
)

type Medication struct {
	ID            string     `json:"id" db:"id"`
	UserID        string     `json:"-" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	Form          Form       `json:"form" db:"form"`
	StrengthValue float64    `json:"strength_value" db:"strength_value"`
	StrengthUnit  string     `json:"strength_unit" db:"strength_unit"`
	RxNumber      *string    `json:"rx_number" db:"rx_number"`
	Notes         *string    `json:"notes" db:"notes"`
	Status        Status     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Soft delete

	// Nested data
	Visuals *MedicationVisual `json:"visuals,omitempty" db:"-"`
}

// MedicationVisual represents the 'medication_visuals' table.
type MedicationVisual struct {
	MedicationID    string `json:"-" db:"medication_id"`
	Shape           string `json:"shape" db:"shape"`
	PrimaryColor    string `json:"primary_color" db:"primary_color"`
	SecondaryColor  string `json:"secondary_color" db:"secondary_color"`
	BackgroundColor string `json:"background_color" db:"background_color"`
}

// Validate validates medication struct
func (m *Medication) Validate() error {
	if m.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if m.Form == "" {
		return errors.NewValidationError("form is required")
	}

	return nil
}

// CreateMedicationRequest represents the request body for creating a medication
type CreateMedicationRequest struct {
	Name          string            `json:"name" binding:"required"`
	Form          Form              `json:"form" binding:"required"`
	StrengthValue float64           `json:"strength_value" binding:"required"`
	StrengthUnit  string            `json:"strength_unit" binding:"required"`
	RxNumber      *string           `json:"rx_number"`
	Notes         *string           `json:"notes"`
	Status        Status            `json:"status"`
	Visuals       *MedicationVisual `json:"visuals"`
}
