package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"
)

func AuthRoutes(deps models.Config, api *gin.RouterGroup) {

	// initialize repositories
	repo := repository.InitAuthRepository(deps.Database, deps.Logger)

	// initialize services
	service := services.InitAuthService(repo, deps.Logger)

	// initialize handlers
	handler := handlers.InitAuthHandler(service, deps.Logger)

	auth := api.Group("/auth")
	{
		auth.POST("/signin", handler.SignIn)
		auth.POST("/signup", handler.SignUp)
	}
}
