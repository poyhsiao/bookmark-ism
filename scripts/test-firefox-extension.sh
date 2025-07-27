#!/bin/bash

# Test script for Firefox extension
set -e

echo "🦊 Testing Firefox Extension..."

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

# Check if web-ext is installed
if ! command -v web-ext &> /dev/null; then
    print_warning "web-ext is not installed. Installing globally..."
    npm install -g web-ext
fi

# Navigate to Firefox extension directory
cd extensions/firefox

# Validate the extension
print_status "Validating Firefox extension..."
web-ext lint

if [ $? -eq 0 ]; then
    print_success "Firefox extension validation passed! ✅"
else
    print_error "Firefox extension validation failed! ❌"
    exit 1
fi

# Build the extension
print_status "Building Firefox extension..."
web-ext build --overwrite-dest

if [ $? -eq 0 ]; then
    print_success "Firefox extension built successfully! ✅"
else
    print_error "Firefox extension build failed! ❌"
    exit 1
fi

# Check if backend is running for integration tests
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Backend is running - extension can be tested with full functionality"
else
    print_warning "Backend is not running - extension will have limited functionality"
fi

print_success "Firefox extension testing completed!"

# Instructions for manual testing
echo ""
echo "📋 Manual Testing Instructions:"
echo "1. Load the extension in Firefox:"
echo "   - Open Firefox and go to about:debugging"
echo "   - Click 'This Firefox'"
echo "   - Click 'Load Temporary Add-on'"
echo "   - Select the 'extensions/firefox/manifest.json' file"
echo ""
echo "2. Alternative: Use web-ext run for automatic loading:"
echo "   - Run: cd extensions/firefox && web-ext run"
echo "   - This will open a new Firefox instance with the extension loaded"
echo ""
echo "3. Test the extension:"
echo "   - Click the extension icon in the toolbar"
echo "   - Try logging in with your credentials"
echo "   - Bookmark the current page"
echo "   - Test the sync functionality"
echo ""
echo "4. Check the console for any errors:"
echo "   - Open Firefox Developer Tools (F12)"
echo "   - Check the Console tab for any JavaScript errors"
echo ""
echo "5. Test the options page:"
echo "   - Go to about:addons"
echo "   - Find the Bookmark Sync extension"
echo "   - Click 'Options' or 'Preferences'"
echo ""
echo "6. Test cross-browser sync:"
echo "   - Install both Chrome and Firefox extensions"
echo "   - Login with the same account on both"
echo "   - Create bookmarks in one browser"
echo "   - Verify they appear in the other browser"

# Run web-ext if requested
if [ "$1" = "--run" ]; then
    print_status "Starting Firefox with extension loaded..."
    web-ext run --start-url="http://localhost:8080" --browser-console
fi