package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/signalmice/signalmice/internal/config"
)

// mockRedisServer creates a test configuration
// In a real scenario, you would use miniredis or a test container
func createTestConfig() *config.Config {
	return &config.Config{
		RedisHost:     "localhost",
		RedisPort:     "6379",
		RedisPassword: "",
		RedisDB:       15, // Use DB 15 for testing
		RedisKey:      "signalmice:test-key",
	}
}

func TestClient_GetKey(t *testing.T) {
	testKey := "signalmice:my-test-key"
	client := &Client{
		key: testKey,
	}

	if client.GetKey() != testKey {
		t.Errorf("expected key '%s', got '%s'", testKey, client.GetKey())
	}
}

func TestNewClient_ConnectionError(t *testing.T) {
	cfg := &config.Config{
		RedisHost:     "non-existent-host",
		RedisPort:     "6379",
		RedisPassword: "",
		RedisDB:       0,
		RedisKey:      "test-key",
	}

	// This should fail to connect
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("expected error when connecting to non-existent host")
	}
}

// TestClient_CheckAndDeleteKey_Integration tests with a real Redis if available
// Skip this test if Redis is not available
func TestClient_CheckAndDeleteKey_Integration(t *testing.T) {
	cfg := createTestConfig()

	// Try to connect - skip if Redis not available
	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Skipping integration test - Redis not available: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testKey := cfg.RedisKey

	// Test 1: Key does not exist
	found, err := client.CheckAndDeleteKey(ctx)
	if err != nil {
		t.Errorf("unexpected error checking non-existent key: %v", err)
	}
	if found {
		t.Error("expected key not found, but it was found")
	}

	// Test 2: Set key and verify it's found and deleted
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	// Set the key
	if err := redisClient.Set(ctx, testKey, "shutdown", 0).Err(); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Check and delete
	found, err = client.CheckAndDeleteKey(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected key to be found, but it was not")
	}

	// Verify key was deleted
	_, err = redisClient.Get(ctx, testKey).Result()
	if err != redis.Nil {
		t.Error("expected key to be deleted, but it still exists")
	}
}

func TestClient_Close(t *testing.T) {
	cfg := createTestConfig()

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
	}

	// Close should not return an error
	if err := client.Close(); err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}
