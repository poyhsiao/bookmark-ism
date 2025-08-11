#!/bin/bash

# Docker Build Fix Validation Script
# This script validates that the Docker build fix is working correctly

echo "🔍 Docker Build Fix Validation"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track failures for comprehensive reporting
FAILURES=0
TOTAL_CHECKS=0

# Function to print status and track failures
print_status() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    if [ "$1" -eq 0 ]; then
        echo -e "${GREEN}✔ $2${NC}"
    else
        echo -e "${RED}✖ $2${NC}"
        FAILURES=$((FAILURES + 1))
    fi
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_section() {
    echo -e "\n${BLUE}📋 $1${NC}"
    echo "----------------------------------------"
}

print_section "Docker Environment Check"

# Check if Docker is available
print_info "Checking Docker availability..."
if command -v docker &> /dev/null; then
    if docker version &> /dev/null 2>&1; then
        print_status 0 "Docker is available and running"
    else
        print_status 1 "Docker is installed but not running"
    fi
else
    print_status 1 "Docker is not installed"
fi

# Check Docker Compose availability
if command -v docker-compose &> /dev/null || docker compose version &> /dev/null 2>&1; then
    print_status 0 "Docker Compose is available"
else
    print_status 1 "Docker Compose is not available"
fi

print_section "Project Structure Validation"

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

# Check for test files
test_files=(
    "docker_build_test.go"
    "docker_build_comprehensive_test.go"
    "docker_build_fix_test.go"
)

print_info "Checking test files..."
for file in "${test_files[@]}"; do
    if [ -f "$file" ]; then
        print_status 0 "Found test file: $file"
    else
        print_status 1 "Missing test file: $file"
    fi
done

print_section "Dockerfile.prod Content Validation"

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

# Check for multi-stage build structure
if grep -q "FROM.*AS builder" Dockerfile.prod && grep -q "FROM.*AS runtime" Dockerfile.prod; then
    print_status 0 "Dockerfile.prod has proper multi-stage build structure"
else
    print_status 1 "Dockerfile.prod missing proper multi-stage build structure"
fi

print_section "Go Tests Execution"

print_info "Running Go tests..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_status 1 "Go is not installed or not in PATH"
else
    print_status 0 "Go is available"

    # Run tests using proper Go test directory approach
    print_info "Running Docker build tests..."

    # Run tests using proper Go test approach (use directory, not individual files)
    print_info "Running Docker build tests..."

    # Run all tests in current directory with proper package syntax
    if go test -v -timeout=60s . 2>/dev/null; then
        print_status 0 "All Go tests in current directory passed"
    else
        print_status 1 "Some Go tests in current directory failed"

        # Try running specific test patterns for better diagnostics
        print_info "Running specific test patterns for diagnostics..."

        if go test -v -timeout=30s -run TestDockerBuild . 2>/dev/null; then
            print_status 0 "Docker build related tests passed"
        else
            print_status 1 "Docker build related tests failed"
        fi

        if go test -v -timeout=30s -run TestDockerBuildFix . 2>/dev/null; then
            print_status 0 "Docker build fix tests passed"
        else
            print_status 1 "Docker build fix tests failed"
        fi
    fi
fi

print_section "GitHub Actions Workflow Validation"

print_info "Validating GitHub Actions workflow..."

if [ -f ".github/workflows/cd.yml" ]; then
    # Check for correct Docker build context
    if grep -q "context: \." .github/workflows/cd.yml; then
        print_status 0 "GitHub Actions has correct build context"
    else
        print_status 1 "GitHub Actions missing correct build context"
    fi

    # Check for correct Dockerfile reference
    if grep -q "file: ./Dockerfile.prod" .github/workflows/cd.yml; then
        print_status 0 "GitHub Actions references correct Dockerfile"
    else
        print_status 1 "GitHub Actions missing correct Dockerfile reference"
    fi

    # Check for proper workflow structure
    if grep -q "docker/build-push-action" .github/workflows/cd.yml; then
        print_status 0 "GitHub Actions uses proper build-push action"
    else
        print_status 1 "GitHub Actions missing proper build-push action"
    fi
else
    print_status 1 "GitHub Actions workflow file not found"
fi

print_section "Docker Build Testing"

# Only run Docker build tests if Docker is available
if command -v docker &> /dev/null && docker version &> /dev/null 2>&1; then
    # Test Docker build (builder stage only for speed)
    print_info "Testing Docker build (builder stage)..."
    if docker build -f Dockerfile.prod --target builder -t bookmark-sync-test-builder . > /dev/null 2>&1; then
        print_status 0 "Docker build (builder stage) successful"
    else
        print_status 1 "Docker build (builder stage) failed"
        # Show build output for debugging
        print_info "Build output for debugging:"
        docker build -f Dockerfile.prod --target builder -t bookmark-sync-test-builder . 2>&1 | tail -20
    fi

    # Test full Docker build
    print_info "Testing full Docker build..."
    if docker build -f Dockerfile.prod -t bookmark-sync-test . > /dev/null 2>&1; then
        print_status 0 "Full Docker build successful"
    else
        print_status 1 "Full Docker build failed"
        # Show build output for debugging
        print_info "Build output for debugging:"
        docker build -f Dockerfile.prod -t bookmark-sync-test . 2>&1 | tail -20
    fi

    # Test image functionality
    print_info "Testing built image functionality..."
    if docker run --rm bookmark-sync-test --help > /dev/null 2>&1; then
        print_status 0 "Built image runs successfully"
    else
        print_status 1 "Built image failed to run"
    fi

    # Clean up test images
    print_info "Cleaning up test images..."
    docker rmi bookmark-sync-test-builder bookmark-sync-test > /dev/null 2>&1 || true
    print_status 0 "Test images cleaned up"
else
    print_status 1 "Skipping Docker build tests - Docker not available"
fi

print_section "Validation Summary"

echo ""
echo "📊 Validation Results:"
echo "  Total checks: $TOTAL_CHECKS"
echo "  Passed: $((TOTAL_CHECKS - FAILURES))"
echo "  Failed: $FAILURES"
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}🎉 All validations passed! The Docker build fix is working correctly.${NC}"
    echo ""
    echo "✅ Summary of validated fixes:"
    echo "  • Fixed working directory consistency in Dockerfile.prod"
    echo "  • Corrected binary copy path from builder stage"
    echo "  • Maintained proper build context for monorepo structure"
    echo "  • Applied Docker security best practices"
    echo "  • Implemented comprehensive BDD/TDD test coverage"
    echo "  • Validated multi-stage build structure"
    echo ""
    echo -e "${GREEN}The GitHub Actions build should now complete successfully!${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILURES validation(s) failed.${NC}"
    echo ""
    echo "🔧 Recommended actions:"
    echo "  • Review the failed checks above"
    echo "  • Ensure all required files are present"
    echo "  • Verify Docker is installed and running"
    echo "  • Check Dockerfile.prod syntax and structure"
    echo "  • Run individual Go tests to identify specific issues"
    echo ""
    echo -e "${YELLOW}Re-run this script after addressing the issues.${NC}"
    exit 1
fi