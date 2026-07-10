package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, logger *zap.Logger) {
	// initialize repositories
	repo := repository.InitRepository(db, logger)

	// initialize services
	service := services.InitService(repo, logger)

	// initialize handlers
	handlers.InitHandler(service, logger)

}
