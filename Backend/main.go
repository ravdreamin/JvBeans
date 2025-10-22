package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CodeRequest is the structure of the request from the frontend.
type CodeRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// PistonRequest is the structure for the Piston API.
type PistonRequest struct {
	Language string              `json:"language"`
	Version  string              `json:"version"`
	Files    []map[string]string `json:"files"`
}

func main() {
	router := gin.Default()

	// Configure CORS
	// This allows requests from your Vercel frontend.
	// You can be more restrictive by specifying origins.
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // For development; for production, list your Vercel URL
	// config.AllowOrigins = []string{"https://your-frontend.vercel.app"}
	config.AllowMethods = []string{"POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "API is running"})
	})

	router.POST("/run", func(c *gin.Context) {
		var req CodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Use "latest" for the version as specified
		payload := PistonRequest{
			Language: req.Language,
			Version:  "latest",
			Files: []map[string]string{
				// The file must be named "Main.java" for this to work
				// with standard `public class Main` Java code.
				{"name": "Main.java", "content": req.Code},
			},
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request payload"})
			return
		}

		// Call the Piston API
		resp, err := http.Post("https://emkc.org/api/v2/piston/execute", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Piston API"})
			return
		}
		defer resp.Body.Close()

		// Decode the JSON response from Piston
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Piston response"})
			return
		}

		// Forward the Piston response to the frontend
		c.JSON(http.StatusOK, result)
	})

	// Run the server on port 8080
	// Render/Railway will automatically use the PORT env var if set,
	// but 8080 is a common default.
	router.Run(":8080")
}
