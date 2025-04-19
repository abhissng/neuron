package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned when a key is not found in Redis.
// This provides a distinct error type compared to the underlying redis.Nil.
var ErrNotFound = errors.New("rediswrapper: key not found")

// Config holds the configuration for the Redis wrapper.
type Config struct {
	Addr     string // e.g., "localhost:6379"
	Password string // Leave empty if no password
	DB       int    // Default is 0
	// DefaultTTL is used for cache-specific methods if no TTL is provided.
	// Set to 0 to disable default TTL (meaning keys persist unless TTL specified).
	DefaultTTL time.Duration
}

func NewConfig() *Config {
	return &Config{
		DefaultTTL: 0,
	}
}

// RedisManager provides a simplified interface over the go-redis client.
type RedisManager struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedisWrapper creates and initializes a new RedisManager.
// It pings the Redis server to ensure connectivity.
func NewRedisWrapper(cfg Config) (*RedisManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close() // Close the client if ping fails
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.Addr, err)
	}

	return &RedisManager{
		client:     rdb,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// Client returns the underlying go-redis client instance for advanced use cases.
func (rw *RedisManager) Client() *redis.Client {
	return rw.client
}

// Close closes the underlying Redis client connection.
func (rw *RedisManager) Close() error {
	if rw.client != nil {
		return rw.client.Close()
	}
	return nil
}

// === Basic Key-Value Operations ===

// Set stores a string value for a key.
// TTL of 0 means the key persists indefinitely.
func (rw *RedisManager) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := rw.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Get retrieves a string value for a key.
// Returns ErrNotFound if the key does not exist.
func (rw *RedisManager) Get(ctx context.Context, key string) (string, error) {
	val, err := rw.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Delete removes one or more keys.
// Returns the number of keys removed and nil error on success.
func (rw *RedisManager) Delete(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	deletedCount, err := rw.client.Del(ctx, keys...).Result()
	if err != nil {
		// Don't wrap if it's redis.Nil, though Del usually doesn't return Nil error
		if errors.Is(err, redis.Nil) {
			return 0, nil // Or return 0, ErrNotFound depending on desired semantics
		}
		return 0, fmt.Errorf("failed to delete keys: %w", err)
	}
	return deletedCount, nil
}

// Exists checks if one or more keys exist.
// Returns the count of existing keys.
func (rw *RedisManager) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	count, err := rw.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence for keys: %w", err)
	}
	return count, nil
}

// === JSON Operations ===

// SetJSON marshals the given value into JSON and stores it.
// TTL of 0 means the key persists indefinitely.
func (rw *RedisManager) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	err = rw.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set JSON for key %s: %w", key, err)
	}
	return nil
}

// GetJSON retrieves a value and unmarshals it from JSON into the destination pointer.
// `dest` must be a pointer to the target type.
// Returns ErrNotFound if the key does not exist.
func (rw *RedisManager) GetJSON(ctx context.Context, key string, dest interface{}) error {
	jsonData, err := rw.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get JSON for key %s: %w", key, err)
	}

	err = json.Unmarshal(jsonData, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}
	return nil
}

// === Cache Specific Operations (using DefaultTTL) ===

// CacheSet stores a value as JSON using the default TTL configured for the wrapper.
// If DefaultTTL is 0, the key will persist.
func (rw *RedisManager) CacheSet(ctx context.Context, key string, value interface{}) error {
	return rw.SetJSON(ctx, key, value, rw.defaultTTL)
}

// CacheGet retrieves a JSON value into the destination pointer.
// Returns ErrNotFound if the key does not exist.
func (rw *RedisManager) CacheGet(ctx context.Context, key string, dest interface{}) error {
	return rw.GetJSON(ctx, key, dest)
}

// === Atomic Operations ===

// Increment atomically increments the integer value of a key by one.
// Returns the value after the increment. If the key does not exist, it's set to 0 before performing the operation.
func (rw *RedisManager) Increment(ctx context.Context, key string) (int64, error) {
	val, err := rw.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Decrement atomically decrements the integer value of a key by one.
// Returns the value after the decrement. If the key does not exist, it's set to 0 before performing the operation.
func (rw *RedisManager) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := rw.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return val, nil
}
