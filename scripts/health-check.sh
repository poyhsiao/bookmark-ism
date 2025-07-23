#!/bin/bash

# Health check script for all services
# This script checks if all services are running and healthy

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

echo "🏥 Bookmark Sync Service - Health Check"
echo "======================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running"
    exit 1
fi

print_success "Docker is running"

# Check if services are running
print_status "Checking service containers..."

services=("supabase-db" "supabase-auth" "supabase-rest" "supabase-realtime" "redis" "typesense" "minio")
all_running=true

for service in "${services[@]}"; do
    if docker ps --format "table {{.Names}}" | grep -q "^$service$"; then
        print_success "$service container is running"
    else
        print_error "$service container is not running"
        all_running=false
    fi
done

if [ "$all_running" = false ]; then
    print_error "Some services are not running. Run 'make docker-up' to start them."
    exit 1
fi

echo ""
print_status "Checking service health endpoints..."

# Function to check HTTP endpoint
check_http() {
    local url=$1
    local service_name=$2
    local timeout=${3:-10}

    if curl -s --max-time $timeout "$url" > /dev/null 2>&1; then
        print_success "$service_name is healthy"
        return 0
    else
        print_error "$service_name is not responding"
        return 1
    fi
}

# Function to check HTTP endpoint with JSON response
check_http_json() {
    local url=$1
    local service_name=$2
    local timeout=${3:-10}

    response=$(curl -s --max-time $timeout "$url" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        print_success "$service_name is healthy"
        echo "    Response: $response"
        return 0
    else
        print_error "$service_name is not responding"
        return 1
    fi
}

# Check PostgreSQL
print_status "Checking PostgreSQL..."
if docker exec supabase-db pg_isready -U postgres > /dev/null 2>&1; then
    print_success "PostgreSQL is ready"
else
    print_error "PostgreSQL is not ready"
fi

# Check Redis
print_status "Checking Redis..."
if docker exec redis redis-cli ping > /dev/null 2>&1; then
    print_success "Redis is ready"
else
    print_error "Redis is not ready"
fi

# Check Supabase Auth
print_status "Checking Supabase Auth..."
check_http "http://localhost:9999/health" "Supabase Auth"

# Check Supabase REST
print_status "Checking Supabase REST..."
check_http "http://localhost:3000/" "Supabase REST"

# Check Supabase Realtime
print_status "Checking Supabase Realtime..."
check_http "http://localhost:4000/api/health" "Supabase Realtime"

# Check Typesense
print_status "Checking Typesense..."
check_http_json "http://localhost:8108/health" "Typesense"

# Check MinIO
print_status "Checking MinIO..."
check_http "http://localhost:9000/minio/health/live" "MinIO"

# Check API Server (if running)
print_status "Checking API Server..."
if check_http "http://localhost:8080/health" "API Server"; then
    # Get detailed health info
    api_health=$(curl -s http://localhost:8080/health 2>/dev/null)
    if [ -n "$api_health" ]; then
        echo "    API Health: $api_health"
    fi
fi

# Check Nginx (if running)
print_status "Checking Nginx..."
check_http "http://localhost:80/health" "Nginx"

echo ""
echo "🎯 Service URLs:"
echo "  API Server:       http://localhost:8080"
echo "  Nginx Gateway:    http://localhost:80"
echo "  Supabase Auth:    http://localhost:9999"
echo "  Supabase REST:    http://localhost:3000"
echo "  Supabase Realtime: ws://localhost:4000"
echo "  PostgreSQL:       localhost:5432"
echo "  Redis:            localhost:6379"
echo "  Typesense:        http://localhost:8108"
echo "  MinIO API:        http://localhost:9000"
echo "  MinIO Console:    http://localhost:9001"

echo ""
print_success "🎉 Health check complete!"