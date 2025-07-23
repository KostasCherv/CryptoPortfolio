package cache

import (
	"context"
	"testing"
	"time"

	"simple_api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCache_Operations(t *testing.T) {
	// Create a mock cache service for testing
	mockCache := NewMockUserCache()
	
	ctx := context.Background()
	
	// Test user
	user := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Test SetUserByID and GetUserByID
	err := mockCache.SetUserByID(ctx, user)
	require.NoError(t, err)
	
	retrievedUser, err := mockCache.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.Name, retrievedUser.Name)
	
	// Test cache miss
	_, err = mockCache.GetUserByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
	
	// Test SetUserByEmail and GetUserByEmail
	err = mockCache.SetUserByEmail(ctx, user)
	require.NoError(t, err)
	
	retrievedUser, err = mockCache.GetUserByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	
	// Test InvalidateUser
	err = mockCache.InvalidateUser(ctx, user.ID, user.Email)
	require.NoError(t, err)
	
	// Should be cache miss after invalidation
	_, err = mockCache.GetUserByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
	
	_, err = mockCache.GetUserByEmail(ctx, user.Email)
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
}

// MockUserCache for testing
type MockUserCache struct {
	users  map[uint]*models.User
	emails map[string]*models.User
}

func NewMockUserCache() *MockUserCache {
	return &MockUserCache{
		users:  make(map[uint]*models.User),
		emails: make(map[string]*models.User),
	}
}

func (m *MockUserCache) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, ErrCacheMiss
}

func (m *MockUserCache) SetUserByID(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserCache) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, exists := m.emails[email]; exists {
		return user, nil
	}
	return nil, ErrCacheMiss
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