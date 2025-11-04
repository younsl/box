# Build stage - Alpine with musl for multi-arch support
FROM --platform=$BUILDPLATFORM rust:1.91-alpine AS builder

# Build arguments for cross-compilation
ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /app

# Install build dependencies for musl
RUN apk add --no-cache \
    musl-dev \
    pkgconfig \
    openssl-dev \
    openssl-libs-static

# Copy source files
COPY Cargo.toml Cargo.lock* ./
COPY src ./src

# Build for the target platform
# rust:1.91-alpine already has the correct musl toolchain for the native arch
RUN cargo build --release

# Runtime stage - Minimal Alpine (multi-arch)
FROM alpine:3.22 AS runtime

# Install only essential runtime dependencies
RUN apk add --no-cache \
    ca-certificates && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -u 1000 -s /bin/sh appuser

# Copy binary from builder
COPY --from=builder /app/target/release/elasticache-backup /usr/local/bin/elasticache-backup

# Switch to non-root user
USER appuser

ENTRYPOINT ["/usr/local/bin/elasticache-backup"]
