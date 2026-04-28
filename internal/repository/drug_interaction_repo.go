// internal/repository/interaction_repo.go

package repository

import (
	"context"
	"database/sql"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type DrugInteractionRepository interface {
	CreateDrugInteraction(ctx context.Context, interaction *models.DrugInteraction) error
	GetDrugInteractions(ctx context.Context, userID string) ([]models.DrugInteraction, error)
	AcknowledgeDrugInteraction(ctx context.Context, interactionID, userID string) error
	DeleteDrugInteraction(ctx context.Context, interactionID, userID string) error
}

type drugInteractionRepository struct {
	db *sql.DB
}

func NewDrugInteractionRepository(db *sql.DB) DrugInteractionRepository {
	return &drugInteractionRepository{db: db}
}

func (r *drugInteractionRepository) CreateDrugInteraction(ctx context.Context, interaction *models.DrugInteraction) error {
	query := `
		INSERT INTO drug_interactions (
			id, user_id, medication_1_id, medication_2_id, severity, description, acknowledged, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (medication_1_id, medication_2_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query,
		interaction.ID,
		interaction.UserID,
		interaction.Medication1ID,
		interaction.Medication2ID,
		string(interaction.Severity),
		interaction.Description,
		interaction.Acknowledged,
		interaction.CreatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("failed to create drug interaction", err)
	}
	return nil
}

func (r *drugInteractionRepository) GetDrugInteractions(ctx context.Context, userID string) ([]models.DrugInteraction, error) {
	query := `
		SELECT 
			id, medication_1_id, medication_2_id, severity, description, acknowledged, created_at
		FROM drug_interactions 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get drug interactions", err)
	}
	defer rows.Close()

	var interactions []models.DrugInteraction
	for rows.Next() {
		var interaction models.DrugInteraction
		err := rows.Scan(
			&interaction.ID,
			&interaction.Medication1ID,
			&interaction.Medication2ID,
			&interaction.Severity,
			&interaction.Description,
			&interaction.Acknowledged,
			&interaction.CreatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan drug interaction", err)
		}
		interactions = append(interactions, interaction)
	}

	return interactions, nil
}

func (r *drugInteractionRepository) AcknowledgeDrugInteraction(ctx context.Context, interactionID string, userID string) error {
	query := `
		UPDATE drug_interactions 
		SET acknowledged = TRUE
		WHERE id = $1 AND user_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, interactionID, userID)
	if err != nil {
		return errors.NewDatabaseError("failed to acknowledge drug interaction", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("drug interaction not found", interactionID)
	}

	return nil
}

func (r *drugInteractionRepository) DeleteDrugInteraction(ctx context.Context, interactionID string, userID string) error {
	query := `
		DELETE FROM drug_interactions 
		WHERE id = $1 AND user_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, interactionID, userID)
	if err != nil {
		return errors.NewDatabaseError("failed to delete drug interaction", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("drug interaction not found", interactionID)
	}

	return nil
}
