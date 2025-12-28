package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/signalmice/signalmice/internal/config"
	"github.com/signalmice/signalmice/internal/logger"
	"github.com/signalmice/signalmice/internal/redis"
	"github.com/signalmice/signalmice/internal/shutdown"
)

const (
	appName    = "signalmice"
	appVersion = "1.0.0"
)

func main() {
	log.Printf("%s v%s starting...", appName, appVersion)

	// Load configuration
	cfg := config.Load()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.InfoWithExtra(ctx, fmt.Sprintf("%s starting", appName), map[string]any{
		"version":        appVersion,
		"check_interval": cfg.CheckInterval.String(),
		"redis_key":      cfg.RedisKey,
	})

	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg)
	if err != nil {
		appLogger.ErrorWithExtra(ctx, "Failed to connect to Redis", map[string]string{"error": err.Error()})
		os.Exit(1)
	}
	defer redisClient.Close()

	appLogger.Info(ctx, "Connected to Redis successfully")

	// Initialize shutdown manager
	shutdownManager := shutdown.NewManager(cfg, appLogger)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the main monitoring loop
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	appLogger.Info(ctx, fmt.Sprintf("Starting Redis key monitoring (key: %s, interval: %s)", cfg.RedisKey, cfg.CheckInterval))

	// Run the initial check immediately
	checkAndShutdown(ctx, redisClient, shutdownManager, appLogger)

	for {
		select {
		case <-ticker.C:
			checkAndShutdown(ctx, redisClient, shutdownManager, appLogger)

		case sig := <-sigChan:
			appLogger.InfoWithExtra(ctx, "Received shutdown signal", map[string]string{"signal": sig.String()})
			cancel()
			appLogger.Info(ctx, "Graceful shutdown complete")
			return
		}
	}
}

// checkAndShutdown checks for the signal key and initiates shutdown if found
func checkAndShutdown(ctx context.Context, redisClient *redis.Client, shutdownManager *shutdown.Manager, appLogger *logger.Logger) {
	found, err := redisClient.CheckAndDeleteKey(ctx)
	if err != nil {
		appLogger.ErrorWithExtra(ctx, "Error checking Redis key", map[string]string{"error": err.Error()})
		return
	}

	if !found {
		appLogger.Debug(ctx, "Redis key not found, continuing to monitor...")
		return
	}

	// Signal key was found and deleted
	appLogger.InfoWithExtra(ctx, "Shutdown signal received! Key found and deleted.", map[string]string{"key": redisClient.GetKey()})

	// Initiate host shutdown
	if err := shutdownManager.NeutralizeStuartLittle(ctx); err != nil {
		appLogger.ErrorWithExtra(ctx, "Failed to initiate host shutdown", map[string]string{"error": err.Error()})
		return
	}

	appLogger.Info(ctx, "Host shutdown initiated successfully")
}
