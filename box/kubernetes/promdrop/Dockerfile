# Build stage
FROM rust:1.91-alpine as builder

# Install build dependencies
RUN apk add --no-cache musl-dev

WORKDIR /build

# Copy manifests
COPY Cargo.toml ./
COPY Cargo.lock* ./

# Copy source code
COPY src ./src

# Build for release
RUN cargo build --release

# Runtime stage
FROM alpine:3.22

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /build/target/release/promdrop /usr/local/bin/promdrop

# Create non-root user
RUN adduser -D -u 1000 promdrop

# Set working directory
WORKDIR /data

# Switch to non-root user
USER promdrop

# Set entrypoint
ENTRYPOINT ["promdrop"]

# Default command (show help)
CMD ["--help"]
