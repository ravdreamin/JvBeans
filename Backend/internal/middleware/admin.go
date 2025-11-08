package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AdminTokenMiddleware checks for optional ADMIN_TOKEN on write operations
func AdminTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check on write methods (POST, PUT, DELETE)
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// If ADMIN_TOKEN is not set in env, skip check
		expectedToken := os.Getenv("ADMIN_TOKEN")
		if expectedToken == "" {
			c.Next()
			return
		}

		// Check if Authorization header matches
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer "+expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing admin token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
