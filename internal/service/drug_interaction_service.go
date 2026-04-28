// internal/service/interaction_service.go

package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type DrugInteractionService struct {
	repo           repository.DrugInteractionRepository
	medicationRepo repository.MedicationRepository
}

func NewDrugInteractionService(repo repository.DrugInteractionRepository, medicationRepo repository.MedicationRepository) *DrugInteractionService {
	return &DrugInteractionService{
		repo:           repo,
		medicationRepo: medicationRepo,
	}
}

func (s *DrugInteractionService) CreateDrugInteraction(ctx context.Context, req *models.CreateDrugInteractionRequest, userID string) (*models.DrugInteraction, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	medication1, err := s.medicationRepo.GetMedicationByID(ctx, req.Medication1ID)
	if err != nil {
		return nil, err
	}

	medication2, err := s.medicationRepo.GetMedicationByID(ctx, req.Medication2ID)
	if err != nil {
		return nil, err
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	interaction := &models.DrugInteraction{
		ID:            uuid.NewString(),
		UserID:        userID,
		Medication1ID: medication1.ID,
		Medication2ID: medication2.ID,
		Severity:      req.Severity,
		Description:   description,
		Acknowledged:  false,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateDrugInteraction(ctx, interaction); err != nil {
		return nil, err
	}

	return interaction, nil
}

func (s *DrugInteractionService) GetDrugInteractions(ctx context.Context, userID string) ([]models.DrugInteraction, error) {
	return s.repo.GetDrugInteractions(ctx, userID)
}

func (s *DrugInteractionService) AcknowledgeDrugInteraction(ctx context.Context, interactionID, userID string) error {
	return s.repo.AcknowledgeDrugInteraction(ctx, interactionID, userID)
}

func (s *DrugInteractionService) DeleteDrugInteraction(ctx context.Context, interactionID, userID string) error {
	return s.repo.DeleteDrugInteraction(ctx, interactionID, userID)
}
