# Multi-stage Dockerfile for openscadgen
# Supports both development and production builds

# Build stage
FROM golang:1.25-alpine AS builder

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

# Clean up export directories and other unnecessary files from examples
# This prevents 4GB+ of generated files from being included in the image
RUN find /app/examples -type d -name export -exec rm -rf {} + 2>/dev/null || true && \
    find /app/examples -type f ! -name "*.scad" ! -name "*.toml" ! -name "*.md" -delete 2>/dev/null || true

# Production stage - use Debian slim (OpenSCAD more reliable here)
FROM debian:bookworm-slim

# Install OpenSCAD and minimal dependencies in single layer
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openscad \
    ca-certificates \
    tzdata \
    git \
    && mkdir -p /usr/share/openscad/libraries && \
    git clone --depth 1 https://github.com/BelfrySCAD/BOSL2.git /usr/share/openscad/libraries/BOSL2 && \
    rm -rf /usr/share/openscad/libraries/BOSL2/.git && \
    apt-get purge -y git && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -s /bin/sh -m appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/openscadgen /usr/local/bin/openscadgen

# Copy example files (export dirs already cleaned in builder stage)
# Only copy .scad and .toml files, exclude everything else
COPY --from=builder --chown=appuser:appgroup /app/examples /app/examples

# Switch to non-root user
USER appuser

# Expose port for server mode
EXPOSE 6767

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD openscadgen --version || exit 1

# Default command - start server with examples folder
CMD ["openscadgen", "-sf", "/app/examples", "-p", "6767"]
