package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"

	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, dbConn *gorm.DB) {
	// initialize repositories
	repo := repository.InitRepositories(dbConn)

	// initialize services
	service := services.InitService(repo)

	// initialize handlers
	handlers.InitHandler(service)

}
