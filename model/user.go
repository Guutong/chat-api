package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	Password       string             `bson:"password" json:"-"`
	ProfilePicture string             `bson:"profilePicture" json:"profilePicture"`
	CreateAt       *time.Time         `bson:"createAt" json:"createAt"`
	UpdateAt       *time.Time         `bson:"updateAt" json:"updateAt"`
}
