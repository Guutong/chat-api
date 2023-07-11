package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/guutong/chat-backend/middleware"
	"github.com/guutong/chat-backend/model"
	"github.com/olahol/melody"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	_ "github.com/guutong/chat-backend/docs"
)

func connectToDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	db = client.Database("chat_app")
	log.Println("Connected to MongoDB!")
}

var db *mongo.Database

const jwtSecret = "chat_app_secret"

// @title Chat API
// @description This is a sample chat application API.
// @version 1
// @host localhost:8080
// @BasePath /
// @schemes http https
// securityDefinitions:
//
//	Bearer:
//	  type: apiKey
//	  name: Authorization
//	  in: header
func main() {
	connectToDB()

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/register", register)
	r.POST("/login", login)

	authRoute := r.Group("/")
	authRoute.Use(middleware.AuthMiddleware())
	authRoute.GET("/users", listUsers)
	authRoute.POST("/conversations", createConversation)
	authRoute.GET("/conversations/:userId", listConversations)
	authRoute.GET("/messages/:conversationId", getMessages)
	authRoute.POST("/messages/:conversationId", sendMessage)

	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	r.Run(":8080")
}

// Register a new user godoc
// @Summary Register a new user
// @Description Register a new user
// @Security Bearer
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegisterUser true "User object"
// @Success 200 {object} string "message: User registered successfully"
// @Failure 400 {object} string "error: Invalid request payload"
// @Failure 500 {object} string "error: Internal server error"
// @Router /register [post]
func register(c *gin.Context) {
	// Parse the request body
	var registerUser RegisterUser
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Generate a unique ID for the user
	now := time.Now()
	user := model.User{
		Username:       registerUser.Username,
		Password:       registerUser.Password,
		ProfilePicture: registerUser.ProfilePicture,
		CreateAt:       &now,
		UpdateAt:       &now,
	}

	// Save the user to the MongoDB collection
	collection := db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func listUsers(c *gin.Context) {
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var users []model.User
	if err := cur.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func login(c *gin.Context) {
	// Parse the request body
	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Perform user authentication (e.g., check credentials against the database)
	// You can use the user's input to query the user document from the MongoDB collection
	collection := db.Collection("users")
	var user model.User

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := collection.FindOne(ctx, bson.M{"username": loginData.Username, "password": loginData.Password}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	tokenString, err := generateToken(&user, jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return the token in the response
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func generateToken(user *model.User, jwtSecret string) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Token expiration time (1 day)
	}

	// Create the token with the claims and sign it with the secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func createConversation(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse the request body
	var requestData struct {
		RecipientID string `json:"recipientId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if the recipient ID is different from the authenticated user ID
	if requestData.RecipientID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient"})
		return
	}

	// Create a conversation document
	conversation := model.Conversation{
		ID:      primitive.NewObjectID(),
		Members: []string{userID.(string), requestData.RecipientID},
	}

	// Save the conversation to the MongoDB collection
	collection := db.Collection("conversations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, conversation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation created"})
}

func listConversations(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve conversations for the user from the MongoDB collection
	collection := db.Collection("conversations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"members": bson.M{"$in": []string{userID.(string)}}}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var conversations []model.Conversation
	if err := cur.All(ctx, &conversations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func getMessages(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse the conversation ID from the request parameters
	conversationID := c.Param("conversationId")

	// Parse pagination parameters from the query string
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Retrieve messages for the conversation from the MongoDB collection
	collection := db.Collection("messages")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"conversationId": conversationID,
		"$or": []bson.M{
			{"sender": userID.(string)},
			{"receiver": userID.(string)},
		},
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var messages []model.Message
	if err := cur.All(ctx, &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func sendMessage(c *gin.Context) {
	// Get the authenticated user ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse the conversation ID from the request parameters
	conversationID := c.Param("conversationId")

	// Parse the message text from the request body
	var requestData struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Create a new message
	message := model.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
		Sender:         userID.(string),
		Text:           requestData.Text,
		CreateAt:       time.Now(),
	}

	// Save the message to the MongoDB collection
	collection := db.Collection("messages")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

type RegisterUser struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	ProfilePicture string `json:"profilePicture" binding:"required"`
}

// func addUser(userId string, socketId string) {
// 	for _, user := range socketUsers {
// 		if user.UserID == userId {
// 			return
// 		}
// 	}
// 	socketUsers = append(socketUsers, SocketUser{UserID: userId, SocketID: socketId})
// }

// func removeUser(socketId string) {
// 	for i, user := range socketUsers {
// 		if user.SocketID == socketId {
// 			// remove this user
// 			socketUsers = append(socketUsers[:i], socketUsers[i+1:]...)
// 			break
// 		}
// 	}
// }

// func getUser(userId string) (SocketUser, bool) {
// 	for _, user := range socketUsers {
// 		if user.UserID == userId {
// 			return user, true
// 		}
// 	}
// 	return SocketUser{}, false
// }
