package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"simple_api/internal/config"
	"simple_api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Config: cfg}
}

// Simple error response helper
func errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// Basic password strength validation
func validatePassword(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters long"
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return "Password must contain at least one uppercase letter"
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return "Password must contain at least one lowercase letter"
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return "Password must contain at least one number"
	}
	return ""
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Server is running",
		"time":    time.Now().UTC(),
	})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *Handler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request data")
			return
		}

		// Validate password strength
		if errMsg := validatePassword(req.Password); errMsg != "" {
			errorResponse(c, http.StatusBadRequest, errMsg)
			return
		}

		// Check if user already exists
		var existingUser models.User
		if err := h.DB.Where("email = ?", strings.ToLower(req.Email)).First(&existingUser).Error; err == nil {
			errorResponse(c, http.StatusConflict, "User with this email already exists")
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to process password")
			return
		}

		// Create user
		user := models.User{
			Email:    strings.ToLower(req.Email),
			Password: string(hashedPassword),
			Name:     strings.TrimSpace(req.Name),
		}

		if err := h.DB.Create(&user).Error; err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Generate JWT token
		token, err := h.generateJWT(user.ID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"token":   token,
			"user": UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

func (h *Handler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request data")
			return
		}

		// Find user by email
		var user models.User
		if err := h.DB.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
			errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Generate JWT token
		token, err := h.generateJWT(user.ID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
			"user": UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

func (h *Handler) GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var user models.User
		if err := h.DB.First(&user, userID).Error; err != nil {
			errorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

func (h *Handler) UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request data")
			return
		}

		var user models.User
		if err := h.DB.First(&user, userID).Error; err != nil {
			errorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		// Update user
		user.Name = strings.TrimSpace(req.Name)
		if err := h.DB.Save(&user).Error; err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to update user")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User updated successfully",
			"user": UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

func (h *Handler) generateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	})
	return token.SignedString([]byte(h.Config.JWT.Secret))
}
