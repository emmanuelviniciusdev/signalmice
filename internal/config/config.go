package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Opensearch configuration
	OpensearchURL           string
	OpensearchUsername      string
	OpensearchPassword      string
	OpensearchIndex         string
	OpensearchUseDailyIndex bool

	// Application configuration
	RedisKey      string
	CheckInterval time.Duration

	// Host configuration
	HostProcPath string // Path to host's /proc for shutdown
}

// DefaultRedisKey is the default key to check in Redis
const DefaultRedisKey = "signalmice:00000000-0000-0000-0000-000000000000"

// Load loads configuration from environment variables
func Load() *Config {
	checkInterval, _ := strconv.Atoi(getEnv("SIGNALMICE_CHECK_INTERVAL", "60"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// Opensearch
		OpensearchURL:           getEnv("OPENSEARCH_URL", "http://localhost:9200"),
		OpensearchUsername:      getEnv("OPENSEARCH_USERNAME", ""),
		OpensearchPassword:      getEnv("OPENSEARCH_PASSWORD", ""),
		OpensearchIndex:         getEnv("OPENSEARCH_INDEX", "signalmice-logs"),
		OpensearchUseDailyIndex: getEnvBool("OPENSEARCH_USE_DAILY_INDEX", true),

		// Application
		RedisKey:      getEnv("SIGNALMICE_KEY", DefaultRedisKey),
		CheckInterval: time.Duration(checkInterval) * time.Second,

		// Host
		HostProcPath: getEnv("HOST_PROC_PATH", "/host/proc"),
	}
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvBool returns the boolean value of an environment variable or a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// RedisAddr returns the Redis address in host:port format
func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}
