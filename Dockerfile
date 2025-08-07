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

# Verify the module structure and dependencies
RUN ls -la backend/pkg/storage/ && \
    go mod tidy && \
    go list -m all

# Build the application with verbose output for debugging
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o main ./backend/cmd/api

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