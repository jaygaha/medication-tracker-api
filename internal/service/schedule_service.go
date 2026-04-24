// internal/service/schedule_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

// ScheduleService handles schedule business logic
type ScheduleService struct {
	repo           repository.ScheduleRepository
	medicationRepo repository.MedicationRepository
}

// NewScheduleService creates a new schedule service
func NewScheduleService(repo repository.ScheduleRepository, medicationRepo repository.MedicationRepository) *ScheduleService {
	return &ScheduleService{
		repo:           repo,
		medicationRepo: medicationRepo,
	}
}

// CreateSchedule creates a new medication schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// validate
	if err := schedule.Validate(); err != nil {
		return err
	}

	// check if medication exists
	medication, err := s.medicationRepo.GetMedicationByID(ctx, schedule.MedicationID)
	if err != nil {
		return err
	}

	schedule.Medication = medication

	// Check for duplicate active schedules
	existingSchedule, err := s.repo.GetActiveByMedicationID(ctx, schedule.MedicationID)
	if err == nil && existingSchedule.ID != "" {
		return errors.NewValidationError("an active schedule already exists for this medication")
	}

	// save schedule
	if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
		return err
	}

	return nil
}

// GetSchedulesByMedicationID retrieves schedules and its related data by medication ID
func (s *ScheduleService) GetSchedulesByMedicationID(ctx context.Context, medicationID string) (*models.Schedule, error) {
	// check if medication exists
	medication, err := s.medicationRepo.GetMedicationByID(ctx, medicationID)
	if err != nil {
		return nil, err
	}

	// get schedules
	schedule, err := s.repo.GetSchedulesByMedicationID(ctx, medicationID)
	if err != nil {
		return nil, err
	}

	schedule.Medication = medication

	return schedule, nil
}

// UpdateSchedule updates a schedule
func (s *ScheduleService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// validate
	if err := schedule.Validate(); err != nil {
		return err
	}

	// check if medication exists
	medication, err := s.medicationRepo.GetMedicationByID(ctx, schedule.MedicationID)
	if err != nil {
		return err
	}

	schedule.Medication = medication

	// Check for duplicate active schedules
	existingSchedule, err := s.repo.GetActiveByMedicationID(ctx, schedule.MedicationID)
	if err == nil && existingSchedule.ID != "" && existingSchedule.ID != schedule.ID {
		return errors.NewValidationError("an active schedule already exists for this medication")
	}

	now := time.Now()
	schedule.UpdatedAt = now

	for i := range schedule.Times {
		schedule.Times[i].ID = uuid.NewString()
		schedule.Times[i].CreatedAt = now
	}

	// save schedule
	if err := s.repo.UpdateSchedule(ctx, schedule); err != nil {
		return err
	}

	return nil
}

// DeleteSchedule deletes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, medicationID, scheduleID string) error {
	// check if medication exists
	_, err := s.medicationRepo.GetMedicationByID(ctx, medicationID)
	if err != nil {
		return err
	}

	// delete schedule
	if err := s.repo.DeleteSchedule(ctx, medicationID, scheduleID); err != nil {
		return err
	}

	return nil
}
