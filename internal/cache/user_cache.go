package cache

import (
	"context"
	"fmt"
	"time"

	"cryptoportfolio/internal/models"
)

// UserCache provides user-specific caching operations
type UserCache struct {
	cacheService CacheProvider
}

// NewUserCache creates a new user cache
func NewUserCache(cacheService CacheProvider) *UserCache {
	return &UserCache{
		cacheService: cacheService,
	}
}

// GetUserByID retrieves a user from cache by ID
func (uc *UserCache) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	var user models.User
	
	err := uc.cacheService.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// SetUserByID stores a user in cache by ID
func (uc *UserCache) SetUserByID(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	return uc.cacheService.Set(ctx, key, user, 30*time.Minute)
}

// GetUserByEmail retrieves a user from cache by email
func (uc *UserCache) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	key := fmt.Sprintf("user:email:%s", email)
	var user models.User
	
	err := uc.cacheService.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// SetUserByEmail stores a user in cache by email
func (uc *UserCache) SetUserByEmail(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:email:%s", user.Email)
	return uc.cacheService.Set(ctx, key, user, 30*time.Minute)
}

// InvalidateUser invalidates all cache entries for a user
func (uc *UserCache) InvalidateUser(ctx context.Context, userID uint, email string) error {
	// Delete user by ID
	keyByID := fmt.Sprintf("user:%d", userID)
	if err := uc.cacheService.Delete(ctx, keyByID); err != nil {
		return err
	}
	
	// Delete user by email
	keyByEmail := fmt.Sprintf("user:email:%s", email)
	if err := uc.cacheService.Delete(ctx, keyByEmail); err != nil {
		return err
	}
	
	return nil
}

// InvalidateAllUsers invalidates all user-related cache entries
func (uc *UserCache) InvalidateAllUsers(ctx context.Context) error {
	pattern := "user:*"
	return uc.cacheService.DeletePattern(ctx, pattern)
} 