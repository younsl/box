# Build stage
FROM rust:1.90-slim AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy manifests
COPY Cargo.toml Cargo.lock* ./

# Copy source code
COPY src ./src

# Build release binary
RUN cargo build --release --locked

# Runtime stage
FROM alpine:3.22 AS runtime

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    libgcc

# Create non-root user
RUN adduser -D -u 1000 -s /bin/sh appuser && \
    chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /app/target/release/elasticache-backup /usr/local/bin/elasticache-backup

# Switch to non-root user
USER appuser

# Health check endpoint (not applicable for batch jobs)
# Health monitoring should be done via Kubernetes Job/CronJob status

ENTRYPOINT ["/usr/local/bin/elasticache-backup"]
