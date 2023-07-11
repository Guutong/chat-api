package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/guutong/chat-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IConversationRepository interface {
	// Create a new conversation
	Create(ctx context.Context, conversation *model.Conversation) (*model.Conversation, error)

	// Find a conversation by user id
	FindByUserID(ctx context.Context, userID string) ([]*model.Conversation, error)

	// Find a conversation by user id pagination
	FindByUserIDPagination(ctx context.Context, userID string, page int64, limit int64) ([]*model.Conversation, error)

	// Find a conversation by id
	FindByID(ctx context.Context, id string) (*model.Conversation, error)

	// Join a conversation
	Join(ctx context.Context, conversationID string, userID string) error

	// Find a conversation by pair of user id
	FindByPair(ctx context.Context, userID string, recipientID string) (*model.Conversation, error)
}

// ConversationRepository is a repository for conversation
type ConversationRepository struct {
	collection *mongo.Collection
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		collection: db.Collection("conversations"),
	}
}

// Create a new conversation
func (r *ConversationRepository) Create(ctx context.Context, conversation *model.Conversation) (*model.Conversation, error) {
	now := time.Now()
	conversation.CreateAt = &now

	res, err := r.collection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	newID, exists := res.InsertedID.(primitive.ObjectID)
	if !exists {
		return nil, err
	}

	conversation.ID = newID
	return conversation, err
}

// Find a conversation by user id
func (r *ConversationRepository) FindByUserID(ctx context.Context, userID string) ([]*model.Conversation, error) {
	conversations := []*model.Conversation{}
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{
		"members": bson.M{
			"$elemMatch": bson.M{
				"_id": bson.M{
					"$in": []primitive.ObjectID{id},
				},
			},
		},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

// Find a conversation by user id pagination
func (r *ConversationRepository) FindByUserIDPagination(ctx context.Context, userID string, page int64, limit int64) ([]*model.Conversation, error) {
	var conversations []*model.Conversation
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{
		"members": bson.M{
			"$elemMatch": bson.M{
				"_id": bson.M{
					"$in": []primitive.ObjectID{id},
				},
			},
		},
	}

	opts := &options.FindOptions{
		Skip:  &page,
		Limit: &limit,
		Sort: bson.M{
			"created_at": -1,
		},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

// Find a conversation by id
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*model.Conversation, error) {
	var conversation *model.Conversation
	filter := bson.M{
		"_id": id,
	}
	if err := r.collection.FindOne(ctx, filter).Decode(&conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

// Join a conversation
func (r *ConversationRepository) Join(ctx context.Context, conversationID string, userID string) error {
	filter := bson.M{
		"_id": conversationID,
	}
	update := bson.M{
		"$addToSet": bson.M{
			"members": userID,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Find a conversation by pair of user id
func (r *ConversationRepository) FindByPair(ctx context.Context, userID string, recipientID string) (*model.Conversation, error) {
	var conversation *model.Conversation
	id, _ := primitive.ObjectIDFromHex(userID)
	recipient, _ := primitive.ObjectIDFromHex(recipientID)
	filter := bson.M{
		"members._id": bson.M{
			"$all": []primitive.ObjectID{id, recipient},
		},
	}

	// filter := bson.M{
	// 	"members": bson.M{
	// 		"$all": []string{userID, recipientID},
	// 	},
	// }
	if err := r.collection.FindOne(ctx, filter).Decode(&conversation); err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(conversation)
	if conversation == nil {
		return nil, errors.New("conversation not found")
	}

	return conversation, nil
}
