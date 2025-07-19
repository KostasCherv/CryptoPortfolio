package routes

import (
	"simple_api/internal/api/handlers"
	"simple_api/internal/api/middleware"
	"simple_api/internal/config"
	"simple_api/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
