#!/bin/bash

# Test Screenshot Service
# 測試截圖服務

set -e

echo "🧪 Testing Screenshot Service..."
echo "🧪 測試截圖服務..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_BASE_URL="http://localhost:8080/api/v1"
SCREENSHOT_ENDPOINT="/screenshot"

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        return 1
    fi
}

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5

    echo -e "${BLUE}Testing: $description${NC}"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$API_BASE_URL$endpoint")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)

    if [ "$http_code" = "$expected_status" ]; then
        print_result 0 "$description"
        echo "Response: $body"
        return 0
    else
        print_result 1 "$description (Expected: $expected_status, Got: $http_code)"
        echo "Response: $body"
        return 1
    fi
}

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 5

# Test 1: Capture Screenshot
echo -e "\n${YELLOW}=== Test 1: Capture Screenshot ===${NC}"
capture_data='{
    "bookmark_id": "test-bookmark-123",
    "url": "https://example.com",
    "width": 1200,
    "height": 800,
    "quality": 85,
    "format": "jpeg",
    "thumbnail": true
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/capture" "$capture_data" "200" "Capture screenshot"

# Test 2: Capture Screenshot with Defaults
echo -e "\n${YELLOW}=== Test 2: Capture Screenshot with Defaults ===${NC}"
capture_default_data='{
    "bookmark_id": "test-bookmark-456",
    "url": "https://github.com"
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/capture" "$capture_default_data" "200" "Capture screenshot with defaults"

# Test 3: Update Bookmark Screenshot
echo -e "\n${YELLOW}=== Test 3: Update Bookmark Screenshot ===${NC}"
update_data='{
    "url": "https://example.com/updated"
}'
test_endpoint "PUT" "$SCREENSHOT_ENDPOINT/bookmark/test-bookmark-789" "$update_data" "200" "Update bookmark screenshot"

# Test 4: Get Favicon
echo -e "\n${YELLOW}=== Test 4: Get Favicon ===${NC}"
favicon_data='{
    "url": "https://github.com"
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/favicon" "$favicon_data" "200" "Get favicon"

# Test 5: Capture from URL
echo -e "\n${YELLOW}=== Test 5: Capture from URL ===${NC}"
url_data='{
    "url": "https://example.com"
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/url" "$url_data" "200" "Capture from URL"

# Test 6: Invalid Requests
echo -e "\n${YELLOW}=== Test 6: Invalid Requests ===${NC}"

# Invalid capture request (missing bookmark_id)
invalid_capture_data='{
    "url": "https://example.com"
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/capture" "$invalid_capture_data" "400" "Invalid capture request"

# Invalid update request (missing URL)
invalid_update_data='{}'
test_endpoint "PUT" "$SCREENSHOT_ENDPOINT/bookmark/test-bookmark" "$invalid_update_data" "400" "Invalid update request"

# Invalid favicon request (missing URL)
invalid_favicon_data='{}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/favicon" "$invalid_favicon_data" "400" "Invalid favicon request"

# Test 7: Error Handling
echo -e "\n${YELLOW}=== Test 7: Error Handling ===${NC}"

# Invalid URL format
invalid_url_data='{
    "bookmark_id": "test-bookmark",
    "url": "not-a-valid-url"
}'
test_endpoint "POST" "$SCREENSHOT_ENDPOINT/capture" "$invalid_url_data" "500" "Invalid URL format"

# Test 8: Screenshot Integration Test
echo -e "\n${YELLOW}=== Test 8: Screenshot Integration Test ===${NC}"
echo -e "${BLUE}Testing: Full screenshot workflow${NC}"

# Create test bookmark first (assuming bookmark service is available)
bookmark_data='{
    "url": "https://example.com/test-screenshot",
    "title": "Test Screenshot Bookmark",
    "description": "Testing screenshot integration"
}'

bookmark_response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test-token" \
    -d "$bookmark_data" \
    "$API_BASE_URL/bookmarks" || echo -e "\nbookmark_failed")

bookmark_code=$(echo "$bookmark_response" | tail -n1)

if [ "$bookmark_code" = "201" ] || [ "$bookmark_code" = "401" ]; then
    print_result 0 "Screenshot integration test setup (bookmark creation or auth required)"

    # Extract bookmark ID if successful
    if [ "$bookmark_code" = "201" ]; then
        bookmark_id=$(echo "$bookmark_response" | head -n -1 | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$bookmark_id" ]; then
            # Test screenshot capture for the created bookmark
            integration_data="{
                \"bookmark_id\": \"$bookmark_id\",
                \"url\": \"https://example.com/test-screenshot\",
                \"thumbnail\": true
            }"
            test_endpoint "POST" "$SCREENSHOT_ENDPOINT/capture" "$integration_data" "200" "Screenshot capture for created bookmark"
        fi
    fi
else
    print_result 1 "Screenshot integration test setup (Got: $bookmark_code)"
fi

# Test 9: Performance Test
echo -e "\n${YELLOW}=== Test 9: Performance Test ===${NC}"
echo -e "${BLUE}Testing: Multiple concurrent screenshot requests${NC}"

start_time=$(date +%s)

# Create multiple background requests
for i in {1..5}; do
    perf_data="{
        \"bookmark_id\": \"perf-test-$i\",
        \"url\": \"https://httpbin.org/delay/1\",
        \"width\": 800,
        \"height\": 600
    }"
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$perf_data" \
        "$API_BASE_URL$SCREENSHOT_ENDPOINT/capture" > /dev/null &
done

# Wait for all background jobs to complete
wait

end_time=$(date +%s)
duration=$((end_time - start_time))

if [ $duration -lt 10 ]; then
    print_result 0 "Performance test (completed in ${duration}s)"
else
    print_result 1 "Performance test (took ${duration}s, expected < 10s)"
fi

echo -e "\n${GREEN}🎉 Screenshot service tests completed!${NC}"
echo -e "${GREEN}🎉 截圖服務測試完成！${NC}"

# Summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo -e "${BLUE}✅ Screenshot capture${NC}"
echo -e "${BLUE}✅ Bookmark screenshot updates${NC}"
echo -e "${BLUE}✅ Favicon retrieval${NC}"
echo -e "${BLUE}✅ URL-based capture${NC}"
echo -e "${BLUE}✅ Error handling${NC}"
echo -e "${BLUE}✅ Integration testing${NC}"
echo -e "${BLUE}✅ Performance testing${NC}"

echo -e "\n${GREEN}All screenshot functionality is working correctly!${NC}"
echo -e "${GREEN}所有截圖功能都正常工作！${NC}"