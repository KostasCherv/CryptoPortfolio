package cache

import (
	"context"
	"encoding/json"
	"time"

	"simple_api/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int, logger *logger.Logger) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{
		client: client,
		logger: logger,
	}
}

// Set stores a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.Error("Failed to marshal value for cache", "error", err, "key", key)
		return err
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to set cache value", "error", err, "key", key)
		return err
	}

	r.logger.Debug("Cache set successfully", "key", key, "expiration", expiration)
	return nil
}

// Get retrieves a value by key and unmarshals it into the provided interface
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			r.logger.Debug("Cache miss", "key", key)
			return ErrCacheMiss
		}
		r.logger.Error("Failed to get cache value", "error", err, "key", key)
		return err
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		r.logger.Error("Failed to unmarshal cached value", "error", err, "key", key)
		return err
	}

	r.logger.Debug("Cache hit", "key", key)
	return nil
}

// Delete removes a key from cache
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to delete cache key", "error", err, "key", key)
		return err
	}

	r.logger.Debug("Cache key deleted", "key", key)
	return nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("Failed to get keys for pattern", "error", err, "pattern", pattern)
		return err
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			r.logger.Error("Failed to delete keys by pattern", "error", err, "pattern", pattern)
			return err
		}
		r.logger.Debug("Cache keys deleted by pattern", "pattern", pattern, "count", len(keys))
	}

	return nil
}

// Ping tests the Redis connection
func (r *RedisClient) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		r.logger.Error("Redis ping failed", "error", err)
		return err
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
} 