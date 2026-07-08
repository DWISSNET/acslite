BINARY_NAME=acsgo
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: all build run dev test lint clean install

## Default: build
all: build

## Build the production binary
build:
	@echo "⚙️  Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## Run directly (development)
run:
	@go run .

## Run with hot-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@air

## Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -count=1 ./tests/...

## Run tests with coverage
coverage:
	@go test -v -coverprofile=coverage.out ./tests/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

## Lint (requires golangci-lint)
lint:
	@golangci-lint run ./...

## Tidy modules
tidy:
	@go mod tidy

## Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

## Cross-compile for Linux amd64
build-linux:
	@echo "🐧 Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

## Install binary to /usr/local/bin
install: build
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"

## Show help
help:
	@echo "ACSGO Makefile targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
