// internal/repository/medication_repo.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

// MedicationService handle

type MedicationRepository interface {
	CreateMedication(ctx context.Context, med *models.Medication) error
	GetMedicationByID(ctx context.Context, id string) (*models.Medication, error)
	ListMedications(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]models.Medication, error)
	UpdateMedication(ctx context.Context, med *models.Medication) error
	DeleteMedication(ctx context.Context, id string) error
}

type medicationRepository struct {
	db *sql.DB
}

func NewMedicationRepository(db *sql.DB) MedicationRepository {
	return &medicationRepository{db: db}
}

func (r *medicationRepository) CreateMedication(ctx context.Context, med *models.Medication) error {
	query := `
		INSERT INTO medications (
			id, user_id, name, form, 
			strength_value, strength_unit, 
			rx_number, notes, status
		) VALUES (
			$1, $2, $3, $4, 
			$5, $6, 
			$7, $8, $9
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		med.ID,
		med.UserID,
		med.Name,
		string(med.Form),
		med.StrengthValue,
		med.StrengthUnit,
		med.RxNumber,
		med.Notes,
		string(med.Status),
	)
	if err != nil {
		return errors.NewDatabaseError("failed to create medication", err)
	}

	return nil
}

func (r *medicationRepository) GetMedicationByID(ctx context.Context, id string) (*models.Medication, error) {
	query := `
		SELECT id, name, form, 
		       strength_value, strength_unit, 
		       rx_number, notes, status,
		       created_at, updated_at
		FROM medications
		WHERE id = $1 AND deleted_at IS NULL
	`

	med := &models.Medication{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&med.ID,
		&med.Name,
		&med.Form,
		&med.StrengthValue,
		&med.StrengthUnit,
		&med.RxNumber,
		&med.Notes,
		&med.Status,
		&med.CreatedAt,
		&med.UpdatedAt,
	)
	if err == nil {
		return med, nil
	}

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("medication", id)
	}

	return nil, errors.NewDatabaseError("failed to get medication", err)
}

func (r *medicationRepository) ListMedications(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]models.Medication, error) {
	query := fmt.Sprintf(`
		SELECT id, name, form, 
		       strength_value, strength_unit, 
		       rx_number, notes, status,
		       created_at, updated_at
		FROM medications
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, orderBy, orderDir)

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to list medications", err)
	}
	defer rows.Close()

	medications := make([]models.Medication, 0)
	for rows.Next() {
		med := models.Medication{}
		err := rows.Scan(
			&med.ID,
			&med.Name,
			&med.Form,
			&med.StrengthValue,
			&med.StrengthUnit,
			&med.RxNumber,
			&med.Notes,
			&med.Status,
			&med.CreatedAt,
			&med.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan medication", err)
		}
		medications = append(medications, med)
	}

	return medications, nil
}

func (r *medicationRepository) UpdateMedication(ctx context.Context, med *models.Medication) error {
	query := `
		UPDATE medications
		SET name = $2,
			form = $3,
			strength_value = $4,
			strength_unit = $5,
			rx_number = $6,
			notes = $7,
			status = $8,
			updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		med.ID,
		med.Name,
		string(med.Form),
		med.StrengthValue,
		med.StrengthUnit,
		med.RxNumber,
		med.Notes,
		string(med.Status),
		time.Now(),
	)
	if err != nil {
		return errors.NewDatabaseError("failed to update medication", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewNotFoundError("medication", med.ID)
	}

	return nil
}

func (r *medicationRepository) DeleteMedication(ctx context.Context, id string) error {
	query := `
		UPDATE medications
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return errors.NewDatabaseError("failed to delete medication", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewNotFoundError("medication", id)
	}

	return nil
}
