package service

import (
	"context"
	"errors"

	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/repository"
)

type IUserService interface {
	// Register a new user
	Register(ctx context.Context, user *model.User) error

	// Find a user by username
	FindByUsername(ctx context.Context, username string) (*model.User, error)

	// Find a user by id
	FindByID(ctx context.Context, id string) (*model.User, error)

	// Update a user
	Update(ctx context.Context, user *model.User) error

	// Find all users
	FindAll(ctx context.Context) ([]*model.User, error)
}

// UserService is a service for user
type UserService struct {
	repository repository.IUserRepository
}

// NewUserService creates a new user service
func NewUserService(repository repository.IUserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
}

// Create a new user
func (s *UserService) Register(ctx context.Context, user *model.User) error {
	// Generate a unique ID for the user
	existsUser, _ := s.FindByUsername(ctx, user.Username)
	if existsUser != nil {
		return errors.New("username already exists")
	}

	return s.repository.Create(ctx, user)
}

// Find a user by username
func (s *UserService) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.repository.FindByUsername(ctx, username)
}

// Find a user by id
func (s *UserService) FindByID(ctx context.Context, id string) (*model.User, error) {
	return s.repository.FindByID(ctx, id)
}

// Update a user
func (s *UserService) Update(ctx context.Context, user *model.User) error {
	return s.repository.Update(ctx, user)
}

// Find all users
func (s *UserService) FindAll(ctx context.Context) ([]*model.User, error) {
	return s.repository.FindAll(ctx)
}
