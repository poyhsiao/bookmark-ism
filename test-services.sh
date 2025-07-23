#!/bin/bash

echo "🔍 Testing Bookmark Sync Service Components..."
echo ""

# Test PostgreSQL
echo "📊 Testing PostgreSQL..."
if docker-compose exec -T supabase-db psql -U postgres -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ PostgreSQL is working"
else
    echo "❌ PostgreSQL is not working"
fi

# Test Redis
echo "📦 Testing Redis..."
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is working"
else
    echo "❌ Redis is not working"
fi

# Test Typesense
echo "🔍 Testing Typesense..."
if curl -s http://localhost:8108/health | grep -q "ok"; then
    echo "✅ Typesense is working"
else
    echo "❌ Typesense is not working"
fi

# Test MinIO
echo "💾 Testing MinIO..."
if curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✅ MinIO is working"
else
    echo "❌ MinIO is not working"
fi

echo ""
echo "🎉 Basic services test completed!"
echo ""
echo "📋 Service URLs:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - Typesense: http://localhost:8108"
echo "  - MinIO Console: http://localhost:9001"
echo "  - MinIO API: http://localhost:9000"
echo ""
echo "🔧 To access services:"
echo "  - PostgreSQL: docker-compose exec supabase-db psql -U postgres"
echo "  - Redis: docker-compose exec redis redis-cli"
echo "  - Typesense: curl http://localhost:8108/health"
echo "  - MinIO Console: open http://localhost:9001 (admin/dev-minio-123)"