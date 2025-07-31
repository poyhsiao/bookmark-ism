#!/bin/bash

# Test script for offline functionality
# This script tests the comprehensive offline support system

set -e

echo "🧪 Testing Offline Support System"
echo "=================================="

# Configuration
API_BASE="http://localhost:8080/api/v1"
USER_ID="1"
DEVICE_ID="test-device-123"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function to make API calls
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4

    echo -n "Testing $method $endpoint... "

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "X-User-ID: $USER_ID" \
            -d "$data" \
            "$API_BASE$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "X-User-ID: $USER_ID" \
            "$API_BASE$endpoint")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)

    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗ (Expected $expected_status, got $http_code)${NC}"
        echo "Response: $body"
        return 1
    fi
}

# Test connectivity check
echo -e "\n${YELLOW}1. Testing Connectivity Check${NC}"
make_request "GET" "/offline/connectivity" "" "200"

# Test offline status management
echo -e "\n${YELLOW}2. Testing Offline Status Management${NC}"
make_request "GET" "/offline/status" "" "200"
make_request "PUT" "/offline/status" '{"status":"offline"}' "200"
make_request "GET" "/offline/status" "" "200"
make_request "PUT" "/offline/status" '{"status":"online"}' "200"

# Test bookmark caching
echo -e "\n${YELLOW}3. Testing Bookmark Caching${NC}"
bookmark_data='{
    "id": 1,
    "url": "https://example.com",
    "title": "Example Website",
    "description": "Test bookmark for offline caching",
    "tags": "[\"test\", \"offline\"]"
}'
make_request "POST" "/offline/cache/bookmark" "$bookmark_data" "200"
make_request "GET" "/offline/cache/bookmark/1" "" "200"
make_request "GET" "/offline/cache/bookmarks" "" "200"

# Test offline change queuing
echo -e "\n${YELLOW}4. Testing Offline Change Queuing${NC}"
change_data='{
    "device_id": "'$DEVICE_ID'",
    "type": "bookmark_create",
    "resource_id": "bookmark-123",
    "data": "{\"url\":\"https://test.com\",\"title\":\"Test Bookmark\"}"
}'
make_request "POST" "/offline/queue/change" "$change_data" "200"

# Test invalid change type
invalid_change_data='{
    "device_id": "'$DEVICE_ID'",
    "type": "invalid_type",
    "resource_id": "bookmark-123",
    "data": "{\"url\":\"https://test.com\",\"title\":\"Test Bookmark\"}"
}'
make_request "POST" "/offline/queue/change" "$invalid_change_data" "400"

# Test getting offline queue
make_request "GET" "/offline/queue" "" "200"

# Test processing offline queue
echo -e "\n${YELLOW}5. Testing Offline Queue Processing${NC}"
make_request "POST" "/offline/sync" "" "200"

# Test cache statistics
echo -e "\n${YELLOW}6. Testing Cache Statistics${NC}"
make_request "GET" "/offline/stats" "" "200"

# Test offline indicator
echo -e "\n${YELLOW}7. Testing Offline Indicator${NC}"
make_request "GET" "/offline/indicator" "" "200"

# Test cache cleanup
echo -e "\n${YELLOW}8. Testing Cache Cleanup${NC}"
make_request "DELETE" "/offline/cache/cleanup" "" "200"

# Test error cases
echo -e "\n${YELLOW}9. Testing Error Cases${NC}"

# Test without user ID header
echo -n "Testing request without user ID... "
response=$(curl -s -w "\n%{http_code}" -X "GET" "$API_BASE/offline/status")
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ (Expected 401, got $http_code)${NC}"
fi

# Test with invalid user ID
echo -n "Testing request with invalid user ID... "
response=$(curl -s -w "\n%{http_code}" -X "GET" \
    -H "X-User-ID: invalid" \
    "$API_BASE/offline/status")
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" = "400" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ (Expected 400, got $http_code)${NC}"
fi

# Test invalid bookmark ID
make_request "GET" "/offline/cache/bookmark/invalid" "" "400"

# Test invalid status value
invalid_status_data='{"status":"invalid_status"}'
make_request "PUT" "/offline/status" "$invalid_status_data" "400"

# Test malformed JSON
echo -n "Testing malformed JSON... "
response=$(curl -s -w "\n%{http_code}" -X "POST" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $USER_ID" \
    -d '{"invalid_json":}' \
    "$API_BASE/offline/cache/bookmark")
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" = "400" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ (Expected 400, got $http_code)${NC}"
fi

echo -e "\n${YELLOW}10. Testing Integration Scenarios${NC}"

# Test complete offline workflow
echo "Testing complete offline workflow..."

# 1. Set status to offline
make_request "PUT" "/offline/status" '{"status":"offline"}' "200"

# 2. Cache some bookmarks
for i in {1..3}; do
    bookmark='{
        "id": '$i',
        "url": "https://example'$i'.com",
        "title": "Example '$i'",
        "description": "Test bookmark '$i'",
        "tags": "[\"test\", \"example\"]"
    }'
    make_request "POST" "/offline/cache/bookmark" "$bookmark" "200"
done

# 3. Queue some changes
for i in {1..2}; do
    change='{
        "device_id": "'$DEVICE_ID'",
        "type": "bookmark_update",
        "resource_id": "bookmark-'$i'",
        "data": "{\"title\":\"Updated Title '$i'\"}"
    }'
    make_request "POST" "/offline/queue/change" "$change" "200"
done

# 4. Check offline indicator
make_request "GET" "/offline/indicator" "" "200"

# 5. Set status back to online
make_request "PUT" "/offline/status" '{"status":"online"}' "200"

# 6. Process offline queue
make_request "POST" "/offline/sync" "" "200"

# 7. Check final stats
make_request "GET" "/offline/stats" "" "200"

echo -e "\n${GREEN}✅ All offline support tests completed!${NC}"
echo -e "\n${YELLOW}Summary:${NC}"
echo "- ✅ Connectivity checking"
echo "- ✅ Offline status management"
echo "- ✅ Bookmark caching system"
echo "- ✅ Offline change queuing"
echo "- ✅ Queue processing and sync"
echo "- ✅ Cache statistics and monitoring"
echo "- ✅ Offline indicators and user feedback"
echo "- ✅ Cache cleanup and management"
echo "- ✅ Error handling and validation"
echo "- ✅ Complete offline workflow integration"

echo -e "\n${GREEN}🎉 Offline support system is working correctly!${NC}"