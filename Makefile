.PHONY: build run test migrate clean test-messaging

# Variables
BINARY_NAME=only-flick-api
GO_FILES=$(shell find . -name "*.go" -type f)

# Build
build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

# Run application
run:
	go run cmd/api/main.go

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Run tests
test:
	go test -v ./...

# Run specific messaging tests
test-messaging:
	go test -v ./internal/services/...

# Run migrations (if using seed)
migrate:
	SEED=true go run cmd/api/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Database operations
db-reset:
	rm -f dev_database.db
	$(MAKE) migrate

# Test API endpoints
test-api:
	go run scripts/test_messaging_api.go

# Start development environment
dev-setup: deps
	@echo "🔧 Setting up development environment..."
	@echo "📦 Dependencies installed"
	@echo "🚀 Run 'make run' to start the server"
	@echo "🧪 Run 'make test-messaging' to test messaging system"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/api/main.go

# Run with Docker
docker-build:
	docker build -t only-flick-api .

docker-run:
	docker run -p 8080:8080 only-flick-api

# Help
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  dev         - Run with hot reload"
	@echo "  test        - Run all tests"
	@echo "  test-messaging - Run messaging tests only"
	@echo "  migrate     - Run with database seeding"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  lint        - Run linter"
	@echo "  db-reset    - Reset database"
	@echo "  test-api    - Test API endpoints"
	@echo "  dev-setup   - Setup development environment"
