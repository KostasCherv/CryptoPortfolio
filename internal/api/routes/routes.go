package routes

import (
	"simple_api/internal/api/handlers"
	"simple_api/internal/api/middleware"
	"simple_api/internal/config"
	"simple_api/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, log *logger.Logger, cfg *config.Config) *gin.Engine {
	handler := handlers.NewHandler(db, cfg)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/auth/register", handler.Register())
		v1.POST("/auth/login", handler.Login())

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.Auth(cfg))
		{
			protected.GET("/users/me", handler.GetCurrentUser())
			protected.PUT("/users/me", handler.UpdateUser())
		}
	}

	return router
}
