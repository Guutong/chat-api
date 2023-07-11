package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversationId" json:"conversationId"`
	Sender         string             `bson:"sender" json:"sender"`
	Text           string             `bson:"text" json:"text"`
	CreateAt       time.Time          `bson:"createAt" json:"createAt"`
}
