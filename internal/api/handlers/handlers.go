package handlers

import (
	"net/http"
	"time"

	"simple_api/internal/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService services.UserService
}

func NewHandler(userService services.UserService) *Handler {
	return &Handler{userService: userService}
}

// Simple error response helper
func errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
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

// Request types for Swagger documentation
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"Password123"`
	Name     string `json:"name" binding:"required,min=2" example:"John Doe"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"Password123"`
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,min=2" example:"John Doe Updated"`
}

// Response types for Swagger documentation
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

type UserProfileResponse struct {
	User UserResponse `json:"user"`
}

type UpdateUserResponse struct {
	Message string       `json:"message" example:"User updated successfully"`
	User    UserResponse `json:"user"`
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

		// Convert to service request
		serviceReq := &services.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
			Name:     req.Name,
		}

		// Call service layer
		ctx := c.Request.Context()
		response, err := h.userService.Register(ctx, serviceReq)
		if err != nil {
			switch err {
			case services.ErrUserAlreadyExists:
				errorResponse(c, http.StatusConflict, "User with this email already exists")
			case services.ErrInvalidPassword:
				errorResponse(c, http.StatusBadRequest, err.Error())
			default:
				errorResponse(c, http.StatusInternalServerError, "Failed to create user")
			}
			return
		}

		c.JSON(http.StatusCreated, response)
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

		// Convert to service request
		serviceReq := &services.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		}

		// Call service layer
		ctx := c.Request.Context()
		response, err := h.userService.Login(ctx, serviceReq)
		if err != nil {
			switch err {
			case services.ErrInvalidCredentials:
				errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
			default:
				errorResponse(c, http.StatusInternalServerError, "Login failed")
			}
			return
		}

		c.JSON(http.StatusOK, response)
	}
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

		// Call service layer
		ctx := c.Request.Context()
		user, err := h.userService.GetUserByID(ctx, userID.(uint))
		if err != nil {
			switch err {
			case services.ErrUserNotFound:
				errorResponse(c, http.StatusNotFound, "User not found")
			default:
				errorResponse(c, http.StatusInternalServerError, "Failed to retrieve user")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
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

		// Convert to service request
		serviceReq := &services.UpdateUserRequest{
			Name: req.Name,
		}

		// Call service layer
		ctx := c.Request.Context()
		user, err := h.userService.UpdateUser(ctx, userID.(uint), serviceReq)
		if err != nil {
			switch err {
			case services.ErrUserNotFound:
				errorResponse(c, http.StatusNotFound, "User not found")
			default:
				errorResponse(c, http.StatusInternalServerError, "Failed to update user")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User updated successfully",
			"user":    user,
		})
	}
}
