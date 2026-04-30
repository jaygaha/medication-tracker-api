// internal/routes/api.go
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaygaha/medication-tracker-api/internal/handler"
	"github.com/jaygaha/medication-tracker-api/internal/middleware"
)

// HealthCheck godoc
//
// @Summary Health Check
// @Description Check if the API is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func SetupRouter(
	medicationHandler *handler.MedicationHandler,
	scheduleHandler *handler.ScheduleHandler,
	logHandler *handler.LogHandler,
	drugInteractionHandler *handler.DrugInteractionHandler,
) *gin.Engine {
	gin.SetMode(gin.DebugMode)

	router := gin.Default()

	// Group routes -> /api/v1
	api := router.Group("/api/v1", middleware.AttachUserMiddleware())

	api.GET("/health", HealthCheck)

	// Medication Routes
	meds := api.Group("/medications")
	{
		meds.POST("", medicationHandler.CreateMedication)
		meds.GET("/:id", medicationHandler.GetMedication)
		meds.GET("", medicationHandler.ListMedication)
		meds.PUT("/:id", medicationHandler.UpdateMedication)
		meds.DELETE("/:id", medicationHandler.DeleteMedication)

		// Schedule Routes
		meds.GET("/:id/schedules", scheduleHandler.GetSchedulesByMedicationID)
		meds.POST("/:id/schedules", scheduleHandler.CreateSchedule)
		meds.PUT("/:id/schedules/:schedule_id", scheduleHandler.UpdateSchedule)
		meds.DELETE("/:id/schedules/:schedule_id", scheduleHandler.DeleteSchedule)

		// Log Routes
		meds.POST("/:id/logs", logHandler.CreateLog)
		meds.GET("/:id/stats", logHandler.GetAdherenceStats)
	}

	// Drug Interaction Routes
	drugInteractions := api.Group("/drug-interactions")
	{
		drugInteractions.POST("", drugInteractionHandler.CreateDrugInteraction)
		drugInteractions.GET("", drugInteractionHandler.GetDrugInteractions)
		drugInteractions.PATCH("/:id/acknowledge", drugInteractionHandler.AcknowledgeDrugInteraction)
		drugInteractions.DELETE("/:id", drugInteractionHandler.DeleteDrugInteraction)
	}

	return router
}
