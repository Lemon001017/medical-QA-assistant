package api

import (
	"medical-qa-assistant/internal/handlers"
	"medical-qa-assistant/internal/middleware"
	"medical-qa-assistant/internal/repositories"
	"medical-qa-assistant/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtSecret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Public routes
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// Add protected routes here in the future
		// Example: protected.GET("/profile", userHandler.GetProfile)
	}

	return router
}
