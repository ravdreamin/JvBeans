package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vault represents a folder/directory within a space
type Vault struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SpaceID   primitive.ObjectID `bson:"spaceId" json:"spaceId"`
	UserID    string             `bson:"userId" json:"userId"`
	Name      string             `bson:"name" json:"name"`
	Path      string             `bson:"path" json:"path"` // Full path like "vault1" or "vault1/subvault"
	ParentID  *primitive.ObjectID `bson:"parentId,omitempty" json:"parentId,omitempty"` // For nested vaults
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
