package service

import (
	"context"

	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/repository"
)

type IMessageService interface {
	// Create a new message
	Create(ctx context.Context, message *model.Message) error

	// Find a message by conversation id
	FindByConversationID(ctx context.Context, conversationID string) ([]*model.Message, error)

	// Find a message by conversation id pagination
	FindByConversationIDPagination(ctx context.Context, conversationID string, page int64, limit int64) ([]*model.Message, error)
}

// MessageService is a service for message
type MessageService struct {
	repository repository.IMessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repository repository.IMessageRepository) *MessageService {
	return &MessageService{
		repository: repository,
	}
}

// Create a new message
func (s *MessageService) Create(ctx context.Context, message *model.Message) error {
	return s.repository.Create(ctx, message)
}

// Find a message by conversation id
func (s *MessageService) FindByConversationID(ctx context.Context, conversationID string) ([]*model.Message, error) {
	return s.repository.FindByConversationID(ctx, conversationID)
}

// Find a message by conversation id pagination
func (s *MessageService) FindByConversationIDPagination(ctx context.Context, conversationID string, page int64, limit int64) ([]*model.Message, error) {
	return s.repository.FindByConversationIDPagination(ctx, conversationID, page, limit)
}
