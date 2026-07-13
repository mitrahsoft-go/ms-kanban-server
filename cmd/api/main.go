package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config"
	"github.com/ms-kanban-server/drivers/postgres"
	"github.com/ms-kanban-server/drivers/redis"
	"github.com/ms-kanban-server/internal/pkg/logger"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/routes"
	"go.uber.org/zap"
)

func main() {

	// Load configuration, initialize database connection, set up routes, and start the server
	config := config.LoadEnv()

	// Initialize the logger
	Logger, err := logger.InitLogger(config)
	if err != nil {
		log.Fatal("Failed to initialize logger :", err)
	}

	// Initialize the database connection
	dbConn, err := postgres.InitDB(config)
	if err != nil {
		Logger.Fatal("Failed to initialize database connection :",
			zap.String("ERROR : ", err.Error()))
	}

	// Initialize the RedisClient
	redisClient, err := redis.InitRedisClient(config)
	if err != nil {
		Logger.Fatal("Failed to initialize Redis client:",
			zap.String("ERROR : ", err.Error()))
	}

	//Initialize the Gin router and set up routes
	router := gin.Default()

	deps := models.Config{
		Database: dbConn,
		Router:   router,
		Redis:    redisClient,
		Logger:   Logger,
	}
	// Set up routes
	routes.SetupRoutes(deps, config)

	// Start the server
	Logger.Info("Server is running",
		zap.String("port ", config.HTTP.Port))
	err = router.Run(fmt.Sprintf(":%s", config.HTTP.Port))
	if err != nil {
		Logger.Fatal("Failed to start server ",
			zap.String("ERROR : ", err.Error()))
	}
}
