// internal/handler/schedule_handler.go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)

// ScheduleHandler handles HTTP requests for schedules
type ScheduleHandler struct {
	service *service.ScheduleService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(service *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// CreateSchedule godoc
//
// @Summary Create medication schedule
// @Description create a new medication schedule with multiple dose times and optional specific weekdays
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param schedule body models.Schedule true "Schedule Object"
// @Success 201 {object} models.Schedule
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/schedules [post]
func (h *ScheduleHandler) CreateSchedule(ctx *gin.Context) {
	medicationID := ctx.Param("id")

	var req models.Schedule
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force the medication ID from the URL to prevent mismatches
	req.MedicationID = medicationID

	// Initialize IDs and Timestamps
	now := time.Now()
	req.ID = uuid.NewString()
	req.CreatedAt = now
	req.UpdatedAt = now

	for i := range req.Times {
		req.Times[i].ID = uuid.NewString()
		req.Times[i].CreatedAt = now
	}

	if err := h.service.CreateSchedule(ctx.Request.Context(), &req); err != nil {
		// Distinguish validation/not-found errors (4xx) from database/internal errors (5xx)
		if _, ok := err.(*errors.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if _, ok := err.(*errors.NotFoundError); ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, req)
}

// GetSchedulesByMedicationID godoc
//
// @Summary Get medication schedules
// @Description get medication schedules by medication ID
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Success 200 {object} models.Schedule
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/schedules [get]
func (h *ScheduleHandler) GetSchedulesByMedicationID(ctx *gin.Context) {
	medicationID := ctx.Param("id")

	schedules, err := h.service.GetSchedulesByMedicationID(ctx.Request.Context(), medicationID)
	if err != nil {
		// Distinguish validation/not-found errors (4xx) from database/internal errors (5xx)
		if _, ok := err.(*errors.NotFoundError); ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, schedules)
}

// UpdateSchedule godoc
//
// @Summary Update medication schedule
// @Description update medication schedule by medication ID and schedule ID
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param schedule_id path string true "Schedule ID"
// @Param schedule body models.Schedule true "Schedule Object"
// @Success 200 {object} models.Schedule
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/schedules/{schedule_id} [put]
func (h *ScheduleHandler) UpdateSchedule(ctx *gin.Context) {
	medicationID := ctx.Param("id")
	scheduleID := ctx.Param("schedule_id")

	var req models.Schedule
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = scheduleID
	req.MedicationID = medicationID

	if err := h.service.UpdateSchedule(ctx.Request.Context(), &req); err != nil {
		// Distinguish validation/not-found errors (4xx) from database/internal errors (5xx)
		if _, ok := err.(*errors.ValidationError); ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if _, ok := err.(*errors.NotFoundError); ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// DeleteSchedule godoc
//
// @Summary Delete medication schedule
// @Description delete medication schedule by medication ID and schedule ID
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param schedule_id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id}/schedules/{schedule_id} [delete]
func (h *ScheduleHandler) DeleteSchedule(ctx *gin.Context) {
	medicationID := ctx.Param("id")
	scheduleID := ctx.Param("schedule_id")

	if err := h.service.DeleteSchedule(ctx.Request.Context(), medicationID, scheduleID); err != nil {
		// Distinguish validation/not-found errors (4xx) from database/internal errors (5xx)
		if _, ok := err.(*errors.NotFoundError); ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "schedule deleted successfully"})
}
