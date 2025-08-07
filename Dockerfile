# Development Dockerfile
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies and verify module
RUN go mod download && go mod verify

# Copy the entire source code
COPY . .

# Verify and tidy dependencies
RUN go mod tidy && \
    go mod verify

# Build the application
RUN echo "=== Build Environment Debug ===" && \
    echo "Go version: $(go version)" && \
    echo "GOOS: $(go env GOOS), GOARCH: $(go env GOARCH)" && \
    echo "Working directory: $(pwd)" && \
    echo "Checking required directories..." && \
    test -d backend/cmd/api || (echo "ERROR: backend/cmd/api directory not found" && exit 1) && \
    test -f backend/cmd/api/main.go || (echo "ERROR: backend/cmd/api/main.go not found" && exit 1) && \
    echo "All required files present" && \
    echo "=== Starting Build ===" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o main ./backend/cmd/api && \
    echo "=== Build Successful ===" && \
    ls -la main && \
    echo "Binary created successfully"

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates wget curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]