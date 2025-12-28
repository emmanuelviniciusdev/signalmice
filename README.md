![signalmice logo](logo.png)

# signalmice

Remote machine shutdown service for Linux (Debian/Ubuntu) systems via Redis signal.

## Overview

signalmice is a lightweight Go application that monitors a Redis key and triggers a host machine shutdown when the key is detected. It runs inside a Docker container but shuts down the **host machine**, not the container.

## Goals

I needed a way to remotely shut down my home lab without physically accessing the machine to press the power button. This application enables shutdown triggers through automation integrations like Apple Shortcuts and Alexa Skills, allowing convenient remote management of my infrastructure from anywhere.

## Features

- Periodic Redis key monitoring (configurable interval)
- Automatic key deletion after detection
- Multiple shutdown methods (nsenter, sysrq-trigger, direct command)
- Opensearch logging for audit trail
- Lightweight Docker image (~15MB)
- Graceful signal handling

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Redis Server   │────▶│   signalmice    │────▶│  Host Machine   │
│  (Signal Key)   │     │   (Container)   │     │   (Shutdown)    │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │   Opensearch    │
                        │     (Logs)      │
                        └─────────────────┘
```

## Requirements

- Docker and Docker Compose
- Redis server
- Opensearch (optional, for logging)
- Linux host machine (Debian/Ubuntu based)

## Quick Start

### 1. Build the Docker image

```bash
docker build -t signalmice:latest .
```

### 2. Run with Docker Compose (Production)

```bash
# Set environment variables
export REDIS_HOST=your-redis-host
export REDIS_PASSWORD=your-redis-password
export OPENSEARCH_URL=http://your-opensearch:9200

# Run
docker-compose up -d
```

### 3. Development/Testing

```bash
# Run with local Redis and Opensearch
docker-compose -f docker-compose.dev.yml up -d
```

## Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis server hostname |
| `REDIS_PORT` | `6379` | Redis server port |
| `REDIS_PASSWORD` | `` | Redis password (empty for no auth) |
| `REDIS_DB` | `0` | Redis database number |
| `OPENSEARCH_URL` | `http://localhost:9200` | Opensearch URL |
| `OPENSEARCH_USERNAME` | `` | Opensearch username |
| `OPENSEARCH_PASSWORD` | `` | Opensearch password |
| `OPENSEARCH_INDEX` | `signalmice-logs` | Opensearch index base name for logs |
| `OPENSEARCH_USE_DAILY_INDEX` | `true` | Use date-based index names (e.g., `signalmice-logs-2024-12-28`) for ISM retention policies |
| `SIGNALMICE_KEY` | `signalmice:00000000-0000-0000-0000-000000000000` | Redis key to monitor |
| `SIGNALMICE_CHECK_INTERVAL` | `60` | Check interval in seconds |
| `HOST_PROC_PATH` | `/host/proc` | Path to host's /proc (mounted) |

## Triggering a Shutdown

To trigger a remote shutdown, simply set the monitored key in Redis:

```bash
# Using redis-cli
redis-cli SET "signalmice:00000000-0000-0000-0000-000000000000" "shutdown"

# Using a custom key
redis-cli SET "signalmice:my-machine-id" "shutdown"
```

The value can be anything - only the key's existence matters.

## Docker Container Requirements

The container needs special privileges to shutdown the host:

```yaml
services:
  signalmice:
    privileged: true      # Required for nsenter
    pid: host             # Required to access host PID namespace
    volumes:
      - /proc:/host/proc:ro  # Mount host's /proc
```

## Shutdown Methods

signalmice tries multiple shutdown methods in order:

1. **nsenter** (preferred): Enters host namespace and runs `poweroff`
2. **sysrq-trigger**: Writes to `/proc/sysrq-trigger` for clean shutdown
3. **direct command**: Runs `poweroff` or `shutdown -h now`

## Logs

### Stdout/Docker logs

```bash
docker logs signalmice
```

### Opensearch

Logs are sent to the configured Opensearch index with the following structure:

```json
{
  "@timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "Shutdown signal received! Key found and deleted.",
  "hostname": "container-hostname",
  "service": "signalmice",
  "redis_key": "signalmice:00000000-0000-0000-0000-000000000000"
}
```

### Log Retention

By default, signalmice uses date-based index names (e.g., `signalmice-logs-2024-12-28`) which enables automatic log retention via OpenSearch Index State Management (ISM) policies.

#### Setting Up 90-Day Log Retention

To automatically delete logs older than 90 days, create an ISM policy in OpenSearch:

```bash
# Create the ISM policy
curl -X PUT "https://your-opensearch:9200/_plugins/_ism/policies/signalmice-log-retention" \
  -H "Content-Type: application/json" \
  -u admin:password \
  -d '{
    "policy": {
      "description": "Delete signalmice logs after 90 days",
      "default_state": "hot",
      "states": [
        {
          "name": "hot",
          "actions": [],
          "transitions": [
            {
              "state_name": "delete",
              "conditions": {
                "min_index_age": "90d"
              }
            }
          ]
        },
        {
          "name": "delete",
          "actions": [
            {
              "delete": {}
            }
          ]
        }
      ],
      "ism_template": {
        "index_patterns": ["signalmice-logs-*"],
        "priority": 100
      }
    }
  }'
```

This policy will:
1. Apply to all indices matching `signalmice-logs-*`
2. Keep indices in the "hot" state initially
3. Automatically delete indices when they are older than 90 days

#### Disabling Date-Based Indexing

If you prefer a single static index (not recommended for production), set:

```bash
OPENSEARCH_USE_DAILY_INDEX=false
```

Note: With a static index, you'll need to manually manage log retention or use document-level cleanup.

## Security Considerations

- The container runs with `privileged: true` which grants full host access
- Use Redis authentication in production
- Consider network policies to restrict Redis access
- Use unique keys per machine for better control
- Monitor Opensearch logs for audit purposes

## Project Structure

```
signalmice/
├── cmd/
│   └── signalmice/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration management
│   │   └── config_test.go       # Config tests
│   ├── logger/
│   │   ├── logger.go            # Opensearch logging
│   │   └── logger_test.go       # Logger tests
│   ├── redis/
│   │   ├── client.go            # Redis client wrapper
│   │   └── client_test.go       # Redis client tests
│   └── shutdown/
│       ├── shutdown.go          # Host shutdown logic
│       └── shutdown_test.go     # Shutdown tests
├── PRPs/
│   └── features/
│       └── prp-signalmice-core.md  # Feature PRP documentation
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Production compose
├── docker-compose.dev.yml       # Development compose with Redis/Opensearch
├── go.mod                       # Go module definition
├── Makefile                     # Build, test, lint commands
├── .golangci.yml                # Linter configuration
├── .env.example                 # Environment variable template
├── .dockerignore
├── .gitignore
└── README.md
```

## Development

### Prerequisites

```bash
# Install Go 1.22+
# Install golangci-lint
make install-tools
```

### Building from Source

```bash
# Build locally
make build

# Build for Linux
make build-linux

# Or manually:
go build -o signalmice ./cmd/signalmice
GOOS=linux GOARCH=amd64 go build -o signalmice ./cmd/signalmice
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run short tests only
make test-short
```

### Linting

```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Run all checks (format, vet, lint, test)
make check
```

### Available Make Commands

```bash
make help  # Show all available commands
```

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run golangci-lint |
| `make lint-fix` | Run golangci-lint with auto-fix |
| `make docker` | Build Docker image |
| `make docker-dev` | Start development environment |
| `make check` | Run all checks (format, vet, lint, test) |
| `make clean` | Clean build artifacts |

## API Reference

### Core Functions

- `shutdown.NeutralizeStuartLittle(ctx)` - Main shutdown function that attempts host shutdown using multiple methods
- `redis.CheckAndDeleteKey(ctx)` - Check for signal key and delete if found
- `logger.Info/Warn/Error/Debug(ctx, message)` - Logging to Opensearch and stdout

## License

MIT
