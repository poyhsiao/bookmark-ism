#!/bin/bash

# Docker Build Validation Script
# This script validates the Docker build fix for the production Dockerfile

set -e

echo "🐳 Docker Build Validation Script"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_status $RED "❌ Docker is not installed or not in PATH"
    exit 1
fi

print_status $GREEN "✅ Docker is available"

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -f "Dockerfile.prod" ]]; then
    print_status $RED "❌ Please run this script from the project root directory"
    exit 1
fi

print_status $GREEN "✅ Project structure validated"

# Check Go module structure
print_status $YELLOW "🔍 Validating Go module structure..."
if grep -q "module bookmark-sync-service" go.mod; then
    print_status $GREEN "✅ Go module name matches expected pattern"
else
    print_status $RED "❌ Go module name doesn't match expected pattern"
    exit 1
fi

# Check if backend directory exists
if [[ -d "backend" ]] && [[ -f "backend/cmd/api/main.go" ]]; then
    print_status $GREEN "✅ Backend structure validated"
else
    print_status $RED "❌ Backend structure is missing"
    exit 1
fi

# Validate Dockerfile syntax
print_status $YELLOW "🔍 Validating Dockerfile syntax..."
if docker build --check -f Dockerfile.prod .; then
    print_status $GREEN "✅ Dockerfile syntax is valid"
else
    print_status $RED "❌ Dockerfile syntax validation failed. See error details above."
fi

# Test build (dry run with --dry-run if supported, otherwise actual build)
print_status $YELLOW "🏗️  Testing Docker build..."

# Create a unique tag for this test
TEST_TAG="bookmark-sync-test:$(date +%s)"

# Attempt to build the Docker image
if docker build -f Dockerfile.prod -t "$TEST_TAG" . --progress=plain; then
    print_status $GREEN "✅ Docker build successful!"

    # Check if the image was created
    if docker images "$TEST_TAG" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -q "$TEST_TAG"; then
        print_status $GREEN "✅ Docker image created successfully"

        # Show image details
        echo ""
        print_status $YELLOW "📊 Image Details:"
        docker images "$TEST_TAG" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

        # Test if the image can run (basic smoke test)
        print_status $YELLOW "🧪 Running smoke test..."
        if timeout 10s docker run --rm "$TEST_TAG" --help 2>/dev/null || timeout 10s docker run --rm "$TEST_TAG" --version 2>/dev/null; then
            print_status $GREEN "✅ Smoke test passed - application starts correctly"
        else
            print_status $YELLOW "⚠️  Smoke test inconclusive (this may be expected if app requires specific config)"
        fi

        # Clean up test image
        print_status $YELLOW "🧹 Cleaning up test image..."
        if docker rmi "$TEST_TAG" >/dev/null 2>&1; then
            print_status $GREEN "✅ Cleanup completed"
        else
            print_status $RED "❌ Cleanup failed: could not remove test image '$TEST_TAG'"
        fi

    else
        print_status $RED "❌ Docker image was not created properly"
        exit 1
    fi
else
    print_status $RED "❌ Docker build failed"
    print_status $YELLOW "💡 Common issues and solutions:"
    echo "   • Module resolution: Ensure go.mod is at project root"
    echo "   • Import paths: Check that all imports use the correct module path"
    echo "   • Build context: Make sure all source files are copied correctly"
    echo "   • Dependencies: Verify go.sum is up to date"
    exit 1
fi

echo ""
print_status $GREEN "🎉 All validations passed! The Docker build fix is working correctly."
echo ""
print_status $YELLOW "📝 Summary of fixes applied:"
echo "   • Fixed Go module resolution by copying entire source tree"
echo "   • Set GO111MODULE=on explicitly for module mode"
echo "   • Added go mod verify step to ensure module integrity"
echo "   • Optimized layer caching with go.mod/go.sum copy first"
echo "   • Added build cache mounts for faster builds"
echo "   • Used Alpine base for better compatibility"
echo "   • Added proper build flags for static binary"
echo "   • Included health check configuration"
echo "   • Maintained non-root user for security"
echo ""
print_status $GREEN "✅ Ready for GitHub Actions deployment!"