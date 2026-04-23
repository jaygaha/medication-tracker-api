// internal/handler/medication_handler.go
// This is controller for medication

package handler

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)

// MedicationHandler handles HTTP requests for medications
type MedicationHandler struct {
	medService *service.MedicationService
}

// NewMedicationHandler creates a new medication handler
func NewMedicationHandler(medService *service.MedicationService) *MedicationHandler {
	return &MedicationHandler{medService: medService}
}

// CreateMedication godoc
//
// @Summary Create medication
// @Description create a new medication
// @Tags Medications
// @Accept json
// @Produce json
// @Param medication body models.CreateMedicationRequest true "Medication"
// @Success 201 {object} models.Medication
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications [post]
func (h *MedicationHandler) CreateMedication(c *gin.Context) {
	var req models.CreateMedicationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user id from context
	userID := c.GetString("user_id")

	med := &models.Medication{
		ID:            uuid.NewString(),
		UserID:        userID,
		Name:          req.Name,
		Form:          req.Form,
		StrengthValue: req.StrengthValue,
		StrengthUnit:  req.StrengthUnit,
		RxNumber:      req.RxNumber,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	if err := h.medService.CreateMedication(c.Request.Context(), med); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, med)
}

// GetMedication godoc
//
// @Summary Get medication by id
// @Description get a medication by id
// @Tags Medications
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Success 200 {object} models.Medication
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id} [get]
func (h *MedicationHandler) GetMedication(c *gin.Context) {
	id := c.Param("id")
	med, err := h.medService.GetMedicationByID(c.Request.Context(), id)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, med)
}

// ListMedication godoc
//
// @Summary List medications for a user
// @Description list medications for a user
// @Tags Medications
// @Accept json
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param page query int false "Page (default 1)"
// @Param order-by query string false "Order by name, form, strength_value, strength_unit, rx_number, notes, status, created_at, updated_at (default created_at)"
// @Param order-dir query string false "Order direction asc or desc (default desc)"
// @Success 200 {object} []models.Medication
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications [get]
func (h *MedicationHandler) ListMedication(c *gin.Context) {
	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("page")); err == nil {
		offset = (o - 1) * limit
	}

	orderBy := c.Query("order-by")
	orderDir := c.Query("order-dir")

	if orderBy == "" {
		orderBy = "created_at"
	}
	if orderDir == "" {
		orderDir = "desc"
	}

	// Allowed order by
	allowedOrderBy := []string{"name", "form", "strength_value", "strength_unit", "rx_number", "notes", "status", "created_at", "updated_at"}
	if !slices.Contains(allowedOrderBy, orderBy) {
		// AbortWithStatusJSON: the standard for Authentication, Authorization, and Validation.
		// It prevents the request from going any further than it needs to, saving unnecessary processing and resources.
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order by"})
		return
	}

	allowedOrderDir := []string{"asc", "desc"}
	if !slices.Contains(allowedOrderDir, orderDir) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order direction"})
		return
	}

	// Get user id from context
	userID := c.GetString("user_id")

	medications, err := h.medService.ListMedications(c.Request.Context(), userID, limit, offset, orderBy, orderDir)
	if err != nil {
		// This is the standard inside your actual Controller/Handler functions once you've already passed all middleware.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  medications,
		"limit": limit,
		"page":  offset,
	})
}

// UpdateMedication godoc
//
// @Summary Update medication by id
// @Description update a medication by id
// @Tags Medications
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Param medication body models.CreateMedicationRequest true "Medication"
// @Success 200 {object} models.Medication
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id} [put]
func (h *MedicationHandler) UpdateMedication(c *gin.Context) {
	id := c.Param("id")
	var req models.CreateMedicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	med := &models.Medication{
		ID:            id,
		Name:          req.Name,
		Form:          req.Form,
		StrengthValue: req.StrengthValue,
		StrengthUnit:  req.StrengthUnit,
		RxNumber:      req.RxNumber,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	if err := h.medService.UpdateMedication(c.Request.Context(), med); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, med)
}

// DeleteMedication godoc
//
// @Summary Delete medication by id
// @Description delete a medication by id
// @Tags Medications
// @Accept json
// @Produce json
// @Param id path string true "Medication ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /medications/{id} [delete]
func (h *MedicationHandler) DeleteMedication(c *gin.Context) {
	id := c.Param("id")
	if err := h.medService.DeleteMedication(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Medication deleted successfully"})
}
