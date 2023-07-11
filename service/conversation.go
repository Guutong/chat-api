package service

import (
	"context"

	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/repository"
)

type IConversationService interface {
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

// ConversationService is a Service for conversation
type ConversationService struct {
	repository repository.IConversationRepository
}

// NewConversationService creates a new conversation Service
func NewConversationService(repository repository.IConversationRepository) *ConversationService {
	return &ConversationService{
		repository: repository,
	}
}

// Create a new conversation
func (s *ConversationService) Create(ctx context.Context, conversation *model.Conversation) (*model.Conversation, error) {
	return s.repository.Create(ctx, conversation)
}

// Find a conversation by user id
func (s *ConversationService) FindByUserID(ctx context.Context, userID string) ([]*model.Conversation, error) {
	return s.repository.FindByUserID(ctx, userID)
}

// Find a conversation by user id pagination
func (s *ConversationService) FindByUserIDPagination(ctx context.Context, userID string, page int64, limit int64) ([]*model.Conversation, error) {
	return s.repository.FindByUserIDPagination(ctx, userID, page, limit)
}

// Find a conversation by id
func (s *ConversationService) FindByID(ctx context.Context, id string) (*model.Conversation, error) {
	return s.repository.FindByID(ctx, id)
}

// Join a conversation
func (s *ConversationService) Join(ctx context.Context, conversationID string, userID string) error {
	return s.repository.Join(ctx, conversationID, userID)
}

// Find a conversation by pair of user id
func (s *ConversationService) FindByPair(ctx context.Context, userID string, recipientID string) (*model.Conversation, error) {
	return s.repository.FindByPair(ctx, userID, recipientID)
}
