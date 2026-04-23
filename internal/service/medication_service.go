// internal/service/medication_service.go
package service

import (
	"context"
	"fmt"
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

	// Save to repo
	if err := s.repo.CreateMedication(ctx, med); err != nil {
		return fmt.Errorf("failed to create medication: %w", err)
	}

	return nil
}

// GetMedicationByID gets a medication by ID
func (s *MedicationService) GetMedicationByID(ctx context.Context, id string) (*models.Medication, error) {
	med, err := s.repo.GetMedicationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get medication by ID: %w", err)
	}

	return med, nil
}

// UpdateMedication updates a medication
func (s *MedicationService) UpdateMedication(ctx context.Context, med *models.Medication) error {
	if err := med.Validate(); err != nil {
		return err
	}

	med.UpdatedAt = time.Now()

	if err := s.repo.UpdateMedication(ctx, med); err != nil {
		return fmt.Errorf("failed to update medication: %w", err)
	}

	return nil
}

// DeleteMedication deletes a medication
func (s *MedicationService) DeleteMedication(ctx context.Context, id string) error {

	if err := s.repo.DeleteMedication(ctx, id); err != nil {
		return fmt.Errorf("failed to delete medication: %w", err)
	}

	return nil
}

// ListMedications lists medications for a user
func (s *MedicationService) ListMedications(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]models.Medication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	medications, err := s.repo.ListMedications(ctx, userID, limit, offset, orderBy, orderDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list medications: %w", err)
	}

	return medications, nil
}
