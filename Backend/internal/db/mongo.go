package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Database *mongo.Database

// ConnectMongoDB initializes the MongoDB connection
func ConnectMongoDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable not set")
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "codeflow"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(30 * time.Second).
		SetConnectTimeout(30 * time.Second)
	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	if err := Client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	Database = Client.Database(dbName)
	log.Printf("Connected to MongoDB database: %s", dbName)

	// Create indexes
	createIndexes()
}

// createIndexes creates necessary indexes for collections
func createIndexes() {
	ctx := context.Background()

	// Spaces collection indexes
	spacesCollection := Database.Collection("spaces")
	spacesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"userId": 1},
	})

	// Vaults collection indexes
	vaultsCollection := Database.Collection("vaults")
	vaultsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: map[string]interface{}{"spaceId": 1}},
		{Keys: map[string]interface{}{"userId": 1}},
		{Keys: map[string]interface{}{"path": 1}},
	})

	// Logs collection indexes
	logsCollection := Database.Collection("logs")
	logsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: map[string]interface{}{"spaceId": 1}},
		{Keys: map[string]interface{}{"vaultId": 1}},
		{Keys: map[string]interface{}{"userId": 1}},
		{Keys: map[string]interface{}{"path": 1}},
	})

	log.Println("Database indexes created successfully")
}

// DisconnectMongoDB closes the MongoDB connection
func DisconnectMongoDB() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}
