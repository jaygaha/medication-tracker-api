// internal/service/medication_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

// MedicationService handles business logic for medications
type MedicationService struct {
	repo repository.MedicationRepository
}

// NewMedicationService creates a new medication service
func NewMedicationService(repo repository.MedicationRepository) *MedicationService {
	return &MedicationService{repo: repo}
}

// CreateMedication creates a new medication
func (s *MedicationService) CreateMedication(ctx context.Context, med *models.Medication) error {
	if err := med.Validate(); err != nil {
		return err
	}

	// Set default values
	med.ID = uuid.New().String()
	med.CreatedAt = time.Now()
	med.UpdatedAt = time.Now()

	// Link visuals if provided
	if med.Visuals != nil {
		med.Visuals.MedicationID = med.ID
	}

	// Save to repo
	if err := s.repo.CreateMedication(ctx, med); err != nil {
		return err
	}

	return nil
}

// GetMedicationByID gets a medication by ID
func (s *MedicationService) GetMedicationByID(ctx context.Context, id string) (*models.Medication, error) {
	return s.repo.GetMedicationByID(ctx, id)
}

// UpdateMedication updates a medication
func (s *MedicationService) UpdateMedication(ctx context.Context, med *models.Medication) error {
	if err := med.Validate(); err != nil {
		return err
	}

	med.UpdatedAt = time.Now()

	return s.repo.UpdateMedication(ctx, med)
}

// DeleteMedication deletes a medication
func (s *MedicationService) DeleteMedication(ctx context.Context, id string) error {
	return s.repo.DeleteMedication(ctx, id)
}

// ListMedications lists medications for a user
func (s *MedicationService) ListMedications(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]models.Medication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListMedications(ctx, userID, limit, offset, orderBy, orderDir)
}
