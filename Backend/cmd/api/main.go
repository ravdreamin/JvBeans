package main

import (
	"log"
	"os"

	"codeflow-backend/internal/db"
	"codeflow-backend/internal/handler"
	"codeflow-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to MongoDB
	db.ConnectMongoDB()
	defer db.DisconnectMongoDB()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes with optional admin token middleware
	api := r.Group("/api")
	api.Use(middleware.AdminTokenMiddleware())
	{
		// Spaces
		api.GET("/spaces", handler.GetSpaces)
		api.GET("/spaces/:id", handler.GetSpace)
		api.POST("/spaces", handler.CreateSpace)
		api.PUT("/spaces/:id", handler.UpdateSpace)
		api.DELETE("/spaces/:id", handler.DeleteSpace)

		// Vaults
		api.GET("/vaults", handler.GetVaults) // Query: ?spaceId=xxx
		api.GET("/vaults/:id", handler.GetVault)
		api.POST("/vaults", handler.CreateVault)
		api.PUT("/vaults/:id", handler.UpdateVault)
		api.DELETE("/vaults/:id", handler.DeleteVault)

		// Logs
		api.GET("/logs", handler.GetLogs) // Query: ?spaceId=xxx or ?vaultId=xxx
		api.GET("/logs/:id", handler.GetLog)
		api.POST("/logs", handler.CreateLog)
		api.PUT("/logs/:id", handler.UpdateLog)
		api.DELETE("/logs/:id", handler.DeleteLog)

		// Tree
		api.GET("/tree", handler.GetTree) // Query: ?spaceId=xxx

		// Run code
		api.POST("/run", handler.RunCode)

		// AI generation
		api.POST("/ai/generate", handler.GenerateCode)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✓ Server starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
