// internal/service/device_token_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type DeviceTokenService struct {
	repo repository.DeviceTokenRepository
}

func NewDeviceTokenService(repo repository.DeviceTokenRepository) *DeviceTokenService {
	return &DeviceTokenService{repo: repo}
}

func (s *DeviceTokenService) RegisterToken(ctx context.Context, req *models.RegisterTokenRequest, userID string) error {
	if err := req.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	now := time.Now()
	deviceToken := &models.DeviceToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     req.Token,
		Platform:  req.Platform,
		LastUsed:  &now,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.RegisterToken(ctx, deviceToken); err != nil {
		return err
	}

	return nil
}

func (s *DeviceTokenService) DeleteToken(ctx context.Context, tokenID, userID string) error {
	return s.repo.DeleteToken(ctx, tokenID, userID)
}
