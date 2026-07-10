package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/postgres"
	"github.com/ms-kanban-server/internal/routes"
)

func main() {

	// Load configuration, initialize database connection, set up routes, and start the server
	config := configs.LoadEnv()

	// Initialize the database connection
	dbConn, err := postgres.InitDB(config)
	if err != nil {
		log.Fatal("Failed to initialize database connection:", err)
	}

	//Initialize the Gin router and set up routes
	router := gin.Default()

	// Set up routes
	routes.SetupRoutes(router, dbConn)

	// Start the server
	log.Println("Server is running on port", config.HTTP.Port)
	router.Run(fmt.Sprintf(":%s", config.HTTP.Port))
}
