package cache

import (
	"context"
	"time"

	"simple_api/pkg/logger"
)

// CacheService provides generic caching functionality
type CacheService struct {
	redis  *RedisClient
	logger *logger.Logger
}

// NewCacheService creates a new cache service
func NewCacheService(redis *RedisClient, logger *logger.Logger) *CacheService {
	return &CacheService{
		redis:  redis,
		logger: logger,
	}
}

// Set stores a key-value pair with optional expiration
func (cs *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cs.redis.Set(ctx, key, value, expiration)
}

// Get retrieves a value by key and unmarshals it into the provided interface
func (cs *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	return cs.redis.Get(ctx, key, dest)
}

// Delete removes a key from cache
func (cs *CacheService) Delete(ctx context.Context, key string) error {
	return cs.redis.Delete(ctx, key)
}

// DeletePattern removes all keys matching a pattern
func (cs *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	return cs.redis.DeletePattern(ctx, pattern)
} 