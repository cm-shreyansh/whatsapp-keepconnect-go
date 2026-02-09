.PHONY: help build run test clean docker-build docker-up docker-down migrate

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Build complete: bin/server"

run: ## Run the application
	@echo "Starting application..."
	@go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t whatsapp-chatbot-go:latest .
	@echo "✅ Docker image built"

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	@docker-compose up -d
	@echo "✅ Services started"

docker-down: ## Stop services with Docker Compose
	@echo "Stopping services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-logs: ## View Docker logs
	@docker-compose logs -f

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/server/main.go migrate
	@echo "✅ Migrations complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "✅ Linting complete"

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

dev: ## Run in development mode with auto-reload
	@echo "Starting development server..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@air

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Tools installed"
