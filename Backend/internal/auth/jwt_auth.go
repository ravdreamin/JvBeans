package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetJWTSecret retrieves the JWT secret key from environment variables
var jwtSecret = getJWTSecret()

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if os.Getenv("GIN_MODE") == "release" {
			log.Fatal("FATAL: JWT_SECRET environment variable not set in release mode")
		}
		// Use a default for development mode only
		return []byte("a_very_unsafe_development_secret_for_dev_only")
	}
	return []byte(secret)
}

// AppClaims defines the custom claims for the JWT
type AppClaims struct {
	UserID uint `json:"userID"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT for a given user ID
func GenerateJWT(userID uint) (string, error) {
	claims := AppClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // Token expires in 3 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates a token string and returns the user ID
func ValidateJWT(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
