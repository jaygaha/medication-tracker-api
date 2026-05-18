// internal/service/user_service.go
package service

import (
	"context"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type UserService interface {
	GetUserProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetUserProfile(ctx, userID)
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) (*models.User, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	return s.userRepo.UpdateUserProfile(ctx, userID, req)
}
