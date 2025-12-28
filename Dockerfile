# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o signalmice \
    ./cmd/signalmice

# Final stage - using scratch for minimal image
FROM alpine:3.19

# Install necessary tools for shutdown commands and nsenter
RUN apk add --no-cache \
    ca-certificates \
    util-linux \
    && rm -rf /var/cache/apk/*

# Create non-root user (though we'll need root for shutdown)
# We keep root as default because shutdown requires privileges

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/signalmice .

# Environment variables with defaults
ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379
ENV REDIS_PASSWORD=""
ENV REDIS_DB=0
ENV OPENSEARCH_URL=http://localhost:9200
ENV OPENSEARCH_USERNAME=""
ENV OPENSEARCH_PASSWORD=""
ENV OPENSEARCH_INDEX=signalmice-logs
ENV SIGNALMICE_KEY=signalmice:00000000-0000-0000-0000-000000000000
ENV SIGNALMICE_CHECK_INTERVAL=60
ENV HOST_PROC_PATH=/host/proc

# Run the application
ENTRYPOINT ["/app/signalmice"]
