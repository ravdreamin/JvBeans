package handler

import (
	"context"
	"net/http"
	"time"

	"codeflow-backend/internal/db"
	"codeflow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TreeNode struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Type     string     `json:"type"` // "space", "vault", or "log"
	Language string     `json:"language,omitempty"`
	Path     string     `json:"path"`
	Children []TreeNode `json:"children,omitempty"`
}

// GetTree returns the complete hierarchical tree structure
func GetTree(c *gin.Context) {
	spaceID := c.Query("spaceId")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spaceId query parameter required"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(spaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch all vaults in this space
	vaultsCollection := db.Database.Collection("vaults")
	vaultsCursor, err := vaultsCollection.Find(ctx, bson.M{"spaceId": objectID, "userId": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
		return
	}
	var vaults []models.Vault
	vaultsCursor.All(ctx, &vaults)

	// Fetch all logs in this space
	logsCollection := db.Database.Collection("logs")
	logsCursor, err := logsCollection.Find(ctx, bson.M{"spaceId": objectID, "userId": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	var logs []models.Log
	logsCursor.All(ctx, &logs)

	// Build tree structure
	tree := buildTree(vaults, logs)
	c.JSON(http.StatusOK, tree)
}

func buildTree(vaults []models.Vault, logs []models.Log) []TreeNode {
	// Create maps for quick lookup
	vaultMap := make(map[primitive.ObjectID]*TreeNode)
	rootVaults := []TreeNode{}

	// Create vault nodes
	for _, vault := range vaults {
		node := TreeNode{
			ID:       vault.ID.Hex(),
			Name:     vault.Name,
			Type:     "vault",
			Path:     vault.Path,
			Children: []TreeNode{},
		}
		vaultMap[vault.ID] = &node

		// If no parent, it's a root vault
		if vault.ParentID == nil {
			rootVaults = append(rootVaults, node)
		}
	}

	// Build vault hierarchy
	for _, vault := range vaults {
		if vault.ParentID != nil {
			if parent, exists := vaultMap[*vault.ParentID]; exists {
				if child, exists := vaultMap[vault.ID]; exists {
					parent.Children = append(parent.Children, *child)
				}
			}
		}
	}

	// Add logs to their respective vaults
	for _, log := range logs {
		logNode := TreeNode{
			ID:       log.ID.Hex(),
			Name:     log.Name,
			Type:     "log",
			Language: log.Language,
			Path:     log.Path,
		}

		if vault, exists := vaultMap[log.VaultID]; exists {
			vault.Children = append(vault.Children, logNode)
		}
	}

	// Update rootVaults with the modified children
	result := []TreeNode{}
	for i, rootVault := range rootVaults {
		if node, exists := vaultMap[vaults[i].ID]; exists {
			rootVault.Children = node.Children
		}
		result = append(result, rootVault)
	}

	return result
}
