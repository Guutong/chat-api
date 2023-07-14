package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/guutong/chat-backend/model"
	"github.com/guutong/chat-backend/service"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser is a struct for registering a new user
type RegisterUser struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	ProfilePicture string `json:"profilePicture" binding:"required"`
}

// LoginUser is a struct for logging in a user
type LoginUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// IUserHandler is an interface for user handlers
type IUserHandler interface {
	// Register a new user
	Register(c *gin.Context)

	// Login a user
	Login(c *gin.Context)

	// Get all users
	GetAll(c *gin.Context)

	// Get user by id
	GetUserByID(c *gin.Context)
}

// UserHandler is a handler for user
type UserHandler struct {
	service service.IUserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service service.IUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Register a new user godoc
// @Summary Register a new user
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegisterUser true "Register User"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	// Parse the request body
	var registerUser RegisterUser
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user := model.User{
		Username:       registerUser.Username,
		Password:       hashedPassword(registerUser.Password),
		ProfilePicture: registerUser.ProfilePicture,
	}

	err := h.service.Register(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// Login a user godoc
// @Summary Login a user
// @Description Login a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body LoginUser true "Login User"
// @Success 200 {object} string "ok"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var loginUser LoginUser
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := h.service.FindByUsername(c, loginUser.Username)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user.ID.Hex(), os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetAll godoc
// @Summary Get all users
// @Description Get all users
// @Security Bearer
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.FindAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := []model.User{}
	for _, user := range users {
		if user.ID.Hex() != c.GetString("userId") {
			responses = append(responses, *user)
		}
	}

	c.JSON(http.StatusOK, responses)
}

func hashedPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return string(hashedPassword)
}

func generateToken(userID string, jwtSecret string) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"userId": userID,
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

// GetUserByID godoc
// @Summary Get user by id
// @Description Get user by id
// @Security Bearer
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} string "ok"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.service.FindByID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
