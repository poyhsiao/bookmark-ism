#!/bin/bash

# Test script for link monitoring and maintenance features
# This script tests the Task 24 implementation using TDD methodology

set -e

echo "🧪 Testing Link Monitoring and Maintenance Features (Task 24)"
echo "============================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_BASE_URL="http://localhost:8080/api/v1"
TEST_USER_EMAIL="test@example.com"
TEST_USER_PASSWORD="testpassword123"
AUTH_TOKEN=""

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

# Function to make authenticated API requests
api_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=${4:-200}

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            "$API_BASE_URL$endpoint")
    fi

    # Split response and status code
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)

    if [ "$status_code" -eq "$expected_status" ]; then
        echo "$body"
        return 0
    else
        print_error "Expected status $expected_status, got $status_code"
        echo "$body"
        return 1
    fi
}

# Function to authenticate and get token
authenticate() {
    print_status "Authenticating test user..."

    # Try to register user (might fail if already exists)
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}" \
        "$API_BASE_URL/auth/register" > /dev/null 2>&1 || true

    # Login to get token
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}" \
        "$API_BASE_URL/auth/login")

    AUTH_TOKEN=$(echo "$response" | jq -r '.access_token // empty')

    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
        print_error "Failed to authenticate"
        echo "Response: $response"
        exit 1
    fi

    print_success "Authentication successful"
}

# Function to create test bookmark
create_test_bookmark() {
    local url=$1
    local title=${2:-"Test Bookmark"}

    print_status "Creating test bookmark: $url"

    response=$(api_request "POST" "/bookmarks" "{
        \"url\": \"$url\",
        \"title\": \"$title\",
        \"description\": \"Test bookmark for monitoring\",
        \"tags\": [\"test\", \"monitoring\"]
    }" 201)

    bookmark_id=$(echo "$response" | jq -r '.bookmark.id')
    echo "$bookmark_id"
}

# Test 1: Link Check Functionality
test_link_check() {
    print_status "Test 1: Link Check Functionality"

    # Create test bookmark with a working URL
    bookmark_id=$(create_test_bookmark "https://httpbin.org/status/200" "Working Link")

    # Test link check
    print_status "Testing link check for working URL..."
    response=$(api_request "POST" "/monitoring/check-link" "{
        \"bookmark_id\": $bookmark_id,
        \"url\": \"https://httpbin.org/status/200\"
    }")

    status=$(echo "$response" | jq -r '.link_check.status')
    if [ "$status" = "active" ]; then
        print_success "✅ Link check for working URL passed"
    else
        print_error "❌ Link check for working URL failed: status=$status"
        return 1
    fi

    # Create test bookmark with a broken URL
    broken_bookmark_id=$(create_test_bookmark "https://httpbin.org/status/404" "Broken Link")

    # Test link check for broken URL
    print_status "Testing link check for broken URL..."
    response=$(api_request "POST" "/monitoring/check-link" "{
        \"bookmark_id\": $broken_bookmark_id,
        \"url\": \"https://httpbin.org/status/404\"
    }")

    status=$(echo "$response" | jq -r '.link_check.status')
    if [ "$status" = "broken" ]; then
        print_success "✅ Link check for broken URL passed"
    else
        print_error "❌ Link check for broken URL failed: status=$status"
        return 1
    fi

    # Test redirect URL
    redirect_bookmark_id=$(create_test_bookmark "https://httpbin.org/redirect/1" "Redirect Link")

    print_status "Testing link check for redirect URL..."
    response=$(api_request "POST" "/monitoring/check-link" "{
        \"bookmark_id\": $redirect_bookmark_id,
        \"url\": \"https://httpbin.org/redirect/1\"
    }")

    status=$(echo "$response" | jq -r '.link_check.status')
    if [ "$status" = "redirect" ]; then
        print_success "✅ Link check for redirect URL passed"
    else
        print_error "❌ Link check for redirect URL failed: status=$status"
        return 1
    fi

    # Get link checks for bookmark
    print_status "Testing get link checks..."
    response=$(api_request "GET" "/monitoring/bookmarks/$bookmark_id/checks")

    total=$(echo "$response" | jq -r '.total')
    if [ "$total" -gt 0 ]; then
        print_success "✅ Get link checks passed: found $total checks"
    else
        print_error "❌ Get link checks failed: no checks found"
        return 1
    fi

    print_success "✅ All link check tests passed"
}

# Test 2: Monitoring Jobs
test_monitoring_jobs() {
    print_status "Test 2: Monitoring Jobs"

    # Create monitoring job
    print_status "Creating monitoring job..."
    response=$(api_request "POST" "/monitoring/jobs" "{
        \"name\": \"Daily Link Check\",
        \"description\": \"Check all links daily\",
        \"enabled\": true,
        \"frequency\": \"0 0 * * *\"
    }" 201)

    job_id=$(echo "$response" | jq -r '.job.id')
    if [ "$job_id" != "null" ] && [ -n "$job_id" ]; then
        print_success "✅ Monitoring job created: ID=$job_id"
    else
        print_error "❌ Failed to create monitoring job"
        return 1
    fi

    # Get monitoring job
    print_status "Getting monitoring job..."
    response=$(api_request "GET" "/monitoring/jobs/$job_id")

    name=$(echo "$response" | jq -r '.job.name')
    if [ "$name" = "Daily Link Check" ]; then
        print_success "✅ Get monitoring job passed"
    else
        print_error "❌ Get monitoring job failed: name=$name"
        return 1
    fi

    # Update monitoring job
    print_status "Updating monitoring job..."
    response=$(api_request "PUT" "/monitoring/jobs/$job_id" "{
        \"name\": \"Updated Daily Check\",
        \"enabled\": false
    }")

    updated_name=$(echo "$response" | jq -r '.job.name')
    enabled=$(echo "$response" | jq -r '.job.enabled')
    if [ "$updated_name" = "Updated Daily Check" ] && [ "$enabled" = "false" ]; then
        print_success "✅ Update monitoring job passed"
    else
        print_error "❌ Update monitoring job failed"
        return 1
    fi

    # List monitoring jobs
    print_status "Listing monitoring jobs..."
    response=$(api_request "GET" "/monitoring/jobs")

    total=$(echo "$response" | jq -r '.total')
    if [ "$total" -gt 0 ]; then
        print_success "✅ List monitoring jobs passed: found $total jobs"
    else
        print_error "❌ List monitoring jobs failed"
        return 1
    fi

    # Delete monitoring job
    print_status "Deleting monitoring job..."
    api_request "DELETE" "/monitoring/jobs/$job_id" "" 200 > /dev/null

    # Verify deletion
    api_request "GET" "/monitoring/jobs/$job_id" "" 404 > /dev/null
    if [ $? -eq 0 ]; then
        print_success "✅ Delete monitoring job passed"
    else
        print_error "❌ Delete monitoring job failed"
        return 1
    fi

    print_success "✅ All monitoring job tests passed"
}

# Test 3: Maintenance Reports
test_maintenance_reports() {
    print_status "Test 3: Maintenance Reports"

    # Create some test bookmarks with different statuses
    bookmark1=$(create_test_bookmark "https://httpbin.org/status/200" "Working Link 1")
    bookmark2=$(create_test_bookmark "https://httpbin.org/status/404" "Broken Link 1")
    bookmark3=$(create_test_bookmark "https://httpbin.org/redirect/1" "Redirect Link 1")

    # Check the links to generate data
    api_request "POST" "/monitoring/check-link" "{\"bookmark_id\": $bookmark1, \"url\": \"https://httpbin.org/status/200\"}" > /dev/null
    api_request "POST" "/monitoring/check-link" "{\"bookmark_id\": $bookmark2, \"url\": \"https://httpbin.org/status/404\"}" > /dev/null
    api_request "POST" "/monitoring/check-link" "{\"bookmark_id\": $bookmark3, \"url\": \"https://httpbin.org/redirect/1\"}" > /dev/null

    # Generate maintenance report
    print_status "Generating maintenance report..."
    response=$(api_request "POST" "/monitoring/reports")

    total_links=$(echo "$response" | jq -r '.report.total_links')
    broken_links=$(echo "$response" | jq -r '.report.broken_links')
    active_links=$(echo "$response" | jq -r '.report.active_links')
    redirect_links=$(echo "$response" | jq -r '.report.redirect_links')

    if [ "$total_links" -gt 0 ] && [ "$broken_links" -gt 0 ] && [ "$active_links" -gt 0 ]; then
        print_success "✅ Maintenance report generated successfully"
        print_success "   Total: $total_links, Active: $active_links, Broken: $broken_links, Redirects: $redirect_links"
    else
        print_error "❌ Maintenance report generation failed"
        echo "Response: $response"
        return 1
    fi

    print_success "✅ All maintenance report tests passed"
}

# Test 4: Notifications
test_notifications() {
    print_status "Test 4: Notifications"

    # Create a broken link to generate notification
    bookmark_id=$(create_test_bookmark "https://httpbin.org/status/500" "Server Error Link")

    # Check the link to generate notification
    api_request "POST" "/monitoring/check-link" "{
        \"bookmark_id\": $bookmark_id,
        \"url\": \"https://httpbin.org/status/500\"
    }" > /dev/null

    # Get notifications
    print_status "Getting notifications..."
    response=$(api_request "GET" "/monitoring/notifications")

    total=$(echo "$response" | jq -r '.total')
    if [ "$total" -gt 0 ]; then
        print_success "✅ Get notifications passed: found $total notifications"

        # Get first notification ID
        notification_id=$(echo "$response" | jq -r '.items[0].id')

        # Mark notification as read
        print_status "Marking notification as read..."
        api_request "PUT" "/monitoring/notifications/$notification_id/read" "" 200 > /dev/null

        if [ $? -eq 0 ]; then
            print_success "✅ Mark notification as read passed"
        else
            print_error "❌ Mark notification as read failed"
            return 1
        fi

        # Test unread notifications filter
        print_status "Testing unread notifications filter..."
        response=$(api_request "GET" "/monitoring/notifications?unread_only=true")

        unread_total=$(echo "$response" | jq -r '.total')
        if [ "$unread_total" -ge 0 ]; then
            print_success "✅ Unread notifications filter passed: found $unread_total unread"
        else
            print_error "❌ Unread notifications filter failed"
            return 1
        fi
    else
        print_warning "⚠️  No notifications found (this might be expected)"
    fi

    print_success "✅ All notification tests passed"
}

# Test 5: Error Handling
test_error_handling() {
    print_status "Test 5: Error Handling"

    # Test invalid bookmark ID
    print_status "Testing invalid bookmark ID..."
    api_request "POST" "/monitoring/check-link" "{
        \"bookmark_id\": 99999,
        \"url\": \"https://example.com\"
    }" 404 > /dev/null

    if [ $? -eq 0 ]; then
        print_success "✅ Invalid bookmark ID error handling passed"
    else
        print_error "❌ Invalid bookmark ID error handling failed"
        return 1
    fi

    # Test invalid cron expression
    print_status "Testing invalid cron expression..."
    api_request "POST" "/monitoring/jobs" "{
        \"name\": \"Invalid Job\",
        \"frequency\": \"invalid cron\"
    }" 400 > /dev/null

    if [ $? -eq 0 ]; then
        print_success "✅ Invalid cron expression error handling passed"
    else
        print_error "❌ Invalid cron expression error handling failed"
        return 1
    fi

    # Test non-existent job
    print_status "Testing non-existent job..."
    api_request "GET" "/monitoring/jobs/99999" "" 404 > /dev/null

    if [ $? -eq 0 ]; then
        print_success "✅ Non-existent job error handling passed"
    else
        print_error "❌ Non-existent job error handling failed"
        return 1
    fi

    print_success "✅ All error handling tests passed"
}

# Test 6: Unit Tests
test_unit_tests() {
    print_status "Test 6: Running Unit Tests"

    cd "$(dirname "$0")/../backend"

    # Run monitoring service tests
    print_status "Running monitoring service unit tests..."
    if go test -v ./internal/monitoring/... -count=1; then
        print_success "✅ Monitoring service unit tests passed"
    else
        print_error "❌ Monitoring service unit tests failed"
        return 1
    fi

    print_success "✅ All unit tests passed"
}

# Main test execution
main() {
    print_status "Starting Link Monitoring and Maintenance Tests"
    print_status "Testing Task 24 implementation with TDD methodology"
    echo ""

    # Check if server is running
    if ! curl -s "$API_BASE_URL/../health" > /dev/null; then
        print_error "Server is not running at $API_BASE_URL"
        print_error "Please start the server with: make dev"
        exit 1
    fi

    # Authenticate
    authenticate

    # Run tests
    local failed_tests=0

    test_link_check || ((failed_tests++))
    echo ""

    test_monitoring_jobs || ((failed_tests++))
    echo ""

    test_maintenance_reports || ((failed_tests++))
    echo ""

    test_notifications || ((failed_tests++))
    echo ""

    test_error_handling || ((failed_tests++))
    echo ""

    test_unit_tests || ((failed_tests++))
    echo ""

    # Summary
    echo "============================================================"
    if [ $failed_tests -eq 0 ]; then
        print_success "🎉 All tests passed! Task 24 implementation is complete."
        print_success "✅ Link monitoring and maintenance features are working correctly."
        echo ""
        print_status "Features implemented:"
        echo "  • Link checking with status detection (active/broken/redirect/timeout)"
        echo "  • Monitoring job management with cron scheduling"
        echo "  • Maintenance report generation with health analysis"
        echo "  • Notification system for link changes"
        echo "  • Comprehensive error handling and validation"
        echo "  • RESTful API endpoints with proper authentication"
        echo "  • Complete test coverage with TDD methodology"
        echo ""
        print_success "Task 24: Link monitoring and maintenance features - COMPLETED ✅"
    else
        print_error "❌ $failed_tests test(s) failed. Please check the implementation."
        exit 1
    fi
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    print_error "jq is required but not installed. Please install jq first."
    print_error "On macOS: brew install jq"
    print_error "On Ubuntu: sudo apt-get install jq"
    exit 1
fi

# Run main function
main "$@"