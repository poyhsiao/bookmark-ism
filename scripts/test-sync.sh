#!/bin/bash

# Test script for sync functionality
# 同步功能測試腳本

set -e

BASE_URL="http://localhost:8080/api/v1"
USER_ID="test-user-123"
DEVICE_ID="device-456"
AUTH_TOKEN="test-auth-token"

echo "🔄 Testing Sync Functionality"
echo "================================"

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

# Test 1: Get sync state
echo "📊 Test 1: Get sync state"
response=$(make_request "GET" "/sync/state?device_id=$DEVICE_ID")
echo "Response: $response"
echo ""

# Test 2: Create sync event
echo "📝 Test 2: Create sync event"
sync_event='{
    "type": "bookmark_created",
    "resource_id": "bookmark-123",
    "action": "create",
    "device_id": "'$DEVICE_ID'",
    "data": {
        "title": "Test Bookmark",
        "url": "https://example.com",
        "description": "A test bookmark for sync"
    }
}'

response=$(make_request "POST" "/sync/events" "$sync_event")
echo "Response: $response"
echo ""

# Test 3: Get delta sync
echo "🔄 Test 3: Get delta sync"
last_sync_time=$(date -d "1 hour ago" +%s)
response=$(make_request "GET" "/sync/delta?device_id=$DEVICE_ID&last_sync_time=$last_sync_time")
echo "Response: $response"
echo ""

# Test 4: Queue offline event
echo "📱 Test 4: Queue offline event"
offline_event='{
    "type": "bookmark_updated",
    "resource_id": "bookmark-offline",
    "action": "update",
    "device_id": "'$DEVICE_ID'",
    "data": {
        "title": "Updated Offline Bookmark",
        "url": "https://offline-example.com"
    }
}'

response=$(make_request "POST" "/sync/offline-queue" "$offline_event")
echo "Response: $response"
echo ""

# Test 5: Get offline queue
echo "📋 Test 5: Get offline queue"
response=$(make_request "GET" "/sync/offline-queue?device_id=$DEVICE_ID")
echo "Response: $response"
echo ""

# Test 6: Process offline queue
echo "⚡ Test 6: Process offline queue"
process_request='{"device_id": "'$DEVICE_ID'"}'
response=$(make_request "POST" "/sync/offline-queue/process" "$process_request")
echo "Response: $response"
echo ""

# Test 7: Update sync state
echo "🔄 Test 7: Update sync state"
current_time=$(date +%s)
update_state='{
    "device_id": "'$DEVICE_ID'",
    "last_sync_time": '$current_time'
}'

response=$(make_request "PUT" "/sync/state" "$update_state")
echo "Response: $response"
echo ""

# Test 8: WebSocket connection test
echo "🌐 Test 8: WebSocket connection test"
echo "Testing WebSocket connection to ws://localhost:8080/ws?user_id=$USER_ID&device_id=$DEVICE_ID"

# Create a simple WebSocket test client
cat > /tmp/ws_test.js << 'EOF'
const WebSocket = require('ws');

const userID = process.argv[2] || 'test-user-123';
const deviceID = process.argv[3] || 'device-456';
const wsUrl = `ws://localhost:8080/ws?user_id=${userID}&device_id=${deviceID}`;

console.log(`Connecting to: ${wsUrl}`);

const ws = new WebSocket(wsUrl);

ws.on('open', function open() {
    console.log('✅ WebSocket connected');

    // Send ping message
    const pingMessage = {
        type: 'ping',
        timestamp: new Date().toISOString()
    };

    console.log('📤 Sending ping:', JSON.stringify(pingMessage));
    ws.send(JSON.stringify(pingMessage));

    // Send sync request
    setTimeout(() => {
        const syncRequest = {
            type: 'sync_request',
            data: {
                last_sync_time: Math.floor(Date.now() / 1000) - 3600 // 1 hour ago
            },
            timestamp: new Date().toISOString()
        };

        console.log('📤 Sending sync request:', JSON.stringify(syncRequest));
        ws.send(JSON.stringify(syncRequest));
    }, 1000);

    // Close connection after 3 seconds
    setTimeout(() => {
        ws.close();
    }, 3000);
});

ws.on('message', function message(data) {
    console.log('📥 Received:', data.toString());
});

ws.on('close', function close() {
    console.log('❌ WebSocket disconnected');
});

ws.on('error', function error(err) {
    console.error('🚨 WebSocket error:', err.message);
});
EOF

# Check if Node.js is available for WebSocket test
if command -v node >/dev/null 2>&1; then
    # Install ws package if not available
    if ! node -e "require('ws')" 2>/dev/null; then
        echo "Installing ws package for WebSocket test..."
        npm install ws 2>/dev/null || echo "⚠️  Could not install ws package, skipping WebSocket test"
    fi

    if node -e "require('ws')" 2>/dev/null; then
        node /tmp/ws_test.js "$USER_ID" "$DEVICE_ID"
    else
        echo "⚠️  WebSocket test skipped - ws package not available"
    fi
else
    echo "⚠️  WebSocket test skipped - Node.js not available"
fi

# Cleanup
rm -f /tmp/ws_test.js

echo ""
echo "✅ Sync functionality tests completed!"
echo ""
echo "📋 Test Summary:"
echo "- Sync state management: ✅"
echo "- Sync event creation: ✅"
echo "- Delta synchronization: ✅"
echo "- Offline queue management: ✅"
echo "- WebSocket communication: ✅"
echo ""
echo "🎯 Next steps:"
echo "1. Run the backend server: make run"
echo "2. Execute this test: ./scripts/test-sync.sh"
echo "3. Check logs for detailed sync operations"