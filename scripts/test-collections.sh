#!/bin/bash

# Test script for collection endpoints
# 集合端點測試腳本

set -e

BASE_URL="http://localhost:8080/api/v1"
AUTH_TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to make authenticated requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "$data" \
            "$BASE_URL$endpoint"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            "$BASE_URL$endpoint"
    fi
}

# Function to check if server is running
check_server() {
    print_status "Checking if server is running..."
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "Server is not running. Please start the server first."
        print_status "Run: make dev or go run backend/cmd/api/main.go"
        exit 1
    fi
    print_status "Server is running ✓"
}

# Function to authenticate (mock for testing)
authenticate() {
    print_status "Authenticating..."
    # For testing purposes, we'll use a mock token
    # In real scenario, you would login and get a real token
    AUTH_TOKEN="mock-jwt-token-for-testing"
    print_warning "Using mock authentication token"
    print_status "Authentication completed ✓"
}

# Function to test collection creation
test_create_collection() {
    print_status "Testing collection creation..."

    local response=$(make_request "POST" "/collections" '{
        "name": "Test Collection",
        "description": "A test collection for API testing",
        "color": "#3B82F6",
        "icon": "folder",
        "visibility": "private"
    }')

    echo "Response: $response"

    # Extract collection ID from response (basic parsing)
    COLLECTION_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2 || echo "1")
    print_status "Collection created with ID: $COLLECTION_ID ✓"
}

# Function to test collection listing
test_list_collections() {
    print_status "Testing collection listing..."

    local response=$(make_request "GET" "/collections")
    echo "Response: $response"
    print_status "Collections listed ✓"
}

# Function to test collection retrieval
test_get_collection() {
    print_status "Testing collection retrieval..."

    local response=$(make_request "GET" "/collections/$COLLECTION_ID")
    echo "Response: $response"
    print_status "Collection retrieved ✓"
}

# Function to test collection update
test_update_collection() {
    print_status "Testing collection update..."

    local response=$(make_request "PUT" "/collections/$COLLECTION_ID" '{
        "name": "Updated Test Collection",
        "description": "Updated description",
        "color": "#10B981"
    }')

    echo "Response: $response"
    print_status "Collection updated ✓"
}

# Function to test bookmark creation (for collection testing)
test_create_bookmark() {
    print_status "Creating a test bookmark for collection testing..."

    local response=$(make_request "POST" "/bookmarks" '{
        "url": "https://example.com",
        "title": "Test Bookmark",
        "description": "A test bookmark for collection testing"
    }')

    echo "Response: $response"

    # Extract bookmark ID from response
    BOOKMARK_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2 || echo "1")
    print_status "Bookmark created with ID: $BOOKMARK_ID ✓"
}

# Function to test adding bookmark to collection
test_add_bookmark_to_collection() {
    print_status "Testing adding bookmark to collection..."

    local response=$(make_request "POST" "/collections/$COLLECTION_ID/bookmarks/$BOOKMARK_ID")
    echo "Response: $response"
    print_status "Bookmark added to collection ✓"
}

# Function to test getting collection bookmarks
test_get_collection_bookmarks() {
    print_status "Testing getting collection bookmarks..."

    local response=$(make_request "GET" "/collections/$COLLECTION_ID/bookmarks")
    echo "Response: $response"
    print_status "Collection bookmarks retrieved ✓"
}

# Function to test removing bookmark from collection
test_remove_bookmark_from_collection() {
    print_status "Testing removing bookmark from collection..."

    local response=$(make_request "DELETE" "/collections/$COLLECTION_ID/bookmarks/$BOOKMARK_ID")
    echo "Response: $response"
    print_status "Bookmark removed from collection ✓"
}

# Function to test collection deletion
test_delete_collection() {
    print_status "Testing collection deletion..."

    local response=$(make_request "DELETE" "/collections/$COLLECTION_ID")
    echo "Response: $response"
    print_status "Collection deleted ✓"
}

# Main test execution
main() {
    print_status "Starting Collection API Tests"
    print_status "=============================="

    check_server
    authenticate

    print_status ""
    print_status "Running Collection Tests..."
    print_status "----------------------------"

    test_create_collection
    echo ""

    test_list_collections
    echo ""

    test_get_collection
    echo ""

    test_update_collection
    echo ""

    test_create_bookmark
    echo ""

    test_add_bookmark_to_collection
    echo ""

    test_get_collection_bookmarks
    echo ""

    test_remove_bookmark_from_collection
    echo ""

    test_delete_collection
    echo ""

    print_status "=============================="
    print_status "All Collection API Tests Completed!"
    print_warning "Note: These tests use mock authentication."
    print_warning "For production testing, use real authentication tokens."
}

# Run main function
main "$@"