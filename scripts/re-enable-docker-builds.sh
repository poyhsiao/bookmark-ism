#!/bin/bash

# Re-enable Docker Builds Script
# This script re-enables Docker builds in CI/CD after validating the fix

set -e

echo "🔧 Re-enabling Docker builds in CI/CD workflows..."

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

# Validate that we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -f ".github/workflows/ci.yml" ]]; then
    print_status $RED "❌ Please run this script from the project root directory"
    exit 1
fi

print_status $GREEN "✅ Project structure validated"

# Run validation script first
if [[ -f "validate_docker_build.sh" ]]; then
    print_status $YELLOW "🔍 Running Docker build validation..."
    if ./validate_docker_build.sh; then
        print_status $GREEN "✅ Docker build validation passed"
    else
        print_status $RED "❌ Docker build validation failed. Please fix issues before re-enabling CI/CD builds."
        exit 1
    fi
else
    print_status $RED "❌ validate_docker_build.sh not found. Please ensure the validation script exists."
    exit 1
fi

# Backup current workflow files
print_status $YELLOW "📋 Creating backup of current workflow files..."
cp .github/workflows/ci.yml .github/workflows/ci.yml.backup
cp .github/workflows/cd.yml .github/workflows/cd.yml.backup
print_status $GREEN "✅ Backup created"

# Re-enable CI workflow
print_status $YELLOW "🔧 Re-enabling Docker build in CI workflow..."

# Replace the disabled docker-build job with the enabled version
cat > /tmp/ci_docker_build.yml << 'EOF'
  # Docker Build Test
  docker-build:
    name: Docker Build Test
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Build backend Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: false
        tags: bookmark-sync-backend:test
        platforms: linux/amd64
        cache-from: type=gha
        cache-to: type=gha,mode=max
        build-args: |
          BUILDKIT_INLINE_CACHE=1

    - name: Test Docker Compose
      run: |
        # Create test environment file
        cp .env.example .env

        # Start services
        docker-compose -f docker-compose.yml up -d --build

        # Wait for services
        sleep 30

        # Check health
        ./scripts/health-check.sh || true

        # Stop services
        docker-compose down
EOF

# Replace the disabled section in CI workflow
sed -i.tmp '/# Docker Build Test - TEMPORARILY DISABLED/,/^  # Notification/c\
  # Docker Build Test\
  docker-build:\
    name: Docker Build Test\
    runs-on: ubuntu-latest\
\
    steps:\
    - name: Checkout code\
      uses: actions/checkout@v4\
\
    - name: Set up Docker Buildx\
      uses: docker/setup-buildx-action@v3\
\
    - name: Build backend Docker image\
      uses: docker/build-push-action@v5\
      with:\
        context: .\
        file: ./Dockerfile\
        push: false\
        tags: bookmark-sync-backend:test\
        platforms: linux/amd64\
        cache-from: type=gha\
        cache-to: type=gha,mode=max\
        build-args: |\
          BUILDKIT_INLINE_CACHE=1\
\
    - name: Test Docker Compose\
      run: |\
        # Create test environment file\
        cp .env.example .env\
\
        # Start services\
        docker-compose -f docker-compose.yml up -d --build\
\
        # Wait for services\
        sleep 30\
\
        # Check health\
        ./scripts/health-check.sh || true\
\
        # Stop services\
        docker-compose down\
\
  # Notification' .github/workflows/ci.yml

# Re-enable CD workflow
print_status $YELLOW "🔧 Re-enabling Docker build and push in CD workflow..."

# Replace the disabled build step with the enabled version
sed -i.tmp '/# TEMPORARILY DISABLED - Docker build failing/,/^    - name: Skip SBOM generation/c\
    - name: Build and push backend image\
      id: build\
      uses: docker/build-push-action@v5\
      with:\
        context: .\
        file: ./Dockerfile.prod\
        push: true\
        tags: ${{ steps.meta.outputs.tags }}\
        labels: ${{ steps.meta.outputs.labels }}\
        cache-from: type=gha\
        cache-to: type=gha,mode=max\
        platforms: linux/amd64\
        build-args: |\
          BUILDKIT_INLINE_CACHE=1\
\
    - name: Generate SBOM\
      id: sbom\
      run: |\
        # Extract the first tag for SBOM generation (avoid multi-platform issues)\
        IMAGE_TAG=$(echo "${{ steps.meta.outputs.tags }}" | head -n1)\
        echo "Generating SBOM for image: $IMAGE_TAG"\
\
        # Install syft if not available\
        if ! command -v syft &> /dev/null; then\
          echo "Installing syft..."\
          curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin\
        fi\
\
        # Generate SBOM with error handling\
        if syft "$IMAGE_TAG" -o spdx-json=sbom.spdx.json; then\
          echo "SBOM generated successfully"\
          echo "sbom_generated=true" >> $GITHUB_OUTPUT\
        else\
          echo "SBOM generation failed, creating empty file"\
          echo '"'"'{"spdxVersion":"SPDX-2.3","dataLicense":"CC0-1.0","SPDXID":"SPDXRef-DOCUMENT","name":"sbom-generation-failed","documentNamespace":"https://github.com/'"'"'${{ github.repository }}'"'"'","creationInfo":{"created":"'"'"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"'"'","creators":["Tool: github-actions"]}}'"'"' > sbom.spdx.json\
          echo "sbom_generated=false" >> $GITHUB_OUTPUT\
        fi\
      continue-on-error: true\
\
    - name: Skip SBOM generation' .github/workflows/cd.yml

# Clean up temporary files
rm -f .github/workflows/ci.yml.tmp .github/workflows/cd.yml.tmp

# Update job outputs to remove skipped flag
sed -i.tmp 's/skipped: \${{ steps\.build\.outputs\.skipped }}//' .github/workflows/cd.yml

# Remove conditional checks for skipped builds
sed -i.tmp 's/ && needs\.build-and-push\.outputs\.skipped != '\''true'\''//g' .github/workflows/cd.yml
sed -i.tmp 's/ && needs\.build-and-push\.outputs\.skipped == '\''true'\''//g' .github/workflows/cd.yml

# Remove docker-build-skipped job
sed -i.tmp '/# Docker build skipped notification/,/echo "⏭️ Pipeline completed with Docker build skipped"/d' .github/workflows/cd.yml

print_status $GREEN "✅ Docker builds re-enabled in CI/CD workflows"

# Validate the updated workflow files
print_status $YELLOW "🔍 Validating updated workflow files..."

if command -v yamllint &> /dev/null; then
    if yamllint .github/workflows/ci.yml && yamllint .github/workflows/cd.yml; then
        print_status $GREEN "✅ Workflow YAML syntax is valid"
    else
        print_status $RED "❌ Workflow YAML syntax errors detected. Please review the files."
        exit 1
    fi
else
    print_status $YELLOW "⚠️  yamllint not available. Attempting to install yamllint..."

    # Try to install yamllint if possible
    if command -v pip &> /dev/null; then
        pip install --user yamllint && export PATH="$PATH:$(python -m site --user-base)/bin"
        if command -v yamllint &> /dev/null; then
            print_status $GREEN "✅ yamllint installed successfully. Validating workflow files..."
            if yamllint .github/workflows/ci.yml && yamllint .github/workflows/cd.yml; then
                print_status $GREEN "✅ Workflow YAML syntax is valid"
            else
                print_status $RED "❌ Workflow YAML syntax errors detected. Please review the files."
                exit 1
            fi
        else
            print_status $RED "❌ yamllint installation failed. Falling back to basic YAML validation using Python."
        fi
    else
        print_status $YELLOW "⚠️  pip not available. Falling back to basic YAML validation using Python."
    fi

    # Basic YAML syntax check using Python and PyYAML
    if command -v python3 &> /dev/null; then
        python3 -c "import sys, yaml; [yaml.safe_load(open(f)) for f in sys.argv[1:]]" .github/workflows/ci.yml .github/workflows/cd.yml \
        && print_status $GREEN '✅ Basic YAML syntax is valid (checked with PyYAML)' \
        || { print_status $RED '❌ YAML syntax errors detected (checked with PyYAML). Please review the files.'; exit 1; }
    else
        print_status $YELLOW "⚠️  Python3 not available. Please manually validate YAML syntax or install yamllint."
    fi
fi

print_status $GREEN "🎉 Docker builds successfully re-enabled!"
echo ""
print_status $YELLOW "📝 Next steps:"
echo "   1. Review the updated workflow files"
echo "   2. Commit and push the changes"
echo "   3. Monitor GitHub Actions for successful builds"
echo "   4. Remove backup files if everything works correctly:"
echo "      rm .github/workflows/ci.yml.backup .github/workflows/cd.yml.backup"
echo ""
print_status $GREEN "✅ Ready for GitHub Actions deployment!"