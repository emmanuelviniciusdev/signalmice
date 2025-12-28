# PRP: signalmice - Remote Machine Shutdown Service

**Status**: Implemented
**Version**: 1.0.0
**Created**: 2024-12-28

---

## Goal

**Feature Goal**: Create a lightweight Go service that monitors a Redis key and triggers a remote shutdown of the host Linux machine when the signal is detected.

**Deliverable**: Docker container running a Go application that:
- Periodically checks Redis for a shutdown signal key
- Logs all activities to Opensearch
- Safely shuts down the host machine (not the container)

**Success Definition**:
- Service starts and connects to Redis successfully
- When a Redis key is set, the host machine initiates shutdown
- All events are logged to Opensearch with proper structure

## User Persona

**Target User**: System administrators managing Linux servers (Debian/Ubuntu)

**Use Case**: Remote shutdown of machines that may be inaccessible via SSH or other management tools

**User Journey**:
1. Admin deploys signalmice container on target machine
2. Admin configures unique Redis key for the machine
3. When shutdown is needed, admin sets the key in Redis
4. Machine shuts down gracefully within the check interval
5. Shutdown event is logged to Opensearch for audit

**Pain Points Addressed**:
- Inability to remotely shutdown machines without SSH access
- Need for audit trail of shutdown events
- Centralized control of multiple machines via Redis

## Why

- **Remote control**: Enable shutdown of machines without direct access
- **Audit compliance**: All shutdown events logged to centralized Opensearch
- **Simplicity**: Single Redis key check, minimal configuration
- **Security**: Unique keys per machine, container isolation

## What

### User-Visible Behavior

1. Container starts and logs connection status
2. Periodic check of Redis key (configurable interval, default 60s)
3. When key exists:
   - Key is deleted from Redis
   - Shutdown event is logged
   - Host machine shutdown is initiated
4. Graceful handling of SIGTERM/SIGINT signals

### Technical Requirements

- Go 1.22+
- Redis client with connection pooling
- Opensearch client with async logging
- Multiple shutdown methods (nsenter, sysrq, direct)
- Multi-stage Docker build for minimal image size

### Success Criteria

- [x] Service connects to Redis on startup
- [x] Service logs to Opensearch (or stdout if unavailable)
- [x] Redis key detection triggers shutdown
- [x] Key is deleted after detection
- [x] Host machine (not container) shuts down
- [x] Docker image < 20MB
- [x] Graceful signal handling

## All Needed Context

### Context Completeness Check

_This PRP contains all necessary information to implement or maintain the signalmice service._

### Documentation & References

```yaml
- url: https://pkg.go.dev/github.com/go-redis/redis/v8
  why: Redis client for Go - connection, get, delete operations
  critical: Use context for all operations, handle redis.Nil for missing keys

- url: https://pkg.go.dev/github.com/opensearch-project/opensearch-go/v2
  why: Opensearch client for structured logging
  critical: TLS config for self-signed certs, async indexing

- url: https://man7.org/linux/man-pages/man8/nsenter.8.html
  why: Enter host namespace from container for shutdown
  critical: Requires --privileged and --pid=host on container

- url: https://www.kernel.org/doc/html/latest/admin-guide/sysrq.html
  why: Alternative shutdown via /proc/sysrq-trigger
  critical: 's' for sync, 'u' for remount-ro, 'o' for poweroff
```

### Current Codebase Tree

```bash
signalmice/
├── cmd/
│   └── signalmice/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment variable configuration
│   ├── logger/
│   │   └── logger.go            # Opensearch + stdout logging
│   ├── redis/
│   │   └── client.go            # Redis client wrapper
│   └── shutdown/
│       └── shutdown.go          # Host shutdown logic
├── Dockerfile                   # Multi-stage build
├── docker-compose.yml           # Production deployment
├── docker-compose.dev.yml       # Development with local Redis/Opensearch
├── go.mod                       # Go module definition
├── .env.example                 # Environment variable template
├── .dockerignore
├── .gitignore
└── README.md
```

### Known Gotchas & Library Quirks

```go
// CRITICAL: Redis key check returns redis.Nil when key doesn't exist
// Must check for redis.Nil specifically, not generic error
if err == redis.Nil {
    // Key doesn't exist - normal case
    return false, nil
}

// CRITICAL: Container needs special privileges for host shutdown
// Docker run flags: --privileged --pid=host -v /proc:/host/proc:ro

// CRITICAL: Opensearch client may fail silently
// Always log to stdout AND Opensearch, never only Opensearch

// CRITICAL: nsenter requires PID 1 access (host's init)
// nsenter --target 1 --mount --uts --ipc --net --pid -- poweroff
```

## Implementation Blueprint

### Data Models and Structure

```go
// Config - internal/config/config.go
type Config struct {
    RedisHost          string
    RedisPort          string
    RedisPassword      string
    RedisDB            int
    OpensearchURL      string
    OpensearchUsername string
    OpensearchPassword string
    OpensearchIndex    string
    RedisKey           string
    CheckInterval      time.Duration
    HostProcPath       string
}

// LogEntry - internal/logger/logger.go
type LogEntry struct {
    Timestamp string `json:"@timestamp"`
    Level     Level  `json:"level"`
    Message   string `json:"message"`
    Hostname  string `json:"hostname"`
    Service   string `json:"service"`
    RedisKey  string `json:"redis_key,omitempty"`
    Extra     any    `json:"extra,omitempty"`
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: CREATE internal/config/config.go
  - IMPLEMENT: Config struct with all environment variables
  - FOLLOW pattern: Standard Go env parsing with defaults
  - NAMING: CamelCase for exported fields
  - VALIDATION: Load() function returns populated Config

Task 2: CREATE internal/redis/client.go
  - IMPLEMENT: Client wrapper with CheckAndDeleteKey method
  - FOLLOW pattern: go-redis/v8 best practices
  - NAMING: Client struct, NewClient constructor
  - DEPENDENCIES: Uses config.Config for connection

Task 3: CREATE internal/logger/logger.go
  - IMPLEMENT: Logger with Opensearch + stdout dual output
  - FOLLOW pattern: opensearch-go/v2 indexing
  - NAMING: Logger struct, Info/Warn/Error/Debug methods
  - DEPENDENCIES: Uses config.Config for Opensearch connection

Task 4: CREATE internal/shutdown/shutdown.go
  - IMPLEMENT: Manager with NeutralizeStuartLittle method (multi-method shutdown)
  - FOLLOW pattern: Try nsenter → sysrq → direct command
  - NAMING: Manager struct, NeutralizeStuartLittle main method
  - DEPENDENCIES: Uses logger for audit logging

Task 5: CREATE cmd/signalmice/main.go
  - IMPLEMENT: Main loop with ticker for periodic checks
  - FOLLOW pattern: Signal handling with context cancellation
  - NAMING: checkAndShutdown helper function
  - DEPENDENCIES: All internal packages

Task 6: CREATE Dockerfile
  - IMPLEMENT: Multi-stage build (golang:1.22-alpine → alpine:3.19)
  - FOLLOW pattern: CGO_ENABLED=0 for static binary
  - OPTIMIZE: -ldflags="-w -s" for smaller binary
  - INCLUDE: util-linux for nsenter command

Task 7: CREATE docker-compose.yml
  - IMPLEMENT: Production deployment configuration
  - CRITICAL: privileged: true, pid: host, /proc mount
  - FOLLOW pattern: Environment variable passthrough
```

### Implementation Patterns & Key Details

```go
// Redis key check pattern
func (c *Client) CheckAndDeleteKey(ctx context.Context) (bool, error) {
    result, err := c.client.Get(ctx, c.key).Result()
    if err == redis.Nil {
        return false, nil  // Key doesn't exist - normal
    }
    if err != nil {
        return false, fmt.Errorf("failed to get key: %w", err)
    }

    // Key exists - delete it
    if err := c.client.Del(ctx, c.key).Err(); err != nil {
        return false, fmt.Errorf("failed to delete key: %w", err)
    }
    return true, nil
}

// Shutdown method chain pattern
func (m *Manager) NeutralizeStuartLittle(ctx context.Context) error {
    methods := []struct {
        name string
        fn   func(context.Context) error
    }{
        {"nsenter", m.shutdownViaNsenter},
        {"sysrq-trigger", m.shutdownViaSysrq},
        {"direct-command", m.shutdownViaDirect},
    }

    for _, method := range methods {
        if err := method.fn(ctx); err != nil {
            continue  // Try next method
        }
        return nil  // Success
    }
    return fmt.Errorf("all shutdown methods failed")
}

// Main loop pattern
ticker := time.NewTicker(cfg.CheckInterval)
for {
    select {
    case <-ticker.C:
        checkAndShutdown(ctx, redisClient, shutdownManager, appLogger)
    case <-sigChan:
        cancel()
        return
    }
}
```

### Integration Points

```yaml
DOCKER:
  - volume: "/proc:/host/proc:ro"
  - flag: "privileged: true"
  - flag: "pid: host"
  - flag: "network_mode: host" (optional, for localhost Redis)

REDIS:
  - key format: "signalmice:{uuid}"
  - default: "signalmice:00000000-0000-0000-0000-000000000000"
  - trigger: SET key to any value

OPENSEARCH:
  - index: "signalmice-logs" (configurable)
  - document: LogEntry struct as JSON
```

## Validation Loop

### Level 1: Syntax & Style (Go)

```bash
# Run after each file creation
golangci-lint run ./...           # Comprehensive linting
go fmt ./...                       # Format code
go vet ./...                       # Static analysis

# Expected: Zero errors
```

### Level 2: Unit Tests

```bash
# Test each package
go test ./internal/config/... -v
go test ./internal/redis/... -v
go test ./internal/logger/... -v
go test ./internal/shutdown/... -v

# Full test suite with coverage
go test ./... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Expected: All tests pass, >80% coverage
```

### Level 3: Integration Testing

```bash
# Build Docker image
docker build -t signalmice:test .

# Start dev environment
docker-compose -f docker-compose.dev.yml up -d

# Check service logs
docker logs signalmice

# Set shutdown key (DO NOT RUN ON PRODUCTION!)
# redis-cli SET "signalmice:00000000-0000-0000-0000-000000000000" "test"

# Expected: Service starts, connects to Redis/Opensearch
```

### Level 4: Production Validation

```bash
# Verify Docker image size
docker images signalmice:latest --format "{{.Size}}"
# Expected: < 20MB

# Verify required capabilities
docker inspect signalmice | jq '.[0].HostConfig.Privileged'
# Expected: true

# Verify /proc mount
docker inspect signalmice | jq '.[0].Mounts[] | select(.Destination=="/host/proc")'
# Expected: Mount exists
```

## Final Validation Checklist

### Technical Validation

- [x] All linting passes: `golangci-lint run ./...`
- [x] All tests pass: `go test ./... -v`
- [x] Docker build succeeds: `docker build -t signalmice .`
- [x] Image size < 20MB

### Feature Validation

- [x] Connects to Redis on startup
- [x] Logs startup to Opensearch
- [x] Detects and deletes Redis key
- [x] Initiates host shutdown (tested in isolated VM)
- [x] Handles SIGTERM gracefully

### Code Quality Validation

- [x] Follows Go best practices (fmt, vet, lint)
- [x] Error wrapping with context
- [x] Context propagation throughout
- [x] Proper resource cleanup (defer Close())

---

## Anti-Patterns to Avoid

- ❌ Don't log sensitive data (passwords, keys)
- ❌ Don't ignore errors - wrap and return them
- ❌ Don't use sync operations in async context
- ❌ Don't hardcode Redis key - use environment variable
- ❌ Don't skip the multi-method shutdown chain
- ❌ Don't test shutdown on production machines!
