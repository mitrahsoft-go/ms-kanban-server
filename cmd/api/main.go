package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/postgres"
	"github.com/ms-kanban-server/internal/routes"
)

func main() {
	config := configs.LoadEnv()
	dbConn := postgres.InitDB()

	router := gin.Default()

	routes.SetupRoutes(router, dbConn, config)

	fmt.Println("Server is running on port", config.HTTP.Port)
	router.Run(fmt.Sprintf(":%d", config.HTTP.Port))
}
