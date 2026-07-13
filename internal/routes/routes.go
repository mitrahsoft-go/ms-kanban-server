package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/redis"
	handlers "github.com/ms-kanban-server/internal/handlers/http"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/repository"
	"github.com/ms-kanban-server/internal/services"
)

const appVersion = "1.0.0"

func SetupRoutes(conn models.Config, cfg *configs.Config) {
	// initialize repositories
	repo := repository.InitRepository(conn.Database, conn.Logger)

	// initialize services
	service := services.InitService(repo, conn.Logger)

	// initialize handlers
	handlers.InitHandler(service, conn.Logger)

	// Health endpoint for liveness checks and readiness validation
	conn.Router.GET("/health", func(c *gin.Context) {
		timestamp := time.Now().UTC().Format(time.RFC3339)
		full := c.Query("full") == "true"

		if !full {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"version":   appVersion,
				"timestamp": timestamp,
			})
			return
		}

		dependencies := map[string]string{
			"database": "healthy",
			"redis":    "healthy",
		}
		statusCode := 200

		// Check database connection
		sqlDB, err := conn.Database.DB()
		if err != nil {
			dependencies["database"] = "unhealthy"
			statusCode = 503
		} else if err := sqlDB.Ping(); err != nil {
			dependencies["database"] = "unhealthy"
			statusCode = 503
		}

		// Check Redis connection
		if err := redis.PingRedis(conn.Redis); err != nil {
			dependencies["redis"] = "unhealthy"
			statusCode = 503
		}

		status := "healthy"
		if statusCode != 200 {
			status = "unhealthy"
		}

		c.JSON(statusCode, gin.H{
			"status":       status,
			"version":      appVersion,
			"timestamp":    timestamp,
			"dependencies": dependencies,
		})
	})
}
