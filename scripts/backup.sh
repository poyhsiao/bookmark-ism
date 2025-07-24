#!/bin/bash

# Database Backup Script for Bookmark Sync Service
# This script creates backups of the PostgreSQL database and MinIO storage

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DB_BACKUP_FILE="$BACKUP_DIR/database_backup_$TIMESTAMP.sql"
MINIO_BACKUP_DIR="$BACKUP_DIR/minio_backup_$TIMESTAMP"

# Create backup directory
mkdir -p "$BACKUP_DIR"
mkdir -p "$MINIO_BACKUP_DIR"

echo "🗄️ Bookmark Sync Service - Database Backup"
echo "==========================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running"
    exit 1
fi

# Check if database container is running
if ! docker ps --format "table {{.Names}}" | grep -q "^supabase-db$"; then
    print_error "Database container is not running. Start it with 'make docker-up'"
    exit 1
fi

# Backup PostgreSQL database
print_status "Creating PostgreSQL database backup..."
if docker exec supabase-db pg_dump -U postgres -d postgres > "$DB_BACKUP_FILE"; then
    print_success "Database backup created: $DB_BACKUP_FILE"

    # Compress the backup
    gzip "$DB_BACKUP_FILE"
    print_success "Database backup compressed: ${DB_BACKUP_FILE}.gz"
else
    print_error "Failed to create database backup"
    exit 1
fi

# Backup MinIO data (if container is running)
if docker ps --format "table {{.Names}}" | grep -q "^bookmark-minio$"; then
    print_status "Creating MinIO storage backup..."

    # Copy MinIO data from container
    if docker cp bookmark-minio:/data "$MINIO_BACKUP_DIR/"; then
        print_success "MinIO backup created: $MINIO_BACKUP_DIR"

        # Compress the backup
        tar -czf "${MINIO_BACKUP_DIR}.tar.gz" -C "$BACKUP_DIR" "$(basename "$MINIO_BACKUP_DIR")"
        rm -rf "$MINIO_BACKUP_DIR"
        print_success "MinIO backup compressed: ${MINIO_BACKUP_DIR}.tar.gz"
    else
        print_warning "Failed to backup MinIO data"
    fi
else
    print_warning "MinIO container is not running, skipping storage backup"
fi

# Create backup metadata
METADATA_FILE="$BACKUP_DIR/backup_metadata_$TIMESTAMP.json"
cat > "$METADATA_FILE" << EOF
{
    "timestamp": "$TIMESTAMP",
    "date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "database_backup": "${DB_BACKUP_FILE}.gz",
    "minio_backup": "${MINIO_BACKUP_DIR}.tar.gz",
    "version": "1.0.0",
    "type": "full_backup"
}
EOF

print_success "Backup metadata created: $METADATA_FILE"

# Cleanup old backups (keep last 7 days)
print_status "Cleaning up old backups (keeping last 7 days)..."
find "$BACKUP_DIR" -name "database_backup_*.sql.gz" -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "minio_backup_*.tar.gz" -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "backup_metadata_*.json" -mtime +7 -delete 2>/dev/null || true

print_success "Old backups cleaned up"

# Display backup summary
echo ""
echo "📊 Backup Summary:"
echo "  Timestamp: $TIMESTAMP"
echo "  Database backup: ${DB_BACKUP_FILE}.gz"
if [ -f "${MINIO_BACKUP_DIR}.tar.gz" ]; then
    echo "  MinIO backup: ${MINIO_BACKUP_DIR}.tar.gz"
fi
echo "  Metadata: $METADATA_FILE"
echo ""

# Calculate backup sizes
DB_SIZE=$(du -h "${DB_BACKUP_FILE}.gz" 2>/dev/null | cut -f1 || echo "N/A")
if [ -f "${MINIO_BACKUP_DIR}.tar.gz" ]; then
    MINIO_SIZE=$(du -h "${MINIO_BACKUP_DIR}.tar.gz" 2>/dev/null | cut -f1 || echo "N/A")
    echo "  Database size: $DB_SIZE"
    echo "  MinIO size: $MINIO_SIZE"
else
    echo "  Database size: $DB_SIZE"
fi

echo ""
print_success "🎉 Backup completed successfully!"
echo ""
echo "To restore from this backup:"
echo "  Database: gunzip -c ${DB_BACKUP_FILE}.gz | docker exec -i supabase-db psql -U postgres -d postgres"
if [ -f "${MINIO_BACKUP_DIR}.tar.gz" ]; then
    echo "  MinIO: tar -xzf ${MINIO_BACKUP_DIR}.tar.gz && docker cp $(basename "$MINIO_BACKUP_DIR")/data bookmark-minio:/"
fi