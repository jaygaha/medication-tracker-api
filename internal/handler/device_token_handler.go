// internal/handler/device_token_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)

type DeviceTokenHandler struct {
	deviceTokenService *service.DeviceTokenService
}

func NewDeviceTokenHandler(deviceTokenService *service.DeviceTokenService) *DeviceTokenHandler {
	return &DeviceTokenHandler{deviceTokenService: deviceTokenService}
}

// RegisterToken godoc
//
// @Summary Register device token
// @Description register device token
// @Tags Device Tokens
// @Accept json
// @Produce json
// @Param deviceToken body models.RegisterTokenRequest true "Device Token"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /device-tokens [post]
func (h *DeviceTokenHandler) RegisterToken(c *gin.Context) {
	var req models.RegisterTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.deviceTokenService.RegisterToken(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// DeleteToken godoc
//
// @Summary Delete device token
// @Description deactivate a device token (unsubscribe from notifications)
// @Tags Device Tokens
// @Accept json
// @Produce json
// @Param id path string true "Device Token ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /device-tokens/{id} [delete]
func (h *DeviceTokenHandler) DeleteToken(c *gin.Context) {
	tokenID := c.Param("id")
	userID := c.GetString("user_id")

	err := h.deviceTokenService.DeleteToken(c.Request.Context(), tokenID, userID)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device token deactivated"})
}
