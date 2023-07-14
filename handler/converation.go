package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/service"
)

// CreateConversation is a struct for creating a new conversation
type CreateConversation struct {
	RecipientID string `json:"recipientId" binding:"required"`
}

type ConversationResponse struct {
	ID            string         `json:"id"`
	Members       []model.User   `json:"members"`
	CreateAt      *time.Time     `json:"createAt"`
	LatestMessage *model.Message `json:"latestMessage"`
	Recipient     *model.User    `json:"recipient"`
}

// IConversationHandler is an interface for conversation handlers
type IConversationHandler interface {
	// Create a new conversation
	Create(c *gin.Context)

	// List conversations by user
	GetAllConversationsByUser(c *gin.Context)

	// Join a conversation
	Join(c *gin.Context)
}

// ConversationHandler is a handler for conversation
type ConversationHandler struct {
	service        service.IConversationService
	userService    service.IUserService
	messageService service.IMessageService
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(
	service service.IConversationService,
	userService service.IUserService,
	messageService service.IMessageService,
) *ConversationHandler {
	return &ConversationHandler{
		service:        service,
		userService:    userService,
		messageService: messageService,
	}
}

// Create a new conversation godoc
// @Summary Create a new conversation
// @Description Create a new conversation
// @Security Bearer
// @Tags conversations
// @Accept json
// @Produce json
// @Param conversation body CreateConversation true "Create Conversation"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/conversations [post]
func (h *ConversationHandler) Create(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var createConversation CreateConversation
	if err := c.ShouldBindJSON(&createConversation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if the recipient ID is different from the authenticated user ID
	if createConversation.RecipientID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient"})
		return
	}

	// check if pair conversation already exists return pair conversation
	conversation, err := h.service.FindByPair(c, userID.(string), createConversation.RecipientID)
	if err == nil {
		c.JSON(http.StatusOK, conversation)
		return
	}

	user, err := h.userService.FindByID(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the recipient exists
	recipient, err := h.userService.FindByID(c, createConversation.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient"})
		return
	}

	user.Password = ""
	recipient.Password = ""

	create := &model.Conversation{
		Members: []model.User{*user, *recipient},
	}

	// Create a new conversation
	created, err := h.service.Create(c, create)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ConversationResponse{
		ID:            created.ID.Hex(),
		Members:       created.Members,
		CreateAt:      created.CreateAt,
		LatestMessage: nil,
		Recipient:     recipient,
	})
}

// List conversations by user godoc
// @Summary List conversations by user
// @Description List conversations by user
// @Security Bearer
// @Tags conversations
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} string "ok"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/conversations [get]
func (h *ConversationHandler) GetAllConversationsByUser(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conversations, err := h.service.FindByUserID(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the latest message of each conversation
	responses := make([]ConversationResponse, len(conversations))
	for i, conversation := range conversations {
		latestMessage, _ := h.messageService.FindLastMessageByConversationID(c, conversation.ID.Hex())
		recipient := conversation.Members[0]
		if recipient.ID.Hex() == userID.(string) {
			recipient = conversation.Members[1]
		}

		responses[i] = ConversationResponse{
			ID:            conversation.ID.Hex(),
			Members:       conversation.Members,
			CreateAt:      conversation.CreateAt,
			LatestMessage: latestMessage,
			Recipient:     &recipient,
		}
	}

	c.JSON(http.StatusOK, responses)
}

// Join a conversation godoc
// @Summary Join a conversation
// @Description Join a conversation
// @Security Bearer
// @Tags conversations
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation ID"
// @Success 200 {object} string "ok"
// @Failure 500 {object} string "Internal server error"
// @Router /api/conversations/{conversationId}/join [post]
func (h *ConversationHandler) Join(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conversationID := c.Param("conversationId")
	if err := h.service.Join(c, conversationID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.service.FindByID(c, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	latestMessage, _ := h.messageService.FindLastMessageByConversationID(c, conversation.ID.Hex())
	recipient := conversation.Members[0]
	if recipient.ID.Hex() == userID.(string) {
		recipient = conversation.Members[1]
	}

	c.JSON(http.StatusOK, ConversationResponse{
		ID:            conversation.ID.Hex(),
		Members:       conversation.Members,
		CreateAt:      conversation.CreateAt,
		LatestMessage: latestMessage,
		Recipient:     &recipient,
	})
}
