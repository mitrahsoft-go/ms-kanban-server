package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/postgres"
	"github.com/ms-kanban-server/internal/pkg/logger"
	"github.com/ms-kanban-server/internal/routes"
	"go.uber.org/zap"
)

func main() {

	// Load configuration, initialize database connection, set up routes, and start the server
	config := configs.LoadEnv()

	// Initialize the logger
	Logger, err := logger.InitLogger(config)
	if err != nil {
		log.Println("Failed to initialize logger:", err)
	}
	// Initialize the database connection
	dbConn, err := postgres.InitDB(config)
	if err != nil {
		Logger.Fatal(fmt.Sprintf("Failed to initialize database connection:%v", err))
	}

	//Initialize the Gin router and set up routes
	router := gin.Default()

	// Set up routes
	routes.SetupRoutes(router, dbConn)

	// Start the server
	Logger.Info("Server is running",
		zap.String("port", config.HTTP.Port))
	err = router.Run(fmt.Sprintf(":%s", config.HTTP.Port))
	if err != nil {
		Logger.Info("Failed to start server ",
			zap.String("ERROR", err.Error()))

	}
}
