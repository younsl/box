# Build stage
FROM rust:1.91-slim as builder

WORKDIR /build

# Copy manifests
COPY Cargo.toml ./
COPY Cargo.lock* ./

# Copy source code
COPY src ./src

# Build for release
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install CA certificates for HTTPS
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /build/target/release/promdrop /usr/local/bin/promdrop

# Create non-root user
RUN useradd -m -u 1000 promdrop

# Set working directory
WORKDIR /data

# Switch to non-root user
USER promdrop

# Set entrypoint
ENTRYPOINT ["promdrop"]

# Default command (show help)
CMD ["--help"]
