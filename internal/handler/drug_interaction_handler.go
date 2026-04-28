// internal/handler/drug_interaction_handler.go
// This is controller for drug interaction

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)

type DrugInteractionHandler struct {
	drugInteractionService *service.DrugInteractionService
}

func NewDrugInteractionHandler(drugInteractionService *service.DrugInteractionService) *DrugInteractionHandler {
	return &DrugInteractionHandler{drugInteractionService: drugInteractionService}
}

// CreateDrugInteraction godoc
//
// @Summary Create drug interaction
// @Description create a new drug interaction
// @Tags Drug Interactions
// @Accept json
// @Produce json
// @Param drugInteraction body models.CreateDrugInteractionRequest true "Drug Interaction"
// @Success 201 {object} models.DrugInteraction
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drug-interactions [post]
func (h *DrugInteractionHandler) CreateDrugInteraction(c *gin.Context) {
	var req models.CreateDrugInteractionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.drugInteractionService.CreateDrugInteraction(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetDrugInteractions godoc
//
// @Summary Get drug interactions for a user
// @Description get all drug interactions for a user
// @Tags Drug Interactions
// @Accept json
// @Produce json
// @Success 200 {object} []models.DrugInteraction
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drug-interactions [get]
func (h *DrugInteractionHandler) GetDrugInteractions(c *gin.Context) {
	drugInteractions, err := h.drugInteractionService.GetDrugInteractions(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drugInteractions)
}

// AcknowledgeDrugInteraction godoc
//
// @Summary Acknowledge drug interaction
// @Description acknowledge a drug interaction
// @Tags Drug Interactions
// @Accept json
// @Produce json
// @Param id path string true "Drug Interaction ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drug-interactions/{id}/acknowledge [patch]
func (h *DrugInteractionHandler) AcknowledgeDrugInteraction(c *gin.Context) {
	id := c.Param("id")
	if err := h.drugInteractionService.AcknowledgeDrugInteraction(c.Request.Context(), id, c.GetString("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteDrugInteraction godoc
//
// @Summary Delete drug interaction
// @Description delete a drug interaction
// @Tags Drug Interactions
// @Accept json
// @Produce json
// @Param id path string true "Drug Interaction ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drug-interactions/{id} [delete]
func (h *DrugInteractionHandler) DeleteDrugInteraction(c *gin.Context) {
	id := c.Param("id")
	if err := h.drugInteractionService.DeleteDrugInteraction(c.Request.Context(), id, c.GetString("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
