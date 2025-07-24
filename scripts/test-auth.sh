#!/bin/bash

# Authentication Testing Script
# This script tests the authentication endpoints

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

echo "🔐 Testing Authentication Endpoints"
echo "=================================="

# Test user registration
print_status "Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "username": "testuser",
    "display_name": "Test User"
  }')

if echo "$REGISTER_RESPONSE" | grep -q '"success":true'; then
    print_success "User registration successful"
    ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    echo "Access token: ${ACCESS_TOKEN:0:50}..."
else
    print_error "User registration failed"
    echo "Response: $REGISTER_RESPONSE"
fi

echo ""

# Test user login
print_status "Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123"
  }')

if echo "$LOGIN_RESPONSE" | grep -q '"success":true'; then
    print_success "User login successful"
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
    echo "Access token: ${ACCESS_TOKEN:0:50}..."
    echo "Refresh token: ${REFRESH_TOKEN:0:50}..."
else
    print_error "User login failed"
    echo "Response: $LOGIN_RESPONSE"
fi

echo ""

# Test protected endpoint (get profile)
if [ -n "$ACCESS_TOKEN" ]; then
    print_status "Testing protected endpoint (get profile)..."
    PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE_URL/auth/profile" \
      -H "Authorization: Bearer $ACCESS_TOKEN")

    if echo "$PROFILE_RESPONSE" | grep -q '"success":true'; then
        print_success "Profile retrieval successful"
        echo "Profile: $PROFILE_RESPONSE"
    else
        print_error "Profile retrieval failed"
        echo "Response: $PROFILE_RESPONSE"
    fi

    echo ""
fi

# Test token refresh
if [ -n "$REFRESH_TOKEN" ]; then
    print_status "Testing token refresh..."
    REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/refresh" \
      -H "Content-Type: application/json" \
      -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

    if echo "$REFRESH_RESPONSE" | grep -q '"success":true'; then
        print_success "Token refresh successful"
        NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
        echo "New access token: ${NEW_ACCESS_TOKEN:0:50}..."
    else
        print_error "Token refresh failed"
        echo "Response: $REFRESH_RESPONSE"
    fi

    echo ""
fi

# Test logout
if [ -n "$ACCESS_TOKEN" ]; then
    print_status "Testing user logout..."
    LOGOUT_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/logout" \
      -H "Authorization: Bearer $ACCESS_TOKEN")

    if echo "$LOGOUT_RESPONSE" | grep -q '"success":true'; then
        print_success "User logout successful"
    else
        print_error "User logout failed"
        echo "Response: $LOGOUT_RESPONSE"
    fi

    echo ""
fi

# Test password reset
print_status "Testing password reset..."
RESET_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/reset" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }')

if echo "$RESET_RESPONSE" | grep -q '"success":true'; then
    print_success "Password reset request successful"
else
    print_error "Password reset request failed"
    echo "Response: $RESET_RESPONSE"
fi

echo ""
print_success "🎉 Authentication testing completed!"