# Build stage
FROM rust:1.91-alpine AS builder

LABEL maintainer="younsl"
LABEL description="Redis console TUI for managing multiple Redis clusters"

# Install build dependencies
RUN apk add --no-cache \
    musl-dev \
    openssl-dev \
    openssl-libs-static \
    pkgconfig

WORKDIR /build

# Copy dependency files
COPY Cargo.toml Cargo.lock ./

# Create a dummy main.rs to cache dependencies
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Copy source code
COPY src ./src

# Build the actual application
RUN touch src/main.rs && \
    cargo build --release --locked

# Runtime stage
FROM alpine:3.22 AS runtime

LABEL maintainer="younsl"
LABEL description="Redis console TUI for managing multiple Redis clusters"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 redis-console && \
    adduser -D -u 1000 -G redis-console redis-console

# Create directories
RUN mkdir -p /etc/redis/clusters /workspace && \
    chown -R redis-console:redis-console /etc/redis/clusters /workspace

# Copy binary from builder
COPY --from=builder /build/target/release/redis-console /usr/local/bin/redis-console

# Switch to non-root user
USER redis-console

WORKDIR /workspace

# Set environment
ENV RUST_LOG=info

ENTRYPOINT ["/usr/local/bin/redis-console"]
