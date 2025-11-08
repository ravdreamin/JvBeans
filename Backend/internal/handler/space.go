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

const userID = "admin" // Single-tenant admin mode

// CreateSpace creates a new space
func CreateSpace(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	space := models.Space{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := db.Database.Collection("spaces")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, space)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create space"})
		return
	}

	c.JSON(http.StatusCreated, space)
}

// GetSpaces retrieves all spaces for the admin user
func GetSpaces(c *gin.Context) {
	collection := db.Database.Collection("spaces")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spaces"})
		return
	}
	defer cursor.Close(ctx)

	var spaces []models.Space
	if err := cursor.All(ctx, &spaces); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode spaces"})
		return
	}

	if spaces == nil {
		spaces = []models.Space{}
	}

	c.JSON(http.StatusOK, spaces)
}

// GetSpace retrieves a single space by ID
func GetSpace(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	collection := db.Database.Collection("spaces")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var space models.Space
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "userId": userID}).Decode(&space)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		return
	}

	c.JSON(http.StatusOK, space)
}

// UpdateSpace updates a space
func UpdateSpace(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := db.Database.Collection("spaces")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":      req.Name,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID, "userId": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update space"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		return
	}

	var space models.Space
	collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&space)
	c.JSON(http.StatusOK, space)
}

// DeleteSpace deletes a space and all its vaults and logs
func DeleteSpace(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete all logs in this space
	logsCollection := db.Database.Collection("logs")
	logsCollection.DeleteMany(ctx, bson.M{"spaceId": objectID})

	// Delete all vaults in this space
	vaultsCollection := db.Database.Collection("vaults")
	vaultsCollection.DeleteMany(ctx, bson.M{"spaceId": objectID})

	// Delete the space
	spacesCollection := db.Database.Collection("spaces")
	result, err := spacesCollection.DeleteOne(ctx, bson.M{"_id": objectID, "userId": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete space"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Space deleted successfully"})
}
