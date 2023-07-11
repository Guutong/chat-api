package repository

import (
	"context"
	"time"

	"github.com/guutong/chat-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IUserRepository interface {
	// Create a new user
	Create(ctx context.Context, user *model.User) error

	// Find a user by username
	FindByUsername(ctx context.Context, username string) (*model.User, error)

	// Find a user by id
	FindByID(ctx context.Context, id string) (*model.User, error)

	// Update a user
	Update(ctx context.Context, user *model.User) error

	// Find all users
	FindAll(ctx context.Context) ([]*model.User, error)
}

// UserRepository is a repository for user
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	now := time.Now()
	user.CreateAt = &now
	user.UpdateAt = &now
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// Find a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	filter := bson.M{"username": username}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// Find a user by id
func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	err = r.collection.FindOne(ctx, filter).Decode(&user)
	return &user, err
}

// Update a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"profilePicture": user.ProfilePicture}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Find all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	filter := bson.M{}
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []*model.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
