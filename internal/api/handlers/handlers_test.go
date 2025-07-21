package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple_api/internal/services"

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

func TestRegisterHandler_InvalidRequest(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"email": "invalid-email", // Invalid email
		"name":  "Test User",
		// Missing password
	}

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request data", response["error"])

	// Service should not be called for invalid requests
	mockService.AssertNotCalled(t, "Register")
}

func TestRegisterHandler_UserAlreadyExists(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	reqBody := RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123",
		Name:     "Existing User",
	}

	// Mock service call to return error
	mockService.On("Register", mock.Anything, &services.RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123",
		Name:     "Existing User",
	}).Return(nil, services.ErrUserAlreadyExists)

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User with this email already exists", response["error"])

	mockService.AssertExpectations(t)
}

func TestLoginHandler_Success(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}

	expectedResponse := &services.AuthResponse{
		Message: "Login successful",
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
	mockService.On("Login", mock.Anything, &services.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}).Return(expectedResponse, nil)

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
	assert.Equal(t, "jwt-token", response["token"])

	mockService.AssertExpectations(t)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}

	// Mock service call to return error
	mockService.On("Login", mock.Anything, &services.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}).Return(nil, services.ErrInvalidCredentials)

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	mockService.AssertExpectations(t)
}

func TestGetCurrentUser_Success(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()

	expectedUser := &services.UserResponse{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock service call
	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(expectedUser, nil)

	// Act
	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Create a custom gin context with user_id set
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	
	// Call the handler directly
	handler.GetCurrentUser()(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["user"])

	mockService.AssertExpectations(t)
}

func TestGetCurrentUser_NotAuthenticated(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Act
	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User not authenticated", response["error"])

	// Service should not be called
	mockService.AssertNotCalled(t, "GetUserByID")
}

func TestUpdateUser_Success(t *testing.T) {
	// Arrange
	handler, mockService := setupTestHandler()

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
	}

	expectedUser := &services.UserResponse{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Updated Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock service call
	mockService.On("UpdateUser", mock.Anything, uint(1), &services.UpdateUserRequest{
		Name: "Updated Name",
	}).Return(expectedUser, nil)

	// Act
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Create a custom gin context with user_id set
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	
	// Call the handler directly
	handler.UpdateUser()(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User updated successfully", response["message"])
	assert.NotNil(t, response["user"])

	mockService.AssertExpectations(t)
}

func TestHealthCheck(t *testing.T) {
	// Arrange
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	// Act
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Server is running", response["message"])
	assert.NotNil(t, response["time"])
} 