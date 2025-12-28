package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	envVars := []string{
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"OPENSEARCH_URL", "OPENSEARCH_USERNAME", "OPENSEARCH_PASSWORD", "OPENSEARCH_INDEX",
		"OPENSEARCH_USE_DAILY_INDEX",
		"SIGNALMICE_KEY", "SIGNALMICE_CHECK_INTERVAL", "HOST_PROC_PATH",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	// Test Redis defaults
	if cfg.RedisHost != "localhost" {
		t.Errorf("expected RedisHost 'localhost', got '%s'", cfg.RedisHost)
	}
	if cfg.RedisPort != "6379" {
		t.Errorf("expected RedisPort '6379', got '%s'", cfg.RedisPort)
	}
	if cfg.RedisPassword != "" {
		t.Errorf("expected empty RedisPassword, got '%s'", cfg.RedisPassword)
	}
	if cfg.RedisDB != 0 {
		t.Errorf("expected RedisDB 0, got %d", cfg.RedisDB)
	}

	// Test Opensearch defaults
	if cfg.OpensearchURL != "http://localhost:9200" {
		t.Errorf("expected OpensearchURL 'http://localhost:9200', got '%s'", cfg.OpensearchURL)
	}
	if cfg.OpensearchUsername != "" {
		t.Errorf("expected empty OpensearchUsername, got '%s'", cfg.OpensearchUsername)
	}
	if cfg.OpensearchPassword != "" {
		t.Errorf("expected empty OpensearchPassword, got '%s'", cfg.OpensearchPassword)
	}
	if cfg.OpensearchIndex != "signalmice-logs" {
		t.Errorf("expected OpensearchIndex 'signalmice-logs', got '%s'", cfg.OpensearchIndex)
	}
	if !cfg.OpensearchUseDailyIndex {
		t.Errorf("expected OpensearchUseDailyIndex true by default, got false")
	}

	// Test Application defaults
	if cfg.RedisKey != DefaultRedisKey {
		t.Errorf("expected RedisKey '%s', got '%s'", DefaultRedisKey, cfg.RedisKey)
	}
	if cfg.CheckInterval != 60*time.Second {
		t.Errorf("expected CheckInterval 60s, got %v", cfg.CheckInterval)
	}
	if cfg.HostProcPath != "/host/proc" {
		t.Errorf("expected HostProcPath '/host/proc', got '%s'", cfg.HostProcPath)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	// Set custom env vars
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "secret123")
	os.Setenv("REDIS_DB", "5")
	os.Setenv("OPENSEARCH_URL", "https://opensearch.example.com:9200")
	os.Setenv("OPENSEARCH_USERNAME", "admin")
	os.Setenv("OPENSEARCH_PASSWORD", "adminpass")
	os.Setenv("OPENSEARCH_INDEX", "custom-logs")
	os.Setenv("SIGNALMICE_KEY", "signalmice:custom-key")
	os.Setenv("SIGNALMICE_CHECK_INTERVAL", "30")
	os.Setenv("HOST_PROC_PATH", "/custom/proc")

	defer func() {
		// Cleanup
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("OPENSEARCH_URL")
		os.Unsetenv("OPENSEARCH_USERNAME")
		os.Unsetenv("OPENSEARCH_PASSWORD")
		os.Unsetenv("OPENSEARCH_INDEX")
		os.Unsetenv("SIGNALMICE_KEY")
		os.Unsetenv("SIGNALMICE_CHECK_INTERVAL")
		os.Unsetenv("HOST_PROC_PATH")
	}()

	cfg := Load()

	if cfg.RedisHost != "redis.example.com" {
		t.Errorf("expected RedisHost 'redis.example.com', got '%s'", cfg.RedisHost)
	}
	if cfg.RedisPort != "6380" {
		t.Errorf("expected RedisPort '6380', got '%s'", cfg.RedisPort)
	}
	if cfg.RedisPassword != "secret123" {
		t.Errorf("expected RedisPassword 'secret123', got '%s'", cfg.RedisPassword)
	}
	if cfg.RedisDB != 5 {
		t.Errorf("expected RedisDB 5, got %d", cfg.RedisDB)
	}
	if cfg.OpensearchURL != "https://opensearch.example.com:9200" {
		t.Errorf("expected OpensearchURL 'https://opensearch.example.com:9200', got '%s'", cfg.OpensearchURL)
	}
	if cfg.OpensearchUsername != "admin" {
		t.Errorf("expected OpensearchUsername 'admin', got '%s'", cfg.OpensearchUsername)
	}
	if cfg.OpensearchPassword != "adminpass" {
		t.Errorf("expected OpensearchPassword 'adminpass', got '%s'", cfg.OpensearchPassword)
	}
	if cfg.OpensearchIndex != "custom-logs" {
		t.Errorf("expected OpensearchIndex 'custom-logs', got '%s'", cfg.OpensearchIndex)
	}
	if cfg.RedisKey != "signalmice:custom-key" {
		t.Errorf("expected RedisKey 'signalmice:custom-key', got '%s'", cfg.RedisKey)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected CheckInterval 30s, got %v", cfg.CheckInterval)
	}
	if cfg.HostProcPath != "/custom/proc" {
		t.Errorf("expected HostProcPath '/custom/proc', got '%s'", cfg.HostProcPath)
	}
}

func TestLoad_InvalidInterval(t *testing.T) {
	os.Setenv("SIGNALMICE_CHECK_INTERVAL", "invalid")
	defer os.Unsetenv("SIGNALMICE_CHECK_INTERVAL")

	cfg := Load()

	// Should default to 0 when parsing fails
	if cfg.CheckInterval != 0 {
		t.Errorf("expected CheckInterval 0 for invalid input, got %v", cfg.CheckInterval)
	}
}

func TestRedisAddr(t *testing.T) {
	cfg := &Config{
		RedisHost: "redis.example.com",
		RedisPort: "6380",
	}

	expected := "redis.example.com:6380"
	if cfg.RedisAddr() != expected {
		t.Errorf("expected RedisAddr '%s', got '%s'", expected, cfg.RedisAddr())
	}
}

func TestDefaultRedisKey(t *testing.T) {
	expected := "signalmice:00000000-0000-0000-0000-000000000000"
	if DefaultRedisKey != expected {
		t.Errorf("expected DefaultRedisKey '%s', got '%s'", expected, DefaultRedisKey)
	}
}

func TestLoad_OpensearchUseDailyIndex_Disabled(t *testing.T) {
	os.Setenv("OPENSEARCH_USE_DAILY_INDEX", "false")
	defer os.Unsetenv("OPENSEARCH_USE_DAILY_INDEX")

	cfg := Load()

	if cfg.OpensearchUseDailyIndex {
		t.Errorf("expected OpensearchUseDailyIndex false when set to 'false', got true")
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"true string", "true", false, true},
		{"1 string", "1", false, true},
		{"yes string", "yes", false, true},
		{"false string", "false", true, false},
		{"0 string", "0", true, false},
		{"no string", "no", true, false},
		{"empty uses default true", "", true, true},
		{"empty uses default false", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL", tt.envValue)
				defer os.Unsetenv("TEST_BOOL")
			} else {
				os.Unsetenv("TEST_BOOL")
			}

			result := getEnvBool("TEST_BOOL", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(%q, %v) = %v, expected %v", tt.envValue, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
