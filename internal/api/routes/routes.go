package routes

import (
	"context"
	"fmt"
	"simple_api/internal/api/handlers"
	"simple_api/internal/api/middleware"
	"simple_api/internal/cache"
	"simple_api/internal/config"
	"simple_api/internal/repository"
	"simple_api/internal/services"
	"simple_api/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, log *logger.Logger, cfg *config.Config) *gin.Engine {
	// Initialize Redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := cache.NewRedisClient(redisAddr, cfg.Redis.Password, cfg.Redis.DB, log)
	
	// Test Redis connection
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Warn("Redis connection failed, continuing without cache", "error", err)
	} else {
		log.Info("Redis connected successfully")
	}
	
	// Initialize cache service
	cacheService := cache.NewCacheService(redisClient, log)
	userCache := cache.NewUserCache(cacheService)
	
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	watchlistRepo := repository.NewWatchlistRepository(db)
	
	// Initialize services with repositories and cache
	userService := services.NewUserService(userRepo, userCache, cfg, log)
	
	// Initialize Web3 service
	web3Service, err := services.NewWeb3Service(cfg, log)
	if err != nil {
		log.Error("Failed to initialize Web3 service", "error", err)
		// Continue without Web3 service for now
	}
	
	// Initialize balance fetcher service
	balanceFetcher := services.NewBalanceFetcherService(watchlistRepo, web3Service, cacheService, log, cfg)
	
	// Start the background balance fetcher
	balanceFetcher.Start(context.Background())
	
	// Initialize watchlist service
	watchlistService := services.NewWatchlistService(watchlistRepo, web3Service, balanceFetcher, cacheService, log)
	
	// Initialize handlers with services
	handler := handlers.NewHandler(userService)
	watchlistHandler := handlers.NewWatchlistHandler(watchlistService, log)

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
			
			// Watchlist routes
			watchlist := protected.Group("/watchlist")
			{
				// Wallet management
				watchlist.POST("/wallets", watchlistHandler.AddWallet())
				watchlist.GET("/wallets", watchlistHandler.GetWallets())
				watchlist.DELETE("/wallets/:id", watchlistHandler.DeleteWallet())
				
				// Token management
				watchlist.POST("/tokens", watchlistHandler.AddToken())
				watchlist.GET("/tokens", watchlistHandler.GetTokens())
				watchlist.DELETE("/tokens/:id", watchlistHandler.DeleteToken())
				
				// Balance management
				watchlist.GET("/balances", watchlistHandler.GetBalances())
				watchlist.POST("/balances/refresh", watchlistHandler.RefreshBalances())
				
				// Balance history
				watchlist.GET("/wallets/:wallet_id/tokens/:token_id/history", watchlistHandler.GetBalanceHistory())
			}
		}
	}

	return router
}
