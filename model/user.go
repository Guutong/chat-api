package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	Password       string             `bson:"password" json:"password"`
	ProfilePicture string             `bson:"profilePicture" json:"profilePicture"`
	CreateAt       *time.Time         `bson:"createAt" json:"createAt"`
	UpdateAt       *time.Time         `bson:"updateAt" json:"updateAt"`
}

type Conversation struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Members  []string           `bson:"members" json:"members"`
	CreateAt *time.Time         `bson:"createAt" json:"createAt"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversationId" json:"conversationId"`
	Sender         string             `bson:"sender" json:"sender"` // user id
	Text           string             `bson:"text" json:"text"`
	CreateAt       time.Time          `bson:"createAt" json:"createAt"`
}
