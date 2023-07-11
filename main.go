package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guutong/chat-backend/handler"
	"github.com/guutong/chat-backend/middleware"
	"github.com/guutong/chat-backend/repository"
	"github.com/guutong/chat-backend/service"
	"github.com/olahol/melody"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	// cors allow all
	r.Use(cors.Default())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api")

	userRepository := repository.NewUserRepository(db)
	conversationRepository := repository.NewConversationRepository(db)
	messageRepository := repository.NewMessageRepository(db)

	userService := service.NewUserService(userRepository)
	conversationService := service.NewConversationService(conversationRepository)
	messageService := service.NewMessageService(messageRepository)

	userHandler := handler.NewUserHandler(userService)
	conversationHandler := handler.NewConversationHandler(conversationService)
	messageHandler := handler.NewMessageHandler(messageService)

	userApi := api.Group("/users")
	conversationRoute := api.Group("/conversations")

	userApi.GET("", middleware.AuthMiddleware(), userHandler.GetAll)
	userApi.POST("/register", userHandler.Register)
	userApi.POST("/login", userHandler.Login)
	userApi.GET("/conversations", middleware.AuthMiddleware(), conversationHandler.GetAllConversationsByUser)

	conversationRoute.POST("", middleware.AuthMiddleware(), conversationHandler.Create)
	conversationRoute.POST("/:id/join", middleware.AuthMiddleware(), conversationHandler.Join)
	conversationRoute.POST("/:id/messages", middleware.AuthMiddleware(), messageHandler.Create)
	conversationRoute.GET("/:id/messages", middleware.AuthMiddleware(), messageHandler.ListMessagesByConversation)
	conversationRoute.GET("/:id/messages/pagination", middleware.AuthMiddleware(), messageHandler.ListMessagesByConversationPagination)

	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	r.Run(":8080")
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
