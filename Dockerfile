# Multi-stage Dockerfile for openscadgen
# Supports both development and production builds

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install templ for template generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o openscadgen .

# Production stage
FROM alpine:latest

# Install OpenSCAD and other dependencies
RUN apk add --no-cache \
    openscad \
    bash \
    curl \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/openscadgen /usr/local/bin/openscadgen

# Copy example files for testing
COPY --from=builder /app/examples /app/examples

# Make binary executable
RUN chmod +x /usr/local/bin/openscadgen

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port for server mode
EXPOSE 6767

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD openscadgen --version || exit 1

# Default command
CMD ["openscadgen", "--help"]
