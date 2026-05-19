// internal/repository/user_repo.go
package repository

import (
	"context"
	"database/sql"

	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type UserRepository interface {
	GetUserProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{DB: db}
}

func (r *userRepository) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, email, timezone, notification_preference
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Timezone, &user.NotificationPreference,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserProfile updates the profile of a user.
// Update the fields that are only provided in the request.
func (r *userRepository) UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users
		SET
			first_name = COALESCE($2, first_name),
			last_name = COALESCE($3, last_name),
			timezone = COALESCE($4, timezone),
			notification_preference = COALESCE($5, notification_preference),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, first_name, last_name, email, timezone, notification_preference, updated_at
	`

	var user models.User
	err := r.DB.QueryRowContext(ctx, query,
		userID, req.FirstName, req.LastName, req.Timezone, req.NotificationPreference,
	).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Timezone, &user.NotificationPreference, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash, timezone, notification_preference)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRowContext(ctx, query,
		user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Timezone, user.NotificationPreference,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, timezone, notification_preference
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Timezone, &user.NotificationPreference,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil to indicate not found vs a real error
		}
		return nil, err
	}

	return &user, nil
}
