# Makefile for Cryptocurrency Match Engine Project

# Build variables
BINARY_NAME=matchengine
BUILD_DIR=bin
MAIN_FILE=cmd/api/main.go
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
COVERAGE_FILE=coverage.out
MIN_COVERAGE=85

# Go commands
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod

# Docker commands
DOCKER=docker
DOCKER_COMPOSE=docker-compose

# Ensure commands run even if files exist with same names
.PHONY: all build clean test coverage lint run help benchmark integration-test security-scan

# Default target
all: lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN) -cache
	@rm -f $(COVERAGE_FILE)
	@echo "Clean complete"

# Run unit tests
test:
	@echo "Running unit tests..."
	@$(GOTEST) -v ./... -cover
	@echo "Tests complete"

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@$(GOTEST) ./... -coverprofile=$(COVERAGE_FILE)
	@$(GO) tool cover -html=$(COVERAGE_FILE)
	@coverage_value=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$coverage_value < $(MIN_COVERAGE)" | bc) -eq 1 ]; then \
		echo "Coverage $$coverage_value% is below minimum $(MIN_COVERAGE)%"; \
		exit 1; \
	fi

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Lint complete"

# Run the application
run:
	@echo "Starting match engine..."
	@$(GO) run $(MAIN_FILE)

# Run benchmark tests
benchmark:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...
	@echo "Benchmarks complete"

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@$(DOCKER_COMPOSE) up -d kafka
	@$(GOTEST) -tags=integration ./...
	@$(DOCKER_COMPOSE) down
	@echo "Integration tests complete"

# Run security scan
security-scan:
	@echo "Running security scan..."
	@trivy filesystem --severity HIGH,CRITICAL .
	@gosec ./...
	@echo "Security scan complete"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@$(GOMOD) tidy
	@$(GOMOD) verify
	@echo "Dependencies updated"

# Run with race detection
race-test:
	@echo "Running tests with race detection..."
	@$(GOTEST) -race ./...
	@echo "Race detection complete"

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@swag init -g $(MAIN_FILE)
	@echo "Documentation generated"

# Help command
help:
	@echo "Available commands:"
	@echo "  make all             - Run lint, test and build"
	@echo "  make build           - Build the application"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make test            - Run unit tests"
	@echo "  make coverage        - Generate test coverage report"
	@echo "  make lint            - Run linter"
	@echo "  make run             - Run the application"
	@echo "  make benchmark       - Run benchmark tests"
	@echo "  make integration-test- Run integration tests"
	@echo "  make security-scan   - Run security scan"
	@echo "  make deps            - Update dependencies"
	@echo "  make race-test       - Run tests with race detection"
	@echo "  make docs            - Generate API documentation" 