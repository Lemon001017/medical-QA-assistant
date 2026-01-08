package api

import (
	"medical-qa-assistant/internal/config"
	"medical-qa-assistant/internal/handlers"
	"medical-qa-assistant/internal/middleware"
	"medical-qa-assistant/internal/repositories"
	"medical-qa-assistant/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) *gin.Engine {
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
	documentRepo := repositories.NewDocumentRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	documentService := services.NewDocumentService(documentRepo)

	var qaService *services.QAService
	switch cfg.LLMProvider {
	case "deepseek":
		qaService = services.NewQAService(cfg.DeepSeekKey, cfg.DeepSeekModel, cfg.DeepSeekBaseURL)
	default:
		qaService = services.NewQAService(cfg.OpenAIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	qaHandler := handlers.NewQAHandler(qaService)

	// Public routes
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/documents", documentHandler.Create)
		protected.POST("/documents/upload", documentHandler.Upload)
		protected.GET("/documents", documentHandler.List)
		protected.GET("/documents/:id", documentHandler.Get)
		protected.DELETE("/documents/:id", documentHandler.Delete)
		protected.POST("/qa/ask", qaHandler.Ask)
		protected.POST("/qa/ask/stream", qaHandler.AskStream)
	}

	return router
}
