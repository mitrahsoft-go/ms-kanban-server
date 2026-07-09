package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config/configs"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"

	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, config *configs.Config) {
	// initialize repositories
	repo := repository.InitRepositories(db)

	// initialize services
	service := services.InitService(repo)

	// initialize handlers
	handlers.InitHandler(service)

}
