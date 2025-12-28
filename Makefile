# signalmice Makefile
# Provides commands for building, testing, and linting

.PHONY: all build test lint fmt vet clean docker docker-dev help install-tools

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=signalmice
BINARY_PATH=./bin/$(BINARY_NAME)

# Docker parameters
DOCKER_IMAGE=signalmice
DOCKER_TAG=latest

# Build flags
LDFLAGS=-ldflags="-w -s"

# Default target
all: lint test build

## Build commands
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/signalmice

build-linux: ## Build for Linux AMD64
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)-linux-amd64 ./cmd/signalmice

## Test commands
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short: ## Run tests in short mode
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

## Lint commands
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Run 'make install-tools'" && exit 1)
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Run 'make install-tools'" && exit 1)
	golangci-lint run --fix ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

## Dependency commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## Docker commands
docker: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-dev: ## Run development environment with Docker Compose
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d

docker-dev-down: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down

docker-dev-logs: ## Show development environment logs
	docker-compose -f docker-compose.dev.yml logs -f

docker-run: ## Run production Docker Compose
	@echo "Starting production environment..."
	docker-compose up -d

docker-stop: ## Stop production Docker Compose
	@echo "Stopping production environment..."
	docker-compose down

## Tool installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully!"

## Clean commands
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

clean-docker: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker-compose down -v --rmi local
	docker-compose -f docker-compose.dev.yml down -v --rmi local

## Validation commands
validate: lint test ## Run all validation (lint + test)
	@echo "All validations passed!"

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
	@echo "All checks passed!"

## Help
help: ## Show this help
	@echo "signalmice - Remote Machine Shutdown Service"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
