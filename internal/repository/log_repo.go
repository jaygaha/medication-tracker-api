// internal/repository/log_repo.go
// This is log repository

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
)

type LogRepository interface {
	CreateLog(ctx context.Context, log *models.MedicationLog) error
	CheckExists(ctx context.Context, medicationID string, scheduleTimeID string, date time.Time) (bool, error)
	GetAdherenceStats(ctx context.Context, medicationID, userID string, from, to time.Time) (*models.AdherenceStats, error)
}

type logRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) CreateLog(ctx context.Context, log *models.MedicationLog) error {
	query := `
		INSERT INTO medication_logs (
			id, user_id, medication_id, schedule_time_id, status, dose_taken, scheduled_timestamp, actual_timestamp, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.MedicationID,
		log.ScheduleTimeID,
		string(log.Status),
		log.DoseTaken,
		log.ScheduledTimestamp,
		log.ActualTimestamp,
		log.Notes,
		log.CreatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("failed to create medication log", err)
	}
	return nil
}

func (r *logRepository) CheckExists(ctx context.Context, medicationID string, scheduleTimeID string, date time.Time) (bool, error) {
	// Cast the actual_timestamp to a DATE type to compare against the given date
	query := `
		SELECT EXISTS(
			SELECT 1 FROM medication_logs 
			WHERE medication_id = $1 
			  AND schedule_time_id = $2 
			  AND DATE(actual_timestamp) = DATE($3)
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, medicationID, scheduleTimeID, date).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseError("failed to check log existence", err)
	}
	return exists, nil
}

func (r *logRepository) GetAdherenceStats(ctx context.Context, medicationID, userID string, from, to time.Time) (*models.AdherenceStats, error) {
	stats := &models.AdherenceStats{
		MedicationID: medicationID,
		From:         from.Format("2006-01-02"),
		To:           to.Format("2006-01-02"),
	}

	// --- Query 1: Aggregate totals in a single round-trip ---
	totalsQuery := `
		SELECT
			COUNT(*)                                   AS total_scheduled,
			COUNT(*) FILTER (WHERE status = 'taken')   AS total_taken,
			COUNT(*) FILTER (WHERE status = 'skipped') AS total_skipped
		FROM medication_logs
		WHERE medication_id        = $1
		  AND user_id              = $2
		  AND actual_timestamp::DATE BETWEEN $3 AND $4
	`
	err := r.db.QueryRowContext(ctx, totalsQuery, medicationID, userID, from, to).
		Scan(&stats.TotalScheduled, &stats.TotalTaken, &stats.TotalSkipped)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get adherence totals", err)
	}

	// Calculate adherence rate (guard division by zero)
	if stats.TotalScheduled > 0 {
		stats.AdherenceRate = float64(stats.TotalTaken) / float64(stats.TotalScheduled) * 100
	}

	// --- Query 2: Distinct taken dates for streak calculation ---
	datesQuery := `
		SELECT DISTINCT actual_timestamp::DATE AS log_date
		FROM medication_logs
		WHERE medication_id        = $1
		  AND user_id              = $2
		  AND status               = 'taken'
		  AND actual_timestamp::DATE BETWEEN $3 AND $4
		ORDER BY log_date DESC
	`
	rows, err := r.db.QueryContext(ctx, datesQuery, medicationID, userID, from, to)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get taken dates", err)
	}
	defer rows.Close()

	var takenDates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, errors.NewDatabaseError("failed to scan taken date", err)
		}
		takenDates = append(takenDates, d)
	}

	// --- Streak calculation in Go ---
	// takenDates is already sorted DESC (most recent first)
	if len(takenDates) > 0 {
		today := time.Now().Truncate(24 * time.Hour)

		// Current streak: consecutive days from today backwards
		current := 0
		expected := today
		for _, d := range takenDates {
			d = d.Truncate(24 * time.Hour)
			if d.Equal(expected) || d.Equal(expected.AddDate(0, 0, -1)) {
				if d.Equal(expected.AddDate(0, 0, -1)) {
					expected = d
				}
				current++
				expected = expected.AddDate(0, 0, -1)
			} else {
				break
			}
		}
		stats.CurrentStreak = current

		// Longest streak: sliding window over all taken dates (DESC → reverse to ASC)
		for i, j := 0, len(takenDates)-1; i < j; i, j = i+1, j-1 {
			takenDates[i], takenDates[j] = takenDates[j], takenDates[i]
		}
		longest, window := 1, 1
		for i := 1; i < len(takenDates); i++ {
			prev := takenDates[i-1].Truncate(24 * time.Hour)
			curr := takenDates[i].Truncate(24 * time.Hour)
			if curr.Equal(prev.AddDate(0, 0, 1)) {
				window++
			} else {
				window = 1
			}
			if window > longest {
				longest = window
			}
		}
		stats.LongestStreak = longest
	}

	return stats, nil
}
