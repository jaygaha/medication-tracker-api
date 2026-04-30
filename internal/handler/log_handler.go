// internal/handler/log_handler.go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)

// LogHandler handles HTTP requests for medication logs
type LogHandler struct {
	logService *service.LogService
}

// NewLogHandler creates a new log handler
func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// CreateLog godoc
//
// @Summary Create medication log
// @Description Create a new medication log
// @Tags Medication Logs
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param log body models.CreateLogRequest true "Medication Log"
// @Success 201 {object} models.MedicationLog
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/logs [post]
func (h *LogHandler) CreateLog(c *gin.Context) {
	var req models.CreateLogRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service , which returns the fully populated log object
	medicationID := c.Param("id")
	log, err := h.logService.CreateLog(c.Request.Context(), medicationID, &req)

	if err != nil {
		if _, ok := err.(*errors.ValidationError); ok {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		if _, ok := err.(*errors.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// GetAdherenceStats godoc
//
// @Summary Get adherence stats for a medication
// @Description Returns total taken/skipped counts, adherence rate, and streaks for a medication. Defaults to the last 30 days.
// @Tags Medication Logs
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} models.AdherenceStats
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/stats [get]
func (h *LogHandler) GetAdherenceStats(c *gin.Context) {
	medicationID := c.Param("id")
	userID := c.GetString("user_id")

	var from, to time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date; use YYYY-MM-DD format"})
			return
		}
		from = parsed
	}

	if toStr := c.Query("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date; use YYYY-MM-DD format"})
			return
		}
		to = parsed
	}

	stats, err := h.logService.GetAdherenceStats(c.Request.Context(), medicationID, userID, from, to)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
