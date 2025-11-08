package handler

import (
	"context"
	"net/http"
	"path"
	"time"

	"codeflow-backend/internal/db"
	"codeflow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateVault creates a new vault
func CreateVault(c *gin.Context) {
	var req struct {
		SpaceID  string  `json:"spaceId" binding:"required"`
		Name     string  `json:"name" binding:"required"`
		ParentID *string `json:"parentId,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spaceID, err := primitive.ObjectIDFromHex(req.SpaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	vault := models.Vault{
		ID:        primitive.NewObjectID(),
		SpaceID:   spaceID,
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Handle parent vault and path
	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := primitive.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent ID"})
			return
		}
		vault.ParentID = &parentID

		// Get parent vault to construct path
		collection := db.Database.Collection("vaults")
		ctx := context.Background()
		var parentVault models.Vault
		err = collection.FindOne(ctx, bson.M{"_id": parentID}).Decode(&parentVault)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent vault not found"})
			return
		}
		vault.Path = path.Join(parentVault.Path, req.Name)
	} else {
		vault.Path = req.Name
	}

	collection := db.Database.Collection("vaults")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, vault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vault"})
		return
	}

	c.JSON(http.StatusCreated, vault)
}

// GetVaults retrieves all vaults for a space
func GetVaults(c *gin.Context) {
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

	collection := db.Database.Collection("vaults")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"spaceId": objectID, "userId": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
		return
	}
	defer cursor.Close(ctx)

	var vaults []models.Vault
	if err := cursor.All(ctx, &vaults); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode vaults"})
		return
	}

	if vaults == nil {
		vaults = []models.Vault{}
	}

	c.JSON(http.StatusOK, vaults)
}

// GetVault retrieves a single vault by ID
func GetVault(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vault ID"})
		return
	}

	collection := db.Database.Collection("vaults")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var vault models.Vault
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "userId": userID}).Decode(&vault)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
		return
	}

	c.JSON(http.StatusOK, vault)
}

// UpdateVault updates a vault
func UpdateVault(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vault ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := db.Database.Collection("vaults")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current vault to update path
	var currentVault models.Vault
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "userId": userID}).Decode(&currentVault)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
		return
	}

	// Update path
	newPath := req.Name
	if currentVault.ParentID != nil {
		var parentVault models.Vault
		collection.FindOne(ctx, bson.M{"_id": currentVault.ParentID}).Decode(&parentVault)
		newPath = path.Join(parentVault.Path, req.Name)
	}

	update := bson.M{
		"$set": bson.M{
			"name":      req.Name,
			"path":      newPath,
			"updatedAt": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID, "userId": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vault"})
		return
	}

	var vault models.Vault
	collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&vault)
	c.JSON(http.StatusOK, vault)
}

// DeleteVault deletes a vault and all its logs
func DeleteVault(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vault ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete all logs in this vault
	logsCollection := db.Database.Collection("logs")
	logsCollection.DeleteMany(ctx, bson.M{"vaultId": objectID})

	// Delete child vaults recursively
	vaultsCollection := db.Database.Collection("vaults")
	cursor, _ := vaultsCollection.Find(ctx, bson.M{"parentId": objectID})
	var childVaults []models.Vault
	cursor.All(ctx, &childVaults)
	for _, child := range childVaults {
		logsCollection.DeleteMany(ctx, bson.M{"vaultId": child.ID})
	}
	vaultsCollection.DeleteMany(ctx, bson.M{"parentId": objectID})

	// Delete the vault
	result, err := vaultsCollection.DeleteOne(ctx, bson.M{"_id": objectID, "userId": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vault"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vault deleted successfully"})
}
