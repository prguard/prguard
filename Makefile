.PHONY: build clean install test lint help

# Build variables
BINARY_NAME=prguard
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/prguard

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f *.db *.db-shm *.db-wal
	@rm -rf exports/

install: build ## Build and install to $GOPATH/bin
	@echo "Installing to $(GOPATH)/bin/$(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/prguard

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

lint: ## Run linters
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

run: build ## Build and run with example config
	@./$(BINARY_NAME) --help
