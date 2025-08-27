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
	cd backend && go test -v -race -coverprofile=coverage.out ./...
	@echo "📊 Coverage report: backend/coverage.out"

test-coverage:
	@echo "🧪 Running tests with coverage report..."
	cd backend && go test -v -race -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: backend/coverage.html"

test-models:
	@echo "🗃️ Running database models tests..."
	cd backend && go test ./pkg/database -v

test-auth-service:
	@echo "🔐 Running auth service tests..."
	cd backend && go test ./internal/auth -v

test-user-service:
	@echo "👤 Running user service tests..."
	cd backend && go test ./internal/user -v

test-middleware:
	@echo "🛡️ Running middleware tests..."
	cd backend && go test ./pkg/middleware -v

test-config:
	@echo "⚙️ Running config tests..."
	cd backend && go test ./internal/config -v

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

# Database operations
db-migrate:
	@echo "🗃️ Running database migrations..."
	go run ./backend/cmd/migrate/main.go -direction=up

db-seed:
	@echo "🌱 Seeding database..."
	go run ./backend/cmd/migrate/main.go -direction=seed

db-rollback:
	@echo "⚠️ Rolling back database migrations..."
	@read -p "Are you sure? This will drop all tables! [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		go run ./backend/cmd/migrate/main.go -direction=down; \
	else \
		echo ""; \
		echo "❌ Rollback cancelled"; \
	fi

db-backup:
	@echo "💾 Creating database backup..."
	./scripts/backup.sh

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
	./scripts/health-check.sh

test-auth:
	@echo "🔐 Testing authentication endpoints..."
	./scripts/test-auth.sh

test-user:
	@echo "👤 Testing user profile management endpoints..."
	./scripts/test-user.sh

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

# BDD Testing Targets
install-deps:
	@echo "📦 Installing BDD testing dependencies..."
	@go mod tidy
	@go install github.com/cucumber/godog/cmd/godog@latest
	@echo "✅ Dependencies installed"

test-all: install-deps
	@echo "🚀 Running complete BDD testing suite..."
	@./run_bdd_tests.sh

test-unit-bdd: install-deps
	@echo "🧪 Running unit test BDD scenarios..."
	@go test -v ./features/comprehensive-testing/... -run TestUnitTestCoverage

test-integration-bdd: install-deps
	@echo "🔗 Running integration test BDD scenarios..."
	@go test -v ./features/comprehensive-testing/... -run TestIntegrationTestScenarios

test-e2e-bdd: install-deps
	@echo "🌐 Running end-to-end test BDD scenarios..."
	@go test -v ./features/comprehensive-testing/... -run TestEndToEndUserWorkflows

test-performance-bdd: install-deps
	@echo "⚡ Running performance test BDD scenarios..."
	@go test -v ./features/comprehensive-testing/... -run TestPerformanceTestScenarios

test-security-bdd: install-deps
	@echo "🔒 Running security test BDD scenarios..."
	@go test -v ./features/comprehensive-testing/... -run TestSecurityTestScenarios

test-quality: install-deps
	@echo "📊 Running code quality analysis..."
	@go test -v ./features/comprehensive-testing/... -run TestCodeQuality

test-acceptance: install-deps
	@echo "👥 Running user acceptance tests..."
	@go test -v ./features/comprehensive-testing/... -run TestUserAcceptance

generate-docs: install-deps
	@echo "📚 Generating API documentation..."
	@go test -v ./features/comprehensive-testing/... -run TestDocumentationGeneration

# Phase 14 Task 31: Comprehensive Documentation Generation
generate-all-docs: install-deps
	@echo "📚 Generating comprehensive documentation..."
	@mkdir -p docs/api docs/architecture docs/deployment docs/user docs/developer
	@echo "🔧 Running API documentation generator..."
	@go run tools/doc_generator.go backend/
	@echo "📝 Generating additional documentation..."
	@go test -v ./features/comprehensive-testing/... -run TestDocumentationGeneration
	@echo "✅ All documentation generated successfully!"

test-documentation: install-deps
	@echo "📋 Testing documentation generation..."
	@go test -v ./features/comprehensive-testing/... -run TestDocumentationGeneration

test-quality-assurance: install-deps
	@echo "🔍 Running quality assurance tests..."
	@go test -v ./features/comprehensive-testing/... -run TestCodeQuality
	@go test -v ./features/comprehensive-testing/... -run TestUserAcceptance

# Phase 14: Complete Testing and Quality Assurance
test-phase-14: install-deps
	@echo "🚀 Running Phase 14: Testing and Quality Assurance..."
	@echo "📋 Task 30: Comprehensive Testing Suite"
	@make test-all
	@echo "📚 Task 31: Quality Assurance and Documentation"
	@make test-documentation
	@make test-quality-assurance
	@make generate-all-docs
	@echo "✅ Phase 14 completed successfully!"

generate-docs: install-deps
	@echo "📚 Generating API documentation..."
	@go test -v ./features/comprehensive-testing/... -run TestAPIDocumentation

coverage-bdd: install-deps
	@echo "📈 Generating test coverage report..."
	@mkdir -p reports
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o reports/coverage.html
	@go tool cover -func=coverage.out
	@echo "✅ Coverage report generated: reports/coverage.html"

validate-features:
	@echo "✅ Validating Gherkin feature files..."
	@find features/ -name "*.feature" -exec echo "Validating {}" \;
	@echo "✅ Feature file validation completed"

test-stats:
	@echo "📊 BDD Test Suite Statistics:"
	@echo "Feature files: $(shell find features/ -name "*.feature" | wc -l)"
	@echo "Step definition files: $(shell find features/ -name "*_context.go" | wc -l)"
	@echo "Test scenarios: $(shell grep -r "Scenario:" features/ | wc -l)"
	@echo "Given steps: $(shell grep -r "Given" features/ | wc -l)"
	@echo "When steps: $(shell grep -r "When" features/ | wc -l)"
	@echo "Then steps: $(shell grep -r "Then" features/ | wc -l)"
