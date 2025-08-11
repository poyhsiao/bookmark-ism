# Development Dockerfile
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

# Build the application with optimized flags
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./backend/cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates wget curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY config ./config

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]