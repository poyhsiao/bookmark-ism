#!/bin/bash

# Bookmark Sync Service Setup Script
# This script sets up the development environment

set -e

echo "🚀 Setting up Bookmark Sync Service development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    print_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or higher."
    exit 1
fi

print_success "All prerequisites are installed!"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    print_status "Creating .env file from template..."
    cp .env.example .env
    print_warning "Please edit .env file with your configuration before starting services"
else
    print_status ".env file already exists"
fi

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p logs
mkdir -p backups
mkdir -p supabase/migrations
mkdir -p nginx/logs
mkdir -p nginx/ssl
mkdir -p nginx/conf.d

# Download Go dependencies
print_status "Downloading Go dependencies..."
go mod download
go mod tidy

print_success "Go dependencies downloaded successfully!"

# Build the application
print_status "Building the application..."
if make build; then
    print_success "Application built successfully!"
else
    print_error "Failed to build application"
    exit 1
fi

# Create initial Supabase migration
print_status "Creating initial database migration..."
cat > supabase/migrations/001_initial_schema.sql << 'EOF'
-- Initial schema for Bookmark Sync Service
-- This will be expanded in the next task

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create initial tables (basic structure)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create basic indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Insert a comment to track migration
INSERT INTO pg_description (objoid, classoid, objsubid, description)
VALUES (
    (SELECT oid FROM pg_class WHERE relname = 'users'),
    (SELECT oid FROM pg_class WHERE relname = 'pg_class'),
    0,
    'Initial schema migration - v0.0.1'
) ON CONFLICT DO NOTHING;
EOF

print_success "Initial migration created!"

# Start infrastructure services
print_status "Starting infrastructure services..."
if docker-compose up -d supabase-db redis typesense minio; then
    print_success "Infrastructure services started!"
else
    print_error "Failed to start infrastructure services"
    exit 1
fi

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 10

# Check service health
print_status "Checking service health..."

# Check PostgreSQL
if docker-compose exec -T supabase-db pg_isready -U postgres > /dev/null 2>&1; then
    print_success "PostgreSQL is ready"
else
    print_warning "PostgreSQL is not ready yet"
fi

# Check Redis
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    print_success "Redis is ready"
else
    print_warning "Redis is not ready yet"
fi

# Check Typesense
if curl -s http://localhost:8108/health > /dev/null 2>&1; then
    print_success "Typesense is ready"
else
    print_warning "Typesense is not ready yet"
fi

# Check MinIO
if curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    print_success "MinIO is ready"
else
    print_warning "MinIO is not ready yet"
fi

echo ""
print_success "🎉 Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run 'make docker-up' to start all services"
echo "3. Run 'make run' to start the API server"
echo "4. Visit http://localhost:8080/health to check API health"
echo ""
echo "Useful commands:"
echo "  make docker-up    - Start all services"
echo "  make docker-down  - Stop all services"
echo "  make run          - Start API server"
echo "  make build        - Build application"
echo "  make test         - Run tests"
echo ""
echo "Service URLs:"
echo "  API Server:       http://localhost:8080"
echo "  Nginx:            http://localhost:80"
echo "  Supabase Auth:    http://localhost:9999"
echo "  Supabase REST:    http://localhost:3000"
echo "  Supabase Realtime: ws://localhost:4000"
echo "  Redis:            localhost:6379"
echo "  Typesense:        http://localhost:8108"
echo "  MinIO Console:    http://localhost:9001"
echo ""