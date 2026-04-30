// internal/service/log_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type LogService struct {
	repo           repository.LogRepository
	medicationRepo repository.MedicationRepository
}

func NewLogService(repo repository.LogRepository, medicationRepo repository.MedicationRepository) *LogService {
	return &LogService{
		repo:           repo,
		medicationRepo: medicationRepo,
	}
}

func (s *LogService) CreateLog(ctx context.Context, medicationID string, req *models.CreateLogRequest) (*models.MedicationLog, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	medication, err := s.medicationRepo.GetMedicationByID(ctx, medicationID)
	if err != nil {
		return nil, err
	}

	// Handle Timestamps (Default to now if missing)
	actualTime := time.Now()
	if req.ActualTimestamp != nil {
		actualTime = *req.ActualTimestamp
	}

	// Check if log already exists
	var scheduleTimeID string
	if req.ScheduleTimeID != nil {
		scheduleTimeID = *req.ScheduleTimeID
		exists, err := s.repo.CheckExists(ctx, medicationID, scheduleTimeID, actualTime)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.NewValidationError("this dose has already been logged today")
		}
	}

	// Build the MedicationLog model
	now := time.Now()
	log := &models.MedicationLog{
		ID:                 uuid.NewString(),
		MedicationID:       medicationID,
		UserID:             medication.UserID,
		ScheduleTimeID:     req.ScheduleTimeID,
		Status:             req.Status,
		DoseTaken:          req.DoseTaken,
		ScheduledTimestamp: req.ScheduledTimestamp,
		ActualTimestamp:    actualTime,
		Notes:              req.Notes,
		CreatedAt:          now,
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *LogService) CheckExists(ctx context.Context, medicationID, scheduleTimeID string, date time.Time) (bool, error) {
	return s.repo.CheckExists(ctx, medicationID, scheduleTimeID, date)
}

func (s *LogService) GetAdherenceStats(ctx context.Context, medicationID, userID string, from, to time.Time) (*models.AdherenceStats, error) {
	// Validate the medication exists and belongs to the user
	if _, err := s.medicationRepo.GetMedicationByID(ctx, medicationID); err != nil {
		return nil, err
	}

	// Default to the last 30 days if not provided
	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now().Truncate(24 * time.Hour)
	}

	return s.repo.GetAdherenceStats(ctx, medicationID, userID, from, to)
}
