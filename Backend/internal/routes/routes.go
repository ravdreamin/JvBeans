package routes

import (
	"jvbeans/internal/handler"
	"jvbeans/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all the API routes for the application
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // In production, restrict this
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Create a new API handler instance
	h := handler.NewHandler(db)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		v1.POST("/register", h.RegisterUser)
		v1.POST("/login", h.LoginUser)

		// Problem routes (protected)
		problems := v1.Group("/problems")
		problems.Use(middleware.AuthMiddleware()) // Protect these routes
		{
			problems.GET("/categories", h.GetCategories)
			problems.GET("/:id", h.GetProblem)
			problems.GET("/:id/solution", h.GetSolution)

			// Admin-only routes (example)
			// admin := problems.Group("/")
			// admin.Use(AdminMiddleware()) // A potential second middleware
			// {
			// 	admin.POST("/", h.CreateProblem)
			// }
		}

		// Code execution route (protected)
		v1.POST("/run", middleware.AuthMiddleware(), h.RunCode)
	}

	// Health check
	router.GET("/", h.HealthCheck)
}
