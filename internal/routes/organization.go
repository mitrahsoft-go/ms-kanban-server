package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/middleware"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"
)

func OrganizationRoutes(deps models.Config, api *gin.RouterGroup) {

	// initialize repositories
	repo := repository.InitOrganizationRepository(deps)

	// initialize services
	service := services.InitOrganizationService(repo, deps.Logger)

	// initialize handlers
	handler := handlers.InitOrganizationHandler(service, deps.Logger)

	middleware := middleware.InitMiddleware(deps.Logger)

	auth := api.Group("/organization")
	{
		auth.DELETE("/delete", middleware.ValidateJWT(), handler.DeleteOrganization)
		auth.POST("/create", handler.CreateOrganization)
		auth.PATCH("/update", middleware.ValidateJWT(), handler.UpdateOrganization)
		auth.GET("/get", middleware.ValidateJWT(), handler.GetOrganizationByID)
	}
}
