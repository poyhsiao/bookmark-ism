# syntax=docker/dockerfile:1

# Development Dockerfile - Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set environment variables for Go build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount for better performance
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy the entire source tree to maintain module structure
COPY . .

# Build the application with optimized flags
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./backend/cmd/api

# Final stage - Alpine for development with debugging tools
FROM alpine:3.21

# Install ca-certificates and debugging tools for development
RUN apk --no-cache add ca-certificates wget curl tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /build

# Copy the binary from builder (fixed path)
COPY --from=builder /build/main .

# Copy configuration files if they exist (optional)
# Note: Configuration is typically provided via environment variables

# Change ownership to non-root user
RUN chown -R appuser:appgroup /build

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]