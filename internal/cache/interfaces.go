package cache

import (
	"context"
	"fmt"
	"time"

	"cryptoportfolio/internal/models"
)

// Common cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

// CacheProvider defines the interface for cache operations
type CacheProvider interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// UserCacheProvider defines user-specific cache operations
type UserCacheProvider interface {
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	SetUserByID(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SetUserByEmail(ctx context.Context, user *models.User) error
	InvalidateUser(ctx context.Context, userID uint, email string) error
	InvalidateAllUsers(ctx context.Context) error
} 