# Flight Search Aggregation System - Makefile
# Cross-platform build automation for Go project

# ==============================================================================
# Variables
# ==============================================================================

# Binary configuration
BINARY_NAME := flight-search
BUILD_DIR := bin
CMD_DIR := cmd/server

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GORUN := $(GOCMD) run
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOGEN := $(GOCMD) generate

# Build flags - strip debug info for smaller binary
LDFLAGS := -ldflags "-s -w"

# Coverage configuration
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# ==============================================================================
# Default target
# ==============================================================================

.DEFAULT_GOAL := help

# ==============================================================================
# Build targets
# ==============================================================================

.PHONY: all
all: fmt lint test build ## Run fmt, lint, test, and build

.PHONY: build
build: ## Build the application binary
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "==> Binary created at $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-debug
build-debug: ## Build with debug symbols
	@echo "==> Building $(BINARY_NAME) (debug)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(CMD_DIR)/main.go

# ==============================================================================
# Run targets
# ==============================================================================

.PHONY: run
run: ## Run the application
	@echo "==> Running $(BINARY_NAME)..."
	$(GORUN) $(CMD_DIR)/main.go

.PHONY: run-dev
run-dev: ## Run with development settings
	@echo "==> Running $(BINARY_NAME) in development mode..."
	LOG_LEVEL=debug LOG_FORMAT=console $(GORUN) $(CMD_DIR)/main.go

# ==============================================================================
# Test targets
# ==============================================================================

.PHONY: test
test: ## Run all tests
	@echo "==> Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests in short mode (skip long tests)
	@echo "==> Running short tests..."
	$(GOTEST) -short -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@echo "==> Running tests with coverage..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "==> Coverage report: $(COVERAGE_HTML)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "==> Running tests with race detector..."
	$(GOTEST) -race -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "==> Running integration tests..."
	$(GOTEST) -v ./test/integration/...

.PHONY: bench
bench: ## Run benchmarks
	@echo "==> Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# ==============================================================================
# Code quality targets
# ==============================================================================

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "==> Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "==> Formatting code..."
	$(GOFMT) ./...

.PHONY: vet
vet: ## Run go vet
	@echo "==> Running go vet..."
	$(GOVET) ./...

.PHONY: check
check: fmt vet lint ## Run all code checks (fmt, vet, lint)

# ==============================================================================
# Dependency management
# ==============================================================================

.PHONY: deps
deps: ## Download dependencies
	@echo "==> Downloading dependencies..."
	$(GOMOD) download

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "==> Tidying modules..."
	$(GOMOD) tidy

.PHONY: verify
verify: ## Verify dependencies
	@echo "==> Verifying dependencies..."
	$(GOMOD) verify

# ==============================================================================
# Code generation
# ==============================================================================

.PHONY: generate
generate: ## Run go generate
	@echo "==> Running go generate..."
	$(GOGEN) ./...

.PHONY: mocks
mocks: ## Generate mocks using mockgen
	@echo "==> Generating mocks..."
	$(GOGEN) ./internal/domain/...
	$(GOGEN) ./internal/usecase/...

.PHONY: swagger
swagger: ## Generate Swagger/OpenAPI documentation
	@echo "==> Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal; \
		echo "==> Swagger docs generated at docs/swagger.json"; \
	else \
		echo "swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

# ==============================================================================
# Cleanup targets
# ==============================================================================

.PHONY: clean
clean: ## Remove build artifacts
	@echo "==> Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "==> Clean complete"

.PHONY: clean-cache
clean-cache: ## Clean Go build cache
	@echo "==> Cleaning build cache..."
	$(GOCMD) clean -cache -testcache

# ==============================================================================
# Docker targets (optional)
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "==> Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "==> Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME):latest

# ==============================================================================
# Help target
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Flight Search Aggregation System - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
