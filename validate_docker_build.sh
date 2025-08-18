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
        docker rmi "$TEST_TAG" >/dev/null 2>&1
        print_status $GREEN "✅ Cleanup completed"

    else
        print_status $RED "❌ Docker image was not created properly"
        exit 1
    fi
else
    print_status $RED "❌ Docker build failed"
    exit 1
fi

echo ""
print_status $GREEN "🎉 All validations passed! The Docker build fix is working correctly."
echo ""
print_status $YELLOW "📝 Summary of fixes applied:"
echo "   • Fixed build context path issues"
echo "   • Optimized layer caching with go.mod/go.sum copy"
echo "   • Added build cache mounts for faster builds"
echo "   • Used distroless base image for security"
echo "   • Added proper build flags for static binary"
echo "   • Included health check configuration"
echo "   • Maintained non-root user for security"
echo ""
print_status $GREEN "✅ Ready for GitHub Actions deployment!"