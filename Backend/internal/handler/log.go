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

// CreateLog creates a new log (code file)
func CreateLog(c *gin.Context) {
	var req struct {
		SpaceID  string `json:"spaceId" binding:"required"`
		VaultID  string `json:"vaultId" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Code     string `json:"code"`
		Language string `json:"language,omitempty"`
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

	vaultID, err := primitive.ObjectIDFromHex(req.VaultID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vault ID"})
		return
	}

	// Get vault to construct full path
	vaultsCollection := db.Database.Collection("vaults")
	ctx := context.Background()
	var vault models.Vault
	err = vaultsCollection.FindOne(ctx, bson.M{"_id": vaultID}).Decode(&vault)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
		return
	}

	// Infer language from filename if not provided
	language := req.Language
	if language == "" {
		language = models.InferLanguageFromFilename(req.Name)
	}

	log := models.Log{
		ID:        primitive.NewObjectID(),
		SpaceID:   spaceID,
		VaultID:   vaultID,
		UserID:    userID,
		Name:      req.Name,
		Path:      path.Join(vault.Path, req.Name),
		Language:  language,
		Code:      req.Code,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := db.Database.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create log"})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// GetLogs retrieves all logs for a vault or space
func GetLogs(c *gin.Context) {
	spaceID := c.Query("spaceId")
	vaultID := c.Query("vaultId")

	filter := bson.M{"userId": userID}

	if spaceID != "" {
		objectID, err := primitive.ObjectIDFromHex(spaceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
			return
		}
		filter["spaceId"] = objectID
	}

	if vaultID != "" {
		objectID, err := primitive.ObjectIDFromHex(vaultID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vault ID"})
			return
		}
		filter["vaultId"] = objectID
	}

	collection := db.Database.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	defer cursor.Close(ctx)

	var logs []models.Log
	if err := cursor.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode logs"})
		return
	}

	if logs == nil {
		logs = []models.Log{}
	}

	c.JSON(http.StatusOK, logs)
}

// GetLog retrieves a single log by ID
func GetLog(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	collection := db.Database.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var log models.Log
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "userId": userID}).Decode(&log)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// UpdateLog updates a log
func UpdateLog(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	var req struct {
		Name string `json:"name,omitempty"`
		Code string `json:"code,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := db.Database.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current log
	var currentLog models.Log
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "userId": userID}).Decode(&currentLog)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Name != "" {
		// Get vault to construct new path
		vaultsCollection := db.Database.Collection("vaults")
		var vault models.Vault
		vaultsCollection.FindOne(ctx, bson.M{"_id": currentLog.VaultID}).Decode(&vault)

		newPath := path.Join(vault.Path, req.Name)
		language := models.InferLanguageFromFilename(req.Name)

		update["$set"].(bson.M)["name"] = req.Name
		update["$set"].(bson.M)["path"] = newPath
		update["$set"].(bson.M)["language"] = language
	}

	if req.Code != "" {
		update["$set"].(bson.M)["code"] = req.Code
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID, "userId": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update log"})
		return
	}

	var log models.Log
	collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&log)
	c.JSON(http.StatusOK, log)
}

// DeleteLog deletes a log
func DeleteLog(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	collection := db.Database.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID, "userId": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete log"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}
