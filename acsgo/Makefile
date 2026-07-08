.PHONY: build run test lint docker clean schema-apply help

BINARY    := acsgo
MAIN      := ./cmd/acsgo/main.go
BUILD_DIR := ./bin

## help: Show this help message
help:
	@echo "ACSGO Makefile"
	@echo "=============="
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: Compile the Go binary
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(MAIN)
	@echo "✅ Binary: $(BUILD_DIR)/$(BINARY)"

## run: Run the server locally (requires .env)
run:
	@test -f .env || (echo "❌ .env not found — copy .env.example and fill in values" && exit 1)
	@export $$(cat .env | grep -v '^#' | xargs) && go run $(MAIN)

## docker: Build the Docker image
docker:
	docker build -t acsgo:latest .

## docker-up: Start all services with Docker Compose
docker-up:
	docker compose up -d

## docker-down: Stop all services
docker-down:
	docker compose down

## schema-apply: Apply PostgreSQL schema (requires DATABASE_URL env var)
schema-apply:
	psql "$$DATABASE_URL" -f schema/001_initial.sql
	@echo "✅ Schema applied"

## test: Run all tests
test:
	go test ./... -v -race -timeout 60s

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## tidy: Tidy Go modules
tidy:
	go mod tidy

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	@echo "✅ Cleaned"

## deps: Install development tools
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
