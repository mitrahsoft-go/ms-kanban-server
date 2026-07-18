package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/middleware"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"
)

func AuthRoutes(deps models.Config, api *gin.RouterGroup) {

	// initialize repositories
	repo := repository.InitAuthRepository(deps)

	// initialize services
	service := services.InitAuthService(repo, deps.Logger)

	// initialize handlers
	handler := handlers.InitAuthHandler(service, deps.Logger)

	middleware := middleware.InitMiddleware(deps.Logger)

	auth := api.Group("/auth")
	{
		auth.POST("/signin", handler.SignIn)
		auth.POST("/refresh", middleware.ValidateJWT(), handler.RefreshToken)
		auth.POST("/logout", middleware.ValidateJWT(), handler.Logout)
		auth.POST("/signup", handler.SignUp)
		auth.POST("/change-password", middleware.ValidateJWT(), handler.ChangePassword)
		auth.POST("/password-reset/request", handler.RequestPasswordReset)
		auth.POST("/password-reset/confirm", handler.ResetPassword)
		auth.PATCH("/update", middleware.ValidateJWT(), handler.Updateuser)
		auth.GET("/me", middleware.ValidateJWT(), handler.GetUser)
	}
}
