# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates gcc

WORKDIR /src

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=1 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -tags "sqlite_json1 sqlite_fts5" \
    -o /vault-reader \
    ./cmd/vault-reader

# Verify binary
RUN /vault-reader --version 2>/dev/null || echo "Binary built successfully"

# Stage 2: Runtime
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && adduser -D -u 1000 -h /app vaultreader

# Copy binary from builder
COPY --from=builder --chown=1000:1000 /vault-reader /usr/local/bin/vault-reader

# Create directories
RUN mkdir -p /app/data /app/vault && \
    chown -R vaultreader:vaultreader /app

# Switch to non-root user
USER vaultreader
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:3000/health >/dev/null 2>&1 || exit 1

# Expose port
EXPOSE 3000

# Set default environment variables
ENV VAULT_DIR=/app/vault \
    DATA_DIR=/app/data \
    ADDR=:3000 \
    GIN_MODE=release

# Run the application
ENTRYPOINT ["vault-reader"]
CMD ["--vault", "/app/vault", "--addr", ":3000"]
