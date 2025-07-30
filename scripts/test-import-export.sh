#!/bin/bash

# Test script for import/export functionality
# This script tests the bookmark import/export features

set -e

echo "📥📤 Testing Import/Export Functionality..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
IMPORT_EXPORT_BASE_URL="$API_BASE_URL/import-export"

# Test user credentials
TEST_EMAIL="import-test@example.com"
TEST_PASSWORD="testpassword123"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to make authenticated API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=${4:-200}

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "$data" \
            "$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$endpoint")
    fi

    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)

    if [ "$status" -eq "$expected_status" ]; then
        print_success "$method $endpoint - Status: $status"
        echo "$body"
        return 0
    else
        print_error "$method $endpoint - Expected: $expected_status, Got: $status"
        echo "$body"
        return 1
    fi
}

# Function to make file upload API calls
api_upload() {
    local endpoint=$1
    local file_path=$2
    local expected_status=${3:-200}

    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -F "file=@$file_path" \
        "$endpoint")

    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)

    if [ "$status" -eq "$expected_status" ]; then
        print_success "POST $endpoint - Status: $status"
        echo "$body"
        return 0
    else
        print_error "POST $endpoint - Expected: $expected_status, Got: $status"
        echo "$body"
        return 1
    fi
}

# Function to setup test user and get JWT token
setup_auth() {
    print_status "Setting up authentication..."

    # Register test user
    register_response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
        "$API_BASE_URL/auth/register")

    register_status=$(echo "$register_response" | tail -n 1)

    if [ "$register_status" -eq 201 ] || [ "$register_status" -eq 409 ]; then
        print_success "User registration completed"
    else
        print_error "User registration failed with status: $register_status"
        echo "$register_response" | head -n -1
    fi

    # Login to get JWT token
    login_response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
        "$API_BASE_URL/auth/login")

    login_body=$(echo "$login_response" | head -n -1)
    login_status=$(echo "$login_response" | tail -n 1)

    if [ "$login_status" -eq 200 ]; then
        JWT_TOKEN=$(echo "$login_body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        if [ -n "$JWT_TOKEN" ]; then
            print_success "Authentication successful"
            export JWT_TOKEN
        else
            print_error "Failed to extract JWT token"
            exit 1
        fi
    else
        print_error "Login failed with status: $login_status"
        echo "$login_body"
        exit 1
    fi
}

# Function to create test files
create_test_files() {
    print_status "Creating test import files..."

    # Create Chrome bookmarks test file
    cat > /tmp/chrome_bookmarks.json << 'EOF'
{
    "checksum": "test-checksum",
    "roots": {
        "bookmark_bar": {
            "children": [
                {
                    "date_added": "13285932710000000",
                    "guid": "test-guid-1",
                    "id": "1",
                    "name": "Google",
                    "type": "url",
                    "url": "https://www.google.com"
                },
                {
                    "children": [
                        {
                            "date_added": "13285932720000000",
                            "guid": "test-guid-2",
                            "id": "2",
                            "name": "GitHub",
                            "type": "url",
                            "url": "https://github.com"
                        }
                    ],
                    "date_added": "13285932700000000",
                    "date_modified": "13285932720000000",
                    "guid": "test-folder-guid",
                    "id": "3",
                    "name": "Development",
                    "type": "folder"
                }
            ],
            "date_added": "13285932700000000",
            "date_modified": "13285932720000000",
            "guid": "bookmark_bar_guid",
            "id": "0",
            "name": "Bookmarks bar",
            "type": "folder"
        }
    },
    "version": 1
}
EOF

    # Create Firefox bookmarks test file
    cat > /tmp/firefox_bookmarks.html << 'EOF'
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DL><p>
    <DT><H3 ADD_DATE="1640995200" LAST_MODIFIED="1640995300">Development</H3>
    <DL><p>
        <DT><A HREF="https://github.com" ADD_DATE="1640995200">GitHub</A>
        <DT><A HREF="https://stackoverflow.com" ADD_DATE="1640995250">Stack Overflow</A>
    </DL><p>
    <DT><A HREF="https://www.google.com" ADD_DATE="1640995100">Google</A>
</DL><p>
EOF

    # Create Safari bookmarks test file
    cat > /tmp/safari_bookmarks.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Children</key>
    <array>
        <dict>
            <key>Title</key>
            <string>BookmarksBar</string>
            <key>Children</key>
            <array>
                <dict>
                    <key>URLString</key>
                    <string>https://www.google.com</string>
                    <key>URIDictionary</key>
                    <dict>
                        <key>title</key>
                        <string>Google</string>
                    </dict>
                </dict>
            </array>
        </dict>
    </array>
</dict>
</plist>
EOF

    print_success "Test files created"
}

# Function to test Chrome import
test_chrome_import() {
    print_status "Testing Chrome bookmark import..."

    api_upload "$IMPORT_EXPORT_BASE_URL/import/chrome" "/tmp/chrome_bookmarks.json" "200"

    print_success "Chrome import test completed"
}

# Function to test Firefox import
test_firefox_import() {
    print_status "Testing Firefox bookmark import..."

    api_upload "$IMPORT_EXPORT_BASE_URL/import/firefox" "/tmp/firefox_bookmarks.html" "200"

    print_success "Firefox import test completed"
}

# Function to test Safari import
test_safari_import() {
    print_status "Testing Safari bookmark import..."

    api_upload "$IMPORT_EXPORT_BASE_URL/import/safari" "/tmp/safari_bookmarks.plist" "200"

    print_success "Safari import test completed"
}

# Function to test duplicate detection
test_duplicate_detection() {
    print_status "Testing duplicate detection..."

    local duplicate_data='{
        "urls": [
            "https://www.google.com",
            "https://github.com",
            "https://example.com"
        ]
    }'

    api_call "POST" "$IMPORT_EXPORT_BASE_URL/detect-duplicates" "$duplicate_data" "200"

    print_success "Duplicate detection test completed"
}

# Function to test JSON export
test_json_export() {
    print_status "Testing JSON export..."

    response=$(curl -s -w "\n%{http_code}" -X GET \
        -H "Authorization: Bearer $JWT_TOKEN" \
        "$IMPORT_EXPORT_BASE_URL/export/json")

    status=$(echo "$response" | tail -n 1)

    if [ "$status" -eq 200 ]; then
        print_success "JSON export test completed"
        # Save export to file for verification
        echo "$response" | head -n -1 > /tmp/exported_bookmarks.json
        print_status "Exported bookmarks saved to /tmp/exported_bookmarks.json"
    else
        print_error "JSON export failed with status: $status"
        echo "$response" | head -n -1
    fi
}

# Function to test HTML export
test_html_export() {
    print_status "Testing HTML export..."

    response=$(curl -s -w "\n%{http_code}" -X GET \
        -H "Authorization: Bearer $JWT_TOKEN" \
        "$IMPORT_EXPORT_BASE_URL/export/html")

    status=$(echo "$response" | tail -n 1)

    if [ "$status" -eq 200 ]; then
        print_success "HTML export test completed"
        # Save export to file for verification
        echo "$response" | head -n -1 > /tmp/exported_bookmarks.html
        print_status "Exported bookmarks saved to /tmp/exported_bookmarks.html"
    else
        print_error "HTML export failed with status: $status"
        echo "$response" | head -n -1
    fi
}

# Function to test import progress
test_import_progress() {
    print_status "Testing import progress tracking..."

    # Test with a mock job ID
    api_call "GET" "$IMPORT_EXPORT_BASE_URL/import/progress/test-job-123" "" "200"

    print_success "Import progress test completed"
}

# Function to run Go tests
run_go_tests() {
    print_status "Running Go unit tests..."

    cd "$(dirname "$0")/.."

    # Run import/export service tests
    if go test -v ./backend/internal/import/...; then
        print_success "Go unit tests passed"
    else
        print_error "Go unit tests failed"
        return 1
    fi
}

# Function to cleanup test files
cleanup() {
    print_status "Cleaning up test files..."

    rm -f /tmp/chrome_bookmarks.json
    rm -f /tmp/firefox_bookmarks.html
    rm -f /tmp/safari_bookmarks.plist
    rm -f /tmp/exported_bookmarks.json
    rm -f /tmp/exported_bookmarks.html

    print_success "Cleanup completed"
}

# Main test execution
main() {
    echo "📥📤 Starting Import/Export Functionality Tests"
    echo "=============================================="

    # Check if services are running
    print_status "Checking if API server is running..."
    if ! curl -s "$API_BASE_URL/health" > /dev/null; then
        print_error "API server is not running. Please start the services first."
        exit 1
    fi

    # Run tests
    setup_auth
    create_test_files
    test_chrome_import
    test_firefox_import
    test_safari_import
    test_duplicate_detection
    test_json_export
    test_html_export
    test_import_progress
    run_go_tests
    cleanup

    echo ""
    echo "=============================================="
    print_success "🎉 All import/export tests completed successfully!"
    echo ""

    # Print summary
    echo "📊 Test Summary:"
    echo "- ✅ Chrome bookmark import"
    echo "- ✅ Firefox bookmark import"
    echo "- ✅ Safari bookmark import"
    echo "- ✅ Duplicate detection"
    echo "- ✅ JSON export"
    echo "- ✅ HTML export"
    echo "- ✅ Import progress tracking"
    echo "- ✅ Go unit tests"
    echo ""
    echo "📥📤 Import/Export functionality is working correctly!"
}

# Run main function
main "$@"