// internal/routes/api.go
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaygaha/medication-tracker-api/internal/handler"
	"github.com/jaygaha/medication-tracker-api/internal/middleware"
)

func SetupRouter(
	medicationHandler *handler.MedicationHandler,
) *gin.Engine {
	gin.SetMode(gin.DebugMode)

	router := gin.Default()

	// Group routes -> /api/v1
	api := router.Group("/api/v1", middleware.AttachUserMiddleware())

	// Health Check Route swagger godocs
	//
	// @Summary Health Check
	// @Description Check if the API is healthy
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 500 {object} map[string]interface{}
	// @Router /health [get]
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Medication Routes
	meds := api.Group("/medications")
	{
		meds.POST("", medicationHandler.CreateMedication)
		meds.GET(":id", medicationHandler.GetMedication)
		meds.GET("", medicationHandler.ListMedication)
		meds.PUT("/:id", medicationHandler.UpdateMedication)
		meds.DELETE("/:id", medicationHandler.DeleteMedication)
	}

	return router
}
