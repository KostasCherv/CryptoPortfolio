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
// @Summary Health check endpoint
// @Description Check if the server is running and healthy
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Server status information"
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Server is running",
		"time":    time.Now().UTC(),
	})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"Password123"`
	Name     string `json:"name" binding:"required,min=2" example:"John Doe"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"Password123"`
}

type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type AuthResponse struct {
	Message string       `json:"message" example:"User registered successfully"`
	Token   string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    UserResponse `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request data"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email, password, and name
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data or weak password"
// @Failure 409 {object} ErrorResponse "User with this email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
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

// Login handles user authentication
// @Summary Authenticate user
// @Description Login with email and password to receive JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
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

type UserProfileResponse struct {
	User UserResponse `json:"user"`
}

// GetCurrentUser retrieves the current authenticated user's profile
// @Summary Get current user profile
// @Description Retrieve the profile information of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/v1/users/me [get]
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
	Name string `json:"name" binding:"required,min=2" example:"John Doe Updated"`
}

type UpdateUserResponse struct {
	Message string       `json:"message" example:"User updated successfully"`
	User    UserResponse `json:"user"`
}

// UpdateUser updates the current authenticated user's profile
// @Summary Update current user profile
// @Description Update the name of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserRequest true "User update data"
// @Success 200 {object} UpdateUserResponse "User updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users/me [put]
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
