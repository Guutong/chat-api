package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guutong/chat-backend/handler"
	"github.com/guutong/chat-backend/middleware"
	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/repository"
	"github.com/guutong/chat-backend/service"
	"github.com/mitchellh/mapstructure"
	"github.com/olahol/melody"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	_ "github.com/guutong/chat-backend/docs"
)

func connectToDB() {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
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
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api")
	userRepository := repository.NewUserRepository(db)
	conversationRepository := repository.NewConversationRepository(db)
	messageRepository := repository.NewMessageRepository(db)

	userService := service.NewUserService(userRepository)
	conversationService := service.NewConversationService(conversationRepository)
	messageService := service.NewMessageService(messageRepository)

	userHandler := handler.NewUserHandler(userService)
	conversationHandler := handler.NewConversationHandler(conversationService, userService, messageService)
	messageHandler := handler.NewMessageHandler(messageService)

	userApi := api.Group("/users")
	conversationRoute := api.Group("/conversations")

	userApi.GET("", middleware.AuthMiddleware(), userHandler.GetAll)
	userApi.POST("/register", userHandler.Register)
	userApi.POST("/login", userHandler.Login)
	userApi.GET("/conversations", middleware.AuthMiddleware(), conversationHandler.GetAllConversationsByUser)

	conversationRoute.POST("", middleware.AuthMiddleware(), conversationHandler.Create)
	conversationRoute.POST("/:conversationId/join", middleware.AuthMiddleware(), conversationHandler.Join)
	conversationRoute.POST("/:conversationId/messages", middleware.AuthMiddleware(), messageHandler.Create)
	conversationRoute.GET("/:conversationId/messages", middleware.AuthMiddleware(), messageHandler.ListMessagesByConversation)
	conversationRoute.GET("/:conversationId/messages/pagination", middleware.AuthMiddleware(), messageHandler.ListMessagesByConversationPagination)

	// 1 user A see the users list B C D E
	// 2 user A click on a user B
	// 3 create a conversation between 2 users A - B at that time
	// 4 user A send a message "hello" to user B
	//		- frontend send a message to websocket server
	//		- if user B is online, user B receive a message (websocket message and message to pulling new message)
	//		- if user B is offline, user B receive a message (websocket message and message to pulling new message)
	m := melody.New()
	m.Config.MaxMessageSize = 2000
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true } // origni check
	socketUsers := make(map[*melody.Session]*SocketUser)
	lock := new(sync.Mutex)

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		fmt.Println(s.Request.URL.Query().Get("userId"))
		fmt.Println(string(msg))
		var message SocketMessage
		json.Unmarshal(msg, &message)

		switch message.Event {
		case "addUser":
			lock.Lock()
			userID := s.Request.URL.Query().Get("userId")
			newUser := SocketUser{
				ID:   userID,
				UUID: uuid.New().String(),
			}
			s.Set("data", newUser)
			socketUsers[s] = &newUser
			lock.Unlock()

			msg := SocketMessage{
				Event:   "getUsers",
				Message: socketUsers,
			}

			b, _ := json.Marshal(msg)
			m.Broadcast(b)
		case "sendMessage":
			var data SocketData
			_ = mapstructure.Decode(message.Message, &data)

			msg := SocketMessage{
				Event: "getMessage",
				Message: model.Message{
					ID:             primitive.NewObjectID(),
					ConversationID: data.ConversationID,
					Sender:         data.SenderID,
					Text:           data.Text,
					CreateAt:       time.Now(),
				},
			}
			b, _ := json.Marshal(msg)
			m.BroadcastFilter(b, func(q *melody.Session) bool {
				ss, _ := m.Sessions()
				for _, s := range ss {
					if s.Request.URL.Query().Get("userId") == data.RecipientID {
						fmt.Println("send to", s.Request.URL.Query().Get("userId"))
						return true
					}
				}
				return false
			})
		}
	})

	m.HandleDisconnect(func(s *melody.Session) {
		fmt.Println("disconnect")
		lock.Lock()
		socketUsers[s] = nil
		s.UnSet("data")
		lock.Unlock()
	})

	r.Run(":8080")
}

type SocketUser struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

type SocketMessage struct {
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

type SocketData struct {
	ConversationID string `json:"conversationId"`
	SenderID       string `json:"senderId"`
	RecipientID    string `json:"recipientId"`
	Text           string `json:"text"`
}
