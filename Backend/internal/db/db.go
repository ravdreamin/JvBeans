package db

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabase initializes and returns a gorm.DB instance.
func InitDatabase() (*gorm.DB, error) {
	// Read database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "jvbeans.db" // Default to a local file
	}

	db, err := gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
