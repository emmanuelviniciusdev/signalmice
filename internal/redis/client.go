package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/signalmice/signalmice/internal/config"
)

// Client wraps the Redis client with application-specific methods
type Client struct {
	client *redis.Client
	key    string
}

// NewClient creates a new Redis client
func NewClient(cfg *config.Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
		key:    cfg.RedisKey,
	}, nil
}

// CheckAndDeleteKey checks if the signal key exists and deletes it if found
// Returns true if the key existed and was deleted, false otherwise
func (c *Client) CheckAndDeleteKey(ctx context.Context) (bool, error) {
	// Use GET to check if key exists
	result, err := c.client.Get(ctx, c.key).Result()
	if err == redis.Nil {
		// Key does not exist
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get key: %w", err)
	}

	// Key exists, delete it
	if err := c.client.Del(ctx, c.key).Err(); err != nil {
		return false, fmt.Errorf("failed to delete key: %w", err)
	}

	// Log the value that was found (for debugging purposes)
	_ = result

	return true, nil
}

// GetKey returns the key being monitored
func (c *Client) GetKey() string {
	return c.key
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	return c.client.Close()
}
