// internal/service/auth_service.go
package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jaygaha/medication-tracker-api/internal/config"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.NewValidationError("email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		PasswordHash:           string(hashedPassword),
		Timezone:               "UTC", // default
		NotificationPreference: models.NotificationNone, // default
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// return generic error to prevent email enumeration
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) generateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.config.Server.JwtExpirationHours) * time.Hour)

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Server.JwtSecret))
}
