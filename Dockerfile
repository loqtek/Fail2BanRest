# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fail2restV2 ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hash-password ./cmd/hash-password

# Runtime stage
FROM alpine:latest

# Install fail2ban-client, sudo (if using sudo mode), and wget for healthcheck
RUN apk add --no-cache \
    fail2ban \
    ca-certificates \
    tzdata \
    wget \
    sudo \
    && rm -rf /var/cache/apk/*

# Create fail2ban group (GID should match host if possible, default 999)
# Create a non-root user and add to fail2ban group
RUN addgroup -g 999 fail2ban 2>/dev/null || true && \
    addgroup -g 1000 fail2rest && \
    adduser -D -u 1000 -G fail2rest -G fail2ban fail2rest || \
    adduser -D -u 1000 -G fail2rest fail2rest

# Create directories
RUN mkdir -p /app /etc/fail2rest && \
    chown -R fail2rest:fail2rest /app /etc/fail2rest

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/fail2restV2 /app/
COPY --from=builder /build/hash-password /app/

# Copy example config (user should mount their own)
COPY config.example.yaml /etc/fail2rest/config.example.yaml

# Make binaries executable
RUN chmod +x /app/fail2restV2 /app/hash-password

# Switch to non-root user
USER fail2rest

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/fail2restV2", "-config", "/etc/fail2rest/config.yaml"]

