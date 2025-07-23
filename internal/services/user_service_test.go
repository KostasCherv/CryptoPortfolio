package services

import (
	"context"
	"testing"
	"time"

	"cryptoportfolio/internal/cache"
	"cryptoportfolio/internal/config"
	"cryptoportfolio/internal/models"
	"cryptoportfolio/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// MockUserCache implements cache.UserCacheProvider for testing
type MockUserCache struct {
	users map[uint]*models.User
	emails map[string]*models.User
}

func NewMockUserCache() *MockUserCache {
	return &MockUserCache{
		users: make(map[uint]*models.User),
		emails: make(map[string]*models.User),
	}
}

func (m *MockUserCache) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, cache.ErrCacheMiss
}

func (m *MockUserCache) SetUserByID(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserCache) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, exists := m.emails[email]; exists {
		return user, nil
	}
	return nil, cache.ErrCacheMiss
}

func (m *MockUserCache) SetUserByEmail(ctx context.Context, user *models.User) error {
	m.emails[user.Email] = user
	return nil
}

func (m *MockUserCache) InvalidateUser(ctx context.Context, userID uint, email string) error {
	delete(m.users, userID)
	delete(m.emails, email)
	return nil
}

func (m *MockUserCache) InvalidateAllUsers(ctx context.Context) error {
	m.users = make(map[uint]*models.User)
	m.emails = make(map[string]*models.User)
	return nil
}

func TestUserService_ValidatePassword(t *testing.T) {
	// Arrange
	service := &userService{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "exactly 8 characters",
			password: "Pass1234",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := service.ValidatePassword(tt.password)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "password must be at least 8 characters long")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_GenerateJWT(t *testing.T) {
	// Arrange
	config := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-jwt-generation",
		},
	}
	logger := logger.New()

	service := &userService{
		config: config,
		logger: logger,
	}

	userID := uint(123)

	// Act
	token, err := service.GenerateJWT(userID)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 50) // JWT tokens are typically long
}

func TestUserService_GenerateJWT_EmptySecret(t *testing.T) {
	// Arrange
	config := &config.Config{
		JWT: config.JWTConfig{
			Secret: "", // Empty secret
		},
	}
	logger := logger.New()

	service := &userService{
		config: config,
		logger: logger,
	}

	userID := uint(123)

	// Act
	token, err := service.GenerateJWT(userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrTokenGeneration, err)
	assert.Empty(t, token)
}

func TestUserService_NewUserService(t *testing.T) {
	// Arrange
	config := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}
	logger := logger.New()
	mockCache := NewMockUserCache()

	// Act
	service := NewUserService(nil, mockCache, config, logger)

	// Assert
	assert.NotNil(t, service)
	assert.Implements(t, (*UserService)(nil), service)
}

func TestUserService_RequestTypes(t *testing.T) {
	// Test that request types are properly defined
	registerReq := &RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123",
		Name:     "Test User",
	}

	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}

	updateReq := &UpdateUserRequest{
		Name: "Updated Name",
	}

	// Assert
	assert.Equal(t, "test@example.com", registerReq.Email)
	assert.Equal(t, "Password123", registerReq.Password)
	assert.Equal(t, "Test User", registerReq.Name)

	assert.Equal(t, "test@example.com", loginReq.Email)
	assert.Equal(t, "Password123", loginReq.Password)

	assert.Equal(t, "Updated Name", updateReq.Name)
}

func TestUserService_ResponseTypes(t *testing.T) {
	// Test that response types are properly defined
	userResp := &UserResponse{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	authResp := &AuthResponse{
		Message: "Success",
		Token:   "jwt-token",
		User:    *userResp,
	}

	// Assert
	assert.Equal(t, uint(1), userResp.ID)
	assert.Equal(t, "test@example.com", userResp.Email)
	assert.Equal(t, "Test User", userResp.Name)

	assert.Equal(t, "Success", authResp.Message)
	assert.Equal(t, "jwt-token", authResp.Token)
	assert.Equal(t, *userResp, authResp.User)
}

func TestUserService_ErrorTypes(t *testing.T) {
	// Test that error types are properly defined
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "user already exists", ErrUserAlreadyExists.Error())
	assert.Equal(t, "invalid credentials", ErrInvalidCredentials.Error())
	assert.Equal(t, "invalid password", ErrInvalidPassword.Error())
	assert.Equal(t, "failed to generate token", ErrTokenGeneration.Error())
} 