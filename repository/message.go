package repository

import (
	"context"

	"github.com/guutong/chat-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMessageRepository interface {
	// Create a new message
	Create(ctx context.Context, message *model.Message) error

	// Find a message by conversation id
	FindByConversationID(ctx context.Context, conversationID string) ([]*model.Message, error)

	// Find a message by conversation id pagination
	FindByConversationIDPagination(ctx context.Context, conversationID string, page int64, limit int64) ([]*model.Message, error)

	// Find last message by conversation id
	FindLastMessageByConversationID(ctx context.Context, conversationID string) (*model.Message, error)
}

// MessageRepository is a repository for message
type MessageRepository struct {
	collection *mongo.Collection
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

// Create a new message
func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	_, err := r.collection.InsertOne(ctx, message)
	return err
}

// Find a message by conversation id
func (r *MessageRepository) FindByConversationID(ctx context.Context, conversationID string) ([]*model.Message, error) {
	var messages []*model.Message

	filter := bson.M{"conversationId": conversationID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Find a message by conversation id pagination
func (r *MessageRepository) FindByConversationIDPagination(ctx context.Context, conversationID string, page int64, limit int64) ([]*model.Message, error) {
	var messages []*model.Message

	filter := bson.M{"conversationId": conversationID}
	opts := &options.FindOptions{
		Skip:  &page,
		Limit: &limit,
		Sort:  map[string]int{"createAt": -1},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Find latest message by conversation id
func (r *MessageRepository) FindLastMessageByConversationID(ctx context.Context, conversationID string) (*model.Message, error) {
	var message *model.Message

	filter := bson.M{"conversationId": conversationID}
	opts := &options.FindOptions{
		Sort: map[string]int{"createAt": -1},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &message); err != nil {
		return nil, err
	}

	return message, nil
}
