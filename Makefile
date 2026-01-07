# Project variables
BINARY_NAME := awsl
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/awsl
GO := go

# Version info (optional, useful later)
VERSION := 0.1.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.DEFAULT_GOAL := help

# ==================== Build ====================

.PHONY: build
build: ## Build the binary
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-release
build-release: ## Build optimized release binary
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: install
install: ## Install binary to GOPATH/bin
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

# ==================== Development ====================

.PHONY: run
run: build ## Build and run with test.awsl
	$(BUILD_DIR)/$(BINARY_NAME) test.awsl

.PHONY: run-file
run-file: build ## Run with specific file: make run-file FILE=script.awsl
	$(BUILD_DIR)/$(BINARY_NAME) $(FILE)

.PHONY: watch
watch: ## Rebuild on file changes (requires entr)
	find . -name '*.go' | entr -r make run

# ==================== Testing ====================

.PHONY: test
test: ## Run all tests
	$(GO) test -v ./...

.PHONY: test-short
test-short: ## Run tests without verbose output
	$(GO) test ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: bench
bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem ./...

# ==================== Code Quality ====================

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	golangci-lint run

.PHONY: tidy
tidy: ## Tidy go.mod
	$(GO) mod tidy

.PHONY: check
check: fmt vet test ## Run all checks (fmt, vet, test)

# ==================== Cleanup ====================

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# ==================== Help ====================

.PHONY: help
help: ## Show this help
	@echo "AWSL - AWS Scripting Language"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
