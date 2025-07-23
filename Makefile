# Bookmark Sync Service Makefile

.PHONY: help build run test clean deps docker-up docker-down setup

# Default target
help:
	@echo "🔖 Bookmark Sync Service - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  setup           - Initial setup of development environment"
	@echo "  build           - Build the Go application"
	@echo "  run             - Run the application locally"
	@echo "  dev             - Start development environment with hot reload"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download and tidy Go dependencies"
	@echo ""
	@echo "Docker & Services:"
	@echo "  docker-up       - Start all services with Docker Compose"
	@echo "  docker-down     - Stop all services"
	@echo "  docker-logs     - Show logs from all services"
	@echo "  docker-restart  - Restart all services"
	@echo "  docker-build    - Build Docker images"
	@echo "  init-buckets    - Initialize MinIO storage buckets"
	@echo ""
	@echo "Production:"
	@echo "  prod-up         - Start production environment"
	@echo "  prod-down       - Stop production environment"
	@echo "  prod-build      - Build production Docker images"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate      - Run database migrations"
	@echo "  db-seed         - Seed database with test data"
	@echo "  db-reset        - Reset database (WARNING: destructive)"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt             - Format Go code"
	@echo "  lint            - Run linter"
	@echo "  security        - Run security checks"
	@echo ""
	@echo "Utilities:"
	@echo "  logs            - Show application logs"
	@echo "  health          - Check service health"
	@echo "  docs            - Generate and serve documentation"

# Initial setup
setup:
	@echo "🚀 Setting up development environment..."
	./scripts/setup.sh

# Build the application
build:
	@echo "🔨 Building application..."
	go build -o bin/api ./backend/cmd/api
	@echo "✅ Build complete: bin/api"

# Run the application locally
run:
	@echo "🏃 Starting application..."
	go run ./backend/cmd/api/main.go

# Development environment with hot reload
dev:
	@echo "🔥 Starting development environment..."
	@echo "💡 Make sure to run 'make docker-up' first to start dependencies"
	@echo "🌐 API will be available at http://localhost:8080"
	go run ./backend/cmd/api/main.go

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "📊 Coverage report: coverage.out"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
	go clean
	docker system prune -f

# Download and tidy dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

# Start all services with Docker Compose
docker-up:
	@echo "🐳 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "🎉 Services started! Run 'make health' to check status"

# Stop all services
docker-down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "✅ All services stopped"

# Show logs from all services
docker-logs:
	@echo "📋 Showing logs from all services..."
	docker-compose logs -f

# Restart all services
docker-restart:
	@echo "🔄 Restarting all services..."
	docker-compose restart
	@echo "✅ Services restarted"

# Build Docker images
docker-build:
	@echo "🏗️ Building Docker images..."
	docker-compose build
	@echo "✅ Docker images built"

# Initialize MinIO buckets
init-buckets:
	@echo "🪣 Initializing storage buckets..."
	./scripts/init-buckets.sh

# Production environment
prod-up:
	@echo "🚀 Starting production environment..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ Production environment started"

prod-down:
	@echo "🛑 Stopping production environment..."
	docker-compose -f docker-compose.prod.yml down
	@echo "✅ Production environment stopped"

prod-build:
	@echo "🏗️ Building production images..."
	docker-compose -f docker-compose.prod.yml build
	@echo "✅ Production images built"

# Database operations (placeholder for future implementation)
db-migrate:
	@echo "🗃️ Running database migrations..."
	@echo "⚠️ Database migrations will be implemented in the next task"

db-seed:
	@echo "🌱 Seeding database..."
	@echo "⚠️ Database seeding will be implemented in the next task"

db-reset:
	@echo "⚠️ This will reset the database and delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "🗃️ Resetting database..."; \
		docker-compose exec supabase-db psql -U postgres -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		echo "✅ Database reset complete"; \
	else \
		echo ""; \
		echo "❌ Database reset cancelled"; \
	fi

# Code quality
fmt:
	@echo "🎨 Formatting Go code..."
	go fmt ./...
	@echo "✅ Code formatted"

lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️ gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Utilities
logs:
	@echo "📋 Showing application logs..."
	docker-compose logs -f api

health:
	@echo "🏥 Checking service health..."
	@echo "API Server:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ API Server not responding"
	@echo ""
	@echo "Supabase Auth:"
	@curl -s http://localhost:9999/health || echo "❌ Supabase Auth not responding"
	@echo ""
	@echo "Typesense:"
	@curl -s http://localhost:8108/health | jq . || echo "❌ Typesense not responding"
	@echo ""
	@echo "MinIO:"
	@curl -s http://localhost:9000/minio/health/live || echo "❌ MinIO not responding"
	@echo ""

docs:
	@echo "📚 Generating and serving documentation..."
	@echo "🌐 Documentation will be available at http://localhost:6060"
	godoc -http=:6060

# Initialize Go modules (for new projects)
init:
	@echo "🆕 Initializing Go modules..."
	go mod init bookmark-sync-service
	go mod tidy
	@echo "✅ Go modules initialized"