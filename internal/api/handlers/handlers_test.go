package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cryptoportfolio/internal/repository"
	"cryptoportfolio/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *services.RegisterRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req *services.LoginRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID uint) (*services.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, userID uint, req *services.UpdateUserRequest) (*services.UserResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserResponse), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, opts *repository.QueryOptions) (*repository.PaginatedResult[services.UserResponse], error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[services.UserResponse]), args.Error(1)
}

func (m *MockUserService) SearchUsers(ctx context.Context, query string, opts *repository.QueryOptions) (*repository.PaginatedResult[services.UserResponse], error) {
	args := m.Called(ctx, query, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[services.UserResponse]), args.Error(1)
}

func (m *MockUserService) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockUserService) GenerateJWT(userID uint) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// setupTestHandler creates a handler with mock service
func setupTestHandler() (*Handler, *MockUserService) {
	mockService := &MockUserService{}
	handler := NewHandler(mockService)
	return handler, mockService
}

// setupTestRouter creates a test router with the handler
func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/health", handler.HealthCheck)
	router.POST("/auth/register", handler.Register())
	router.POST("/auth/login", handler.Login())
	router.GET("/users/me", handler.GetCurrentUser())
	router.PUT("/users/me", handler.UpdateUser())
	
	return router
}

func TestRegisterHandler_Success(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123",
		Name:     "Test User",
	}

	expectedResponse := &services.AuthResponse{
		Message: "User registered successfully",
		Token:   "jwt-token",
		User: services.UserResponse{
			ID:        1,
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Mock service call
	mockService.On("Register", mock.Anything, &services.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123",
		Name:     "Test User",
	}).Return(expectedResponse, nil)

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User registered successfully", response["message"])
	assert.Equal(t, "jwt-token", response["token"])

	mockService.AssertExpectations(t)
}

func TestHealthCheck(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}
