package models

import (
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Log represents a code file within a vault
type Log struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SpaceID   primitive.ObjectID `bson:"spaceId" json:"spaceId"`
	VaultID   primitive.ObjectID `bson:"vaultId" json:"vaultId"`
	UserID    string             `bson:"userId" json:"userId"`
	Name      string             `bson:"name" json:"name"` // filename like "app.js"
	Path      string             `bson:"path" json:"path"` // Full path like "vault1/app.js"
	Language  string             `bson:"language" json:"language"`
	Code      string             `bson:"code" json:"code"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// InferLanguageFromFilename determines language from file extension
func InferLanguageFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".go":
		return "go"
	case ".ts":
		return "typescript"
	case ".rs":
		return "rust"
	default:
		return "javascript" // default fallback
	}
}
