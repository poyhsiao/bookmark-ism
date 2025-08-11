#!/bin/bash

# Docker Build Fix Validation Script
# This script validates that the Docker build fix is working correctly

set -e

echo "🔍 Docker Build Fix Validation"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check if Docker is available
print_info "Checking Docker availability..."
if command -v docker &> /dev/null; then
    if docker version &> /dev/null; then
        print_status 0 "Docker is available and running"
    else
        print_status 1 "Docker is installed but not running"
    fi
else
    print_status 1 "Docker is not installed"
fi

# Check project structure
print_info "Validating project structure..."
required_files=(
    "go.mod"
    "go.sum"
    "backend/cmd/api/main.go"
    "Dockerfile"
    "Dockerfile.prod"
    ".github/workflows/cd.yml"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        print_status 0 "Found required file: $file"
    else
        print_status 1 "Missing required file: $file"
    fi
done

# Validate Dockerfile.prod content
print_info "Validating Dockerfile.prod content..."

# Check for correct working directory
if grep -q "WORKDIR /build" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has correct working directory"
else
    print_status 1 "Dockerfile.prod missing correct working directory"
fi

# Check for correct copy commands
if grep -q "COPY go.mod go.sum ./" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has correct go.mod copy command"
else
    print_status 1 "Dockerfile.prod missing correct go.mod copy command"
fi

if grep -q "COPY backend ./backend" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has correct backend copy command"
else
    print_status 1 "Dockerfile.prod missing correct backend copy command"
fi

# Check for correct build path
if grep -q "./backend/cmd/api" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has correct build path"
else
    print_status 1 "Dockerfile.prod missing correct build path"
fi

# Check for correct binary copy
if grep -q "COPY --from=builder /build/main" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has correct binary copy command"
else
    print_status 1 "Dockerfile.prod missing correct binary copy command"
fi

# Run Go tests
print_info "Running Go tests..."
if go test -v ./docker_build_comprehensive_test.go -run TestDockerBuildFix > /dev/null 2>&1; then
    print_status 0 "Go tests passed"
else
    print_status 1 "Go tests failed"
fi

# Test Docker build (builder stage only for speed)
print_info "Testing Docker build (builder stage)..."
if docker build -f Dockerfile.prod --target builder -t bookmark-sync-test-builder . > /dev/null 2>&1; then
    print_status 0 "Docker build (builder stage) successful"
else
    print_status 1 "Docker build (builder stage) failed"
fi

# Test full Docker build
print_info "Testing full Docker build..."
if docker build -f Dockerfile.prod -t bookmark-sync-test . > /dev/null 2>&1; then
    print_status 0 "Full Docker build successful"
else
    print_status 1 "Full Docker build failed"
fi

# Clean up test images
print_info "Cleaning up test images..."
docker rmi bookmark-sync-test-builder bookmark-sync-test > /dev/null 2>&1 || true

echo ""
echo -e "${GREEN}🎉 All validations passed! The Docker build fix is working correctly.${NC}"
echo ""
echo "Summary of fixes applied:"
echo "- Fixed working directory consistency in Dockerfile.prod"
echo "- Corrected binary copy path from builder stage"
echo "- Maintained proper build context for monorepo structure"
echo "- Applied Docker security best practices"
echo "- Implemented comprehensive BDD/TDD test coverage"
echo ""
echo "The GitHub Actions build should now complete successfully!"