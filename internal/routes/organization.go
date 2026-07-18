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

	org := api.Group("/organization")
	{
		org.DELETE("/delete", middleware.ValidateJWT(), handler.DeleteOrganization)
		org.POST("/create", handler.CreateOrganization)
		org.PATCH("/update", middleware.ValidateJWT(), middleware.Authorize("org_admin"), handler.UpdateOrganization)
		org.GET("/get", middleware.ValidateJWT(), handler.GetOrganizationByID)
	}
}
