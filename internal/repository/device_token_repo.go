// internal/repository/device_token_repo.go
package repository

import (
	"context"
	"database/sql"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type DeviceTokenRepository interface {
	RegisterToken(ctx context.Context, deviceToken *models.DeviceToken) error
	DeleteToken(ctx context.Context, tokenID, userID string) error
	GetActiveTokensByUserID(ctx context.Context, userID string) ([]*models.DeviceToken, error)
}

type deviceTokenRepository struct {
	DB *sql.DB
}

func NewDeviceTokenRepository(db *sql.DB) DeviceTokenRepository {
	return &deviceTokenRepository{DB: db}
}

func (r *deviceTokenRepository) RegisterToken(ctx context.Context, deviceToken *models.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (id, user_id, token, platform, last_used, is_active, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, token) DO UPDATE SET 
		is_active = true, updated_at = now()
	`

	_, err := r.DB.ExecContext(ctx, query,
		deviceToken.ID,
		deviceToken.UserID,
		deviceToken.Token,
		deviceToken.Platform,
		deviceToken.LastUsed,
		deviceToken.IsActive,
		deviceToken.CreatedAt,
		deviceToken.UpdatedAt,
		deviceToken.DeletedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("failed to register device token", err)
	}

	return nil
}

func (r *deviceTokenRepository) DeleteToken(ctx context.Context, tokenID, userID string) error {
	query := `
		UPDATE device_tokens
		SET is_active = false, deleted_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`

	result, err := r.DB.ExecContext(ctx, query, tokenID, userID)
	if err != nil {
		return errors.NewDatabaseError("failed to delete device token", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewNotFoundError("device token", tokenID)
	}

	return nil
}

func (r *deviceTokenRepository) GetActiveTokensByUserID(ctx context.Context, userID string) ([]*models.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, last_used, is_active, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1 AND is_active = true
	`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get active device tokens", err)
	}
	defer rows.Close()

	var tokens []*models.DeviceToken
	for rows.Next() {
		var t models.DeviceToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Token, &t.Platform, &t.LastUsed, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, errors.NewDatabaseError("failed to scan device token", err)
		}
		tokens = append(tokens, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating device tokens", err)
	}

	return tokens, nil
}
