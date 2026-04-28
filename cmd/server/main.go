// cmd/server/main.go
package main

import (
	"fmt"
	"log"

	_ "github.com/jaygaha/medication-tracker-api/docs"
	"github.com/jaygaha/medication-tracker-api/internal/config"
	"github.com/jaygaha/medication-tracker-api/internal/handler"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
	"github.com/jaygaha/medication-tracker-api/internal/routes"
	"github.com/jaygaha/medication-tracker-api/internal/service"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Medication Tracker API
// @version 1.0
// @description This is a medication tracking server.

// @contact.name Jay Gaha
// @contact.email jaygaha@gmail.com

// @BasePath /api/v1
// @host localhost:5010
func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Database connection
	db, err := config.InitDB(cfg)

	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// initialize repositories
	medRepo := repository.NewMedicationRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	logRepo := repository.NewLogRepository(db)
	interactionRepo := repository.NewDrugInteractionRepository(db)

	// Initialize services
	medService := service.NewMedicationService(medRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, medRepo)
	logService := service.NewLogService(logRepo, medRepo)
	interactionService := service.NewDrugInteractionService(interactionRepo, medRepo)

	// Initialize handlers
	medHandler := handler.NewMedicationHandler(medService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	logHandler := handler.NewLogHandler(logService)
	interactionHandler := handler.NewDrugInteractionHandler(interactionService)

	// Initialize router
	router := routes.SetupRouter(medHandler, scheduleHandler, logHandler, interactionHandler)

	// Initialize swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// Add a new feature. The feature is that to transform to AI prompt for user initial data
//   enterred.Its like  Enter your initial thoughts and let our Prompt Generator surprise you. The
//   option be Simple, Advanced, & Expert. First add APIs.
