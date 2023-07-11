package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Members  []User             `bson:"members" json:"members"`
	CreateAt *time.Time         `bson:"createAt" json:"createAt"`
}
