# syntax=docker/dockerfile:1

# Development Dockerfile with enhanced debugging and testing capabilities
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount for better performance
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy only necessary source code (backend directory and go files)
COPY backend ./backend
COPY config ./config

# Enhanced build with debugging and validation
RUN --mount=type=cache,target=/root/.cache/go-build \
    echo "=== Development Build Environment Debug ===" && \
    echo "Go version: $(go version)" && \
    echo "GOOS: $(go env GOOS), GOARCH: $(go env GOARCH)" && \
    echo "Working directory: $(pwd)" && \
    echo "Checking required directories..." && \
    test -d backend/cmd/api || (echo "ERROR: backend/cmd/api directory not found" && exit 1) && \
    test -f backend/cmd/api/main.go || (echo "ERROR: backend/cmd/api/main.go not found" && exit 1) && \
    echo "All required files present" && \
    echo "=== Starting Development Build ===" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o main ./backend/cmd/api && \
    echo "=== Development Build Successful ===" && \
    ls -la main && \
    file main

# Test stage for development
FROM builder AS test
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    echo "=== Running Tests ===" && \
    go test ./backend/... -v

# Final stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget curl tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY config ./config

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]