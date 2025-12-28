package logger

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/signalmice/signalmice/internal/config"
)

// Level represents the log level
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelDebug Level = "DEBUG"
)

// LogEntry represents a log entry to be sent to Opensearch
type LogEntry struct {
	Timestamp string `json:"@timestamp"`
	Level     Level  `json:"level"`
	Message   string `json:"message"`
	Hostname  string `json:"hostname"`
	Service   string `json:"service"`
	RedisKey  string `json:"redis_key,omitempty"`
	Extra     any    `json:"extra,omitempty"`
}

// Logger handles logging to both stdout and Opensearch
type Logger struct {
	client   *opensearch.Client
	index    string
	hostname string
	redisKey string
}

// NewLogger creates a new logger that writes to Opensearch
func NewLogger(cfg *config.Config) (*Logger, error) {
	hostname, _ := os.Hostname()

	// Create Opensearch client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Allow self-signed certificates
		},
	}

	osConfig := opensearch.Config{
		Addresses: []string{cfg.OpensearchURL},
		Transport: transport,
	}

	// Add authentication if provided
	if cfg.OpensearchUsername != "" {
		osConfig.Username = cfg.OpensearchUsername
		osConfig.Password = cfg.OpensearchPassword
	}

	client, err := opensearch.NewClient(osConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opensearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		log.Printf("[WARN] Could not connect to Opensearch: %v. Logging will continue to stdout only.", err)
		return &Logger{
			client:   nil,
			index:    cfg.OpensearchIndex,
			hostname: hostname,
			redisKey: cfg.RedisKey,
		}, nil
	}
	defer res.Body.Close()

	return &Logger{
		client:   client,
		index:    cfg.OpensearchIndex,
		hostname: hostname,
		redisKey: cfg.RedisKey,
	}, nil
}

// log sends a log entry to Opensearch and prints to stdout
func (l *Logger) log(ctx context.Context, level Level, message string, extra any) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Hostname:  l.hostname,
		Service:   "signalmice",
		RedisKey:  l.redisKey,
		Extra:     extra,
	}

	// Always log to stdout
	log.Printf("[%s] %s", level, message)

	// Send to Opensearch if client is available
	if l.client != nil {
		go l.sendToOpensearch(ctx, entry)
	}
}

// sendToOpensearch sends a log entry to Opensearch
func (l *Logger) sendToOpensearch(ctx context.Context, entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal log entry: %v", err)
		return
	}

	res, err := l.client.Index(
		l.index,
		bytes.NewReader(data),
		l.client.Index.WithContext(ctx),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to send log to Opensearch: %v", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[ERROR] Opensearch returned error: %s", res.Status())
	}
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string) {
	l.log(ctx, LevelInfo, message, nil)
}

// InfoWithExtra logs an info message with extra data
func (l *Logger) InfoWithExtra(ctx context.Context, message string, extra any) {
	l.log(ctx, LevelInfo, message, extra)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string) {
	l.log(ctx, LevelWarn, message, nil)
}

// WarnWithExtra logs a warning message with extra data
func (l *Logger) WarnWithExtra(ctx context.Context, message string, extra any) {
	l.log(ctx, LevelWarn, message, extra)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string) {
	l.log(ctx, LevelError, message, nil)
}

// ErrorWithExtra logs an error message with extra data
func (l *Logger) ErrorWithExtra(ctx context.Context, message string, extra any) {
	l.log(ctx, LevelError, message, extra)
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string) {
	l.log(ctx, LevelDebug, message, nil)
}

// DebugWithExtra logs a debug message with extra data
func (l *Logger) DebugWithExtra(ctx context.Context, message string, extra any) {
	l.log(ctx, LevelDebug, message, extra)
}
