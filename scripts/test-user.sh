#!/bin/bash

# User Profile Management Testing Script
# This script tests the user profile management endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

API_BASE_URL="http://localhost:8080/api/v1"

echo "👤 Testing User Profile Management Endpoints"
echo "==========================================="

# First, register a user to get an access token
print_status "Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profile-test@example.com",
    "password": "testpassword123",
    "username": "profiletest",
    "display_name": "Profile Test User"
  }')

if echo "$REGISTER_RESPONSE" | grep -q '"success":true'; then
    print_success "User registration successful"
    ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    echo "Access token: ${ACCESS_TOKEN:0:50}..."
else
    print_error "User registration failed"
    echo "Response: $REGISTER_RESPONSE"
    exit 1
fi

echo ""

# Test get profile
print_status "Testing get profile..."
PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE_URL/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$PROFILE_RESPONSE" | grep -q '"success":true'; then
    print_success "Get profile successful"
    echo "Profile: $PROFILE_RESPONSE"
else
    print_error "Get profile failed"
    echo "Response: $PROFILE_RESPONSE"
fi

echo ""

# Test update profile
print_status "Testing update profile..."
UPDATE_RESPONSE=$(curl -s -X PUT "$API_BASE_URL/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Updated Profile Test User",
    "username": "updatedprofiletest"
  }')

if echo "$UPDATE_RESPONSE" | grep -q '"success":true'; then
    print_success "Update profile successful"
    echo "Updated profile: $UPDATE_RESPONSE"
else
    print_error "Update profile failed"
    echo "Response: $UPDATE_RESPONSE"
fi

echo ""

# Test get preferences
print_status "Testing get preferences..."
PREFS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/user/preferences" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$PREFS_RESPONSE" | grep -q '"success":true'; then
    print_success "Get preferences successful"
    echo "Preferences: $PREFS_RESPONSE"
else
    print_error "Get preferences failed"
    echo "Response: $PREFS_RESPONSE"
fi

echo ""

# Test update preferences
print_status "Testing update preferences..."
UPDATE_PREFS_RESPONSE=$(curl -s -X PUT "$API_BASE_URL/user/preferences" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "theme": "dark",
    "gridSize": "large",
    "defaultView": "list",
    "language": "zh-CN"
  }')

if echo "$UPDATE_PREFS_RESPONSE" | grep -q '"success":true'; then
    print_success "Update preferences successful"
    echo "Updated preferences: $UPDATE_PREFS_RESPONSE"
else
    print_error "Update preferences failed"
    echo "Response: $UPDATE_PREFS_RESPONSE"
fi

echo ""

# Test get stats
print_status "Testing get stats..."
STATS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/user/stats" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$STATS_RESPONSE" | grep -q '"success":true'; then
    print_success "Get stats successful"
    echo "Stats: $STATS_RESPONSE"
else
    print_error "Get stats failed"
    echo "Response: $STATS_RESPONSE"
fi

echo ""

# Test export data
print_status "Testing export data..."
EXPORT_RESPONSE=$(curl -s -X POST "$API_BASE_URL/user/export" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$EXPORT_RESPONSE" | grep -q '"success":true'; then
    print_success "Export data successful"
    echo "Export data (truncated): ${EXPORT_RESPONSE:0:200}..."
else
    print_error "Export data failed"
    echo "Response: $EXPORT_RESPONSE"
fi

echo ""

# Test avatar upload (create a simple test image)
print_status "Testing avatar upload..."
# Create a simple 1x1 PNG image (base64 encoded)
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9jU77mgAAAABJRU5ErkJggg==" | base64 -d > /tmp/test_avatar.png

AVATAR_RESPONSE=$(curl -s -X POST "$API_BASE_URL/user/avatar" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "avatar=@/tmp/test_avatar.png")

if echo "$AVATAR_RESPONSE" | grep -q '"success":true'; then
    print_success "Avatar upload successful"
    echo "Avatar response: $AVATAR_RESPONSE"
else
    print_error "Avatar upload failed"
    echo "Response: $AVATAR_RESPONSE"
fi

# Clean up test file
rm -f /tmp/test_avatar.png

echo ""

# Test unauthorized access (without token)
print_status "Testing unauthorized access..."
UNAUTH_RESPONSE=$(curl -s -X GET "$API_BASE_URL/user/profile")

if echo "$UNAUTH_RESPONSE" | grep -q '"error"'; then
    print_success "Unauthorized access properly blocked"
else
    print_error "Unauthorized access not properly blocked"
    echo "Response: $UNAUTH_RESPONSE"
fi

echo ""
print_success "🎉 User profile management testing completed!"