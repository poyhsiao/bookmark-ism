#!/bin/bash

# Test script for Chrome extension
set -e

echo "🧪 Testing Chrome Extension..."

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

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js to run extension tests."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed. Please install npm to run extension tests."
    exit 1
fi

# Navigate to extensions directory
cd extensions

# Check if package.json exists, if not create it
if [ ! -f "package.json" ]; then
    print_status "Creating package.json for extension tests..."
    cat > package.json << EOF
{
  "name": "bookmark-sync-extension-tests",
  "version": "1.0.0",
  "description": "Tests for Bookmark Sync Browser Extension",
  "type": "module",
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "lint": "eslint . --ext .js",
    "lint:fix": "eslint . --ext .js --fix"
  },
  "devDependencies": {
    "vitest": "^1.0.0",
    "jsdom": "^23.0.0",
    "eslint": "^8.0.0",
    "@vitest/coverage-v8": "^1.0.0"
  }
}
EOF
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    print_status "Installing test dependencies..."
    npm install
fi

# Create vitest config if it doesn't exist
if [ ! -f "vitest.config.js" ]; then
    print_status "Creating Vitest configuration..."
    cat > vitest.config.js << EOF
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.js']
  }
});
EOF
fi

# Create test setup file
mkdir -p tests
if [ ! -f "tests/setup.js" ]; then
    print_status "Creating test setup file..."
    cat > tests/setup.js << EOF
// Test setup for browser extension
import { vi } from 'vitest';

// Mock browser APIs that are not available in test environment
global.chrome = {
  runtime: {
    sendMessage: vi.fn(),
    onMessage: { addListener: vi.fn() },
    onInstalled: { addListener: vi.fn() },
    onStartup: { addListener: vi.fn() },
    lastError: null
  },
  storage: {
    local: {
      get: vi.fn(),
      set: vi.fn(),
      remove: vi.fn(),
      clear: vi.fn(),
      getBytesInUse: vi.fn(),
      QUOTA_BYTES: 10485760
    },
    sync: {
      get: vi.fn(),
      set: vi.fn(),
      clear: vi.fn()
    }
  },
  tabs: {
    query: vi.fn(),
    create: vi.fn(),
    onUpdated: { addListener: vi.fn() }
  },
  contextMenus: {
    create: vi.fn(),
    onClicked: { addListener: vi.fn() }
  },
  notifications: {
    create: vi.fn()
  },
  identity: {
    getRedirectURL: vi.fn(),
    launchWebAuthFlow: vi.fn()
  },
  action: {
    openPopup: vi.fn()
  }
};

// Mock WebSocket
global.WebSocket = vi.fn();

// Mock fetch
global.fetch = vi.fn();

// Mock URL constructor
global.URL = vi.fn().mockImplementation((url, base) => ({
  href: base ? new URL(url, base).href : url,
  origin: 'https://example.com'
}));
EOF
fi

# Run the tests
print_status "Running extension tests..."

# Check if backend is running for integration tests
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Backend is running - running full test suite including integration tests"
    npm test
else
    print_warning "Backend is not running - running unit tests only"
    npm test -- --exclude="**/integration/**"
fi

# Check test results
if [ $? -eq 0 ]; then
    print_success "All extension tests passed! ✅"
else
    print_error "Some extension tests failed! ❌"
    exit 1
fi

# Run linting if eslint is available
if command -v npx &> /dev/null && [ -f "node_modules/.bin/eslint" ]; then
    print_status "Running ESLint..."
    npx eslint . --ext .js || print_warning "Linting issues found"
fi

# Generate coverage report if requested
if [ "$1" = "--coverage" ]; then
    print_status "Generating coverage report..."
    npm run test:coverage
fi

print_success "Chrome extension testing completed!"

# Instructions for manual testing
echo ""
echo "📋 Manual Testing Instructions:"
echo "1. Load the extension in Chrome:"
echo "   - Open Chrome and go to chrome://extensions/"
echo "   - Enable 'Developer mode'"
echo "   - Click 'Load unpacked' and select the 'extensions/chrome' directory"
echo ""
echo "2. Test the extension:"
echo "   - Click the extension icon in the toolbar"
echo "   - Try logging in with your credentials"
echo "   - Bookmark the current page"
echo "   - Test the sync functionality"
echo ""
echo "3. Check the console for any errors:"
echo "   - Right-click the extension popup and select 'Inspect'"
echo "   - Check the Console tab for any JavaScript errors"
echo ""
echo "4. Test the options page:"
echo "   - Right-click the extension icon and select 'Options'"
echo "   - Verify all settings work correctly"