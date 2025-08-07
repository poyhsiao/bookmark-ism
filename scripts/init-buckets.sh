#!/bin/bash

# Initialize MinIO buckets for S3-compatible storage
# This script creates the necessary buckets for the bookmark service
# MinIO serves as the S3-compatible storage implementation

set -e

echo "🪣 Initializing MinIO buckets..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Wait for MinIO to be ready
print_status "Waiting for MinIO to be ready..."
until curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; do
    echo "Waiting for MinIO..."
    sleep 2
done

print_success "MinIO is ready!"

# Install MinIO client if not available
if ! command -v mc &> /dev/null; then
    print_status "Installing MinIO client..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install minio/stable/mc
        else
            curl -O https://dl.min.io/client/mc/release/darwin-amd64/mc
            chmod +x mc
            sudo mv mc /usr/local/bin/
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -O https://dl.min.io/client/mc/release/linux-amd64/mc
        chmod +x mc
        sudo mv mc /usr/local/bin/
    else
        echo "Please install MinIO client manually: https://docs.min.io/docs/minio-client-quickstart-guide.html"
        exit 1
    fi
fi

# Configure MinIO client
print_status "Configuring MinIO client..."
mc alias set local http://localhost:9000 ${MINIO_ROOT_USER:-minioadmin} ${MINIO_ROOT_PASSWORD:-minioadmin}

# Create buckets
print_status "Creating buckets..."

# Main bookmarks bucket
mc mb local/bookmarks --ignore-existing
print_success "Created 'bookmarks' bucket"

# Screenshots bucket
mc mb local/screenshots --ignore-existing
print_success "Created 'screenshots' bucket"

# Avatars bucket
mc mb local/avatars --ignore-existing
print_success "Created 'avatars' bucket"

# Backups bucket
mc mb local/backups --ignore-existing
print_success "Created 'backups' bucket"

# Exports bucket
mc mb local/exports --ignore-existing
print_success "Created 'exports' bucket"

# Set bucket policies for public read access where needed
print_status "Setting bucket policies..."

# Screenshots should be publicly readable
cat > /tmp/screenshots-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": ["*"]
      },
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::screenshots/*"]
    }
  ]
}
EOF

mc policy set-json /tmp/screenshots-policy.json local/screenshots
print_success "Set public read policy for screenshots bucket"

# Avatars should be publicly readable
cat > /tmp/avatars-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": ["*"]
      },
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::avatars/*"]
    }
  ]
}
EOF

mc policy set-json /tmp/avatars-policy.json local/avatars
print_success "Set public read policy for avatars bucket"

# Clean up temporary files
rm -f /tmp/screenshots-policy.json /tmp/avatars-policy.json

print_success "🎉 MinIO buckets initialized successfully!"

echo ""
echo "Created buckets:"
echo "  - bookmarks   (private) - Main bookmark data"
echo "  - screenshots (public)  - Website screenshots"
echo "  - avatars     (public)  - User avatars"
echo "  - backups     (private) - System backups"
echo "  - exports     (private) - Data exports"
echo ""
echo "MinIO Console: http://localhost:9001"
echo "Username: ${MINIO_ROOT_USER:-minioadmin}"
echo "Password: ${MINIO_ROOT_PASSWORD:-minioadmin}"
echo ""