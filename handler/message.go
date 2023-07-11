package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateMessage struct {
	Text string `json:"text" binding:"required"`
}

// IMessageHandler is an interface for message handlers
type IMessageHandler interface {
	// Create a new message
	Create(c *gin.Context)

	// List messages by conversation
	ListMessagesByConversation(c *gin.Context)

	// List messages by conversation pagination
	ListMessagesByConversationPagination(c *gin.Context)
}

// MessageHandler is a handler for message
type MessageHandler struct {
	service service.IMessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(service service.IMessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// Create a new message godoc
// @Summary Create a new message
// @Description Create a new message
// @Security Bearer
// @Tags messages
// @Accept json
// @Produce json
// @Param message body CreateMessage true "Create Message"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/conversations/{conversationId}/messages [post]
func (h *MessageHandler) Create(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conversationID := c.Param("conversationId")
	var createMessage CreateMessage
	if err := c.ShouldBindJSON(&createMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Create a new message
	message := model.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
		Sender:         userID.(string),
		Text:           createMessage.Text,
		CreateAt:       time.Now(),
	}

	err := h.service.Create(context.Background(), &message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

// List messages by conversation godoc
// @Summary List messages by conversation
// @Description List messages by conversation
// @Security Bearer
// @Tags messages
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation ID"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/conversations/{conversationId}/messages [get]
func (h *MessageHandler) ListMessagesByConversation(c *gin.Context) {
	conversationID := c.Param("conversationId")

	messages, err := h.service.FindByConversationID(context.Background(), conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// List messages by conversation pagination godoc
// @Summary List messages by conversation pagination
// @Description List messages by conversation pagination
// @Security Bearer
// @Tags messages
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation ID"
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/conversations/{conversationId}/messages/pagination [get]
func (h *MessageHandler) ListMessagesByConversationPagination(c *gin.Context) {
	conversationID := c.Param("conversationId")
	pageQuery, exists := c.GetQuery("page")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	limitQuery, exists := c.GetQuery("limit")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	page, err := strconv.ParseInt(pageQuery, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	limit, err := strconv.ParseInt(limitQuery, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	messages, err := h.service.FindByConversationIDPagination(context.Background(), conversationID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
