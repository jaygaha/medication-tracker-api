// internal/handler/log_handler.go
package handler

import (
	"net/http"

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
