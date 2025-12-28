package logger

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/signalmice/signalmice/internal/config"
)

func createTestConfig() *config.Config {
	return &config.Config{
		OpensearchURL:      "http://localhost:9200",
		OpensearchUsername: "",
		OpensearchPassword: "",
		OpensearchIndex:    "test-logs",
		RedisKey:           "signalmice:test-key",
	}
}

func TestLogEntry_Structure(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     LevelInfo,
		Message:   "Test message",
		Hostname:  "test-host",
		Service:   "signalmice",
		RedisKey:  "signalmice:test",
		Extra:     map[string]string{"key": "value"},
	}

	if entry.Level != LevelInfo {
		t.Errorf("expected level INFO, got %s", entry.Level)
	}
	if entry.Service != "signalmice" {
		t.Errorf("expected service 'signalmice', got '%s'", entry.Service)
	}
}

func TestLevelConstants(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelDebug, "DEBUG"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.level)
		}
	}
}

func TestNewLogger_WithoutOpensearch(t *testing.T) {
	cfg := &config.Config{
		OpensearchURL:   "http://non-existent:9200",
		OpensearchIndex: "test-logs",
		RedisKey:        "test-key",
	}

	// Should not fail even if Opensearch is not available
	// Logger should work with stdout only
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Errorf("NewLogger should not fail even without Opensearch: %v", err)
	}

	if logger == nil {
		t.Error("logger should not be nil")
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil, // No Opensearch client
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.Info(ctx, "Test info message")

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected output to contain '[INFO]', got: %s", output)
	}
	if !strings.Contains(output, "Test info message") {
		t.Errorf("expected output to contain 'Test info message', got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.Warn(ctx, "Test warning message")

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected output to contain '[WARN]', got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.Error(ctx, "Test error message")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("expected output to contain '[ERROR]', got: %s", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.Debug(ctx, "Test debug message")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected output to contain '[DEBUG]', got: %s", output)
	}
}

func TestLogger_InfoWithExtra(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	extra := map[string]string{"key": "value"}
	logger.InfoWithExtra(ctx, "Test message with extra", extra)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected output to contain '[INFO]', got: %s", output)
	}
	if !strings.Contains(output, "Test message with extra") {
		t.Errorf("expected output to contain message, got: %s", output)
	}
}

func TestLogger_WarnWithExtra(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.WarnWithExtra(ctx, "Warning with extra", map[string]int{"count": 5})

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected output to contain '[WARN]', got: %s", output)
	}
}

func TestLogger_ErrorWithExtra(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.ErrorWithExtra(ctx, "Error with extra", map[string]string{"error": "test error"})

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("expected output to contain '[ERROR]', got: %s", output)
	}
}

func TestLogger_DebugWithExtra(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := &Logger{
		client:   nil,
		index:    "test",
		hostname: "test-host",
		redisKey: "test-key",
	}

	ctx := context.Background()
	logger.DebugWithExtra(ctx, "Debug with extra", map[string]bool{"verbose": true})

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected output to contain '[DEBUG]', got: %s", output)
	}
}
