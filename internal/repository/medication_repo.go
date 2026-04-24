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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("failed to start transaction", err)
	}

	// 1. Insert Medication
	query := `
		INSERT INTO medications (
			id, user_id, name, form, 
			strength_value, strength_unit, 
			rx_number, notes, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, query,
		med.ID, med.UserID, med.Name, string(med.Form),
		med.StrengthValue, med.StrengthUnit,
		med.RxNumber, med.Notes, string(med.Status),
		med.CreatedAt, med.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to create medication", err)
	}

	// 2. Insert Visuals if provided
	if med.Visuals != nil {
		visualQuery := `
			INSERT INTO medication_visuals (
				medication_id, shape, primary_color, secondary_color, background_color
			) VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.ExecContext(ctx, visualQuery,
			med.ID, med.Visuals.Shape, med.Visuals.PrimaryColor,
			med.Visuals.SecondaryColor, med.Visuals.BackgroundColor,
		)
		if err != nil {
			tx.Rollback()
			return errors.NewDatabaseError("failed to create medication visuals", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("failed to commit transaction", err)
	}

	return nil
}

func (r *medicationRepository) GetMedicationByID(ctx context.Context, id string) (*models.Medication, error) {
	query := `
		SELECT m.id, m.name, m.form, 
		       m.strength_value, m.strength_unit, 
		       m.rx_number, m.notes, m.status,
		       m.created_at, m.updated_at,
		       mv.shape, mv.primary_color, mv.secondary_color, mv.background_color
		FROM medications m
		LEFT JOIN medication_visuals mv ON m.id = mv.medication_id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`

	med := &models.Medication{}
	visuals := &models.MedicationVisual{}

	var shape, pColor, sColor, bColor sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&med.ID, &med.Name, &med.Form,
		&med.StrengthValue, &med.StrengthUnit,
		&med.RxNumber, &med.Notes, &med.Status,
		&med.CreatedAt, &med.UpdatedAt,
		&shape, &pColor, &sColor, &bColor,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("medication", id)
		}
		return nil, errors.NewDatabaseError("failed to get medication", err)
	}

	if shape.Valid {
		visuals.MedicationID = med.ID
		visuals.Shape = shape.String
		visuals.PrimaryColor = pColor.String
		visuals.SecondaryColor = sColor.String
		visuals.BackgroundColor = bColor.String
		med.Visuals = visuals
	}

	return med, nil
}

func (r *medicationRepository) ListMedications(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]models.Medication, error) {
	query := fmt.Sprintf(`
		SELECT m.id, m.name, m.form, 
		       m.strength_value, m.strength_unit, 
		       m.rx_number, m.notes, m.status,
		       m.created_at, m.updated_at,
		       mv.shape, mv.primary_color, mv.secondary_color, mv.background_color
		FROM medications m
		LEFT JOIN medication_visuals mv ON m.id = mv.medication_id
		WHERE m.user_id = $1 AND m.deleted_at IS NULL
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, "m."+orderBy, orderDir)

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to list medications", err)
	}
	defer rows.Close()

	medications := make([]models.Medication, 0)
	for rows.Next() {
		med := models.Medication{}
		visuals := &models.MedicationVisual{}
		var shape, pColor, sColor, bColor sql.NullString

		err := rows.Scan(
			&med.ID, &med.Name, &med.Form,
			&med.StrengthValue, &med.StrengthUnit,
			&med.RxNumber, &med.Notes, &med.Status,
			&med.CreatedAt, &med.UpdatedAt,
			&shape, &pColor, &sColor, &bColor,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan medication", err)
		}

		if shape.Valid {
			visuals.MedicationID = med.ID
			visuals.Shape = shape.String
			visuals.PrimaryColor = pColor.String
			visuals.SecondaryColor = sColor.String
			visuals.BackgroundColor = bColor.String
			med.Visuals = visuals
		}

		medications = append(medications, med)
	}

	return medications, nil
}

func (r *medicationRepository) UpdateMedication(ctx context.Context, med *models.Medication) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("failed to start transaction", err)
	}

	// 1. Update Medication
	query := `
		UPDATE medications
		SET name = $2, form = $3, strength_value = $4, strength_unit = $5,
			rx_number = $6, notes = $7, status = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := tx.ExecContext(ctx, query,
		med.ID, med.Name, string(med.Form),
		med.StrengthValue, med.StrengthUnit,
		med.RxNumber, med.Notes, string(med.Status),
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return errors.NewDatabaseError("failed to update medication", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return errors.NewNotFoundError("medication", med.ID)
	}

	// 2. Update Visuals (Upsert)
	if med.Visuals != nil {
		visualQuery := `
			INSERT INTO medication_visuals (
				medication_id, shape, primary_color, secondary_color, background_color
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (medication_id) DO UPDATE SET
				shape = EXCLUDED.shape,
				primary_color = EXCLUDED.primary_color,
				secondary_color = EXCLUDED.secondary_color,
				background_color = EXCLUDED.background_color,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err = tx.ExecContext(ctx, visualQuery,
			med.ID, med.Visuals.Shape, med.Visuals.PrimaryColor,
			med.Visuals.SecondaryColor, med.Visuals.BackgroundColor,
		)
		if err != nil {
			tx.Rollback()
			return errors.NewDatabaseError("failed to update medication visuals", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("failed to commit transaction", err)
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
