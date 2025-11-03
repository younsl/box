# Build stage
FROM rust:1.91-alpine AS builder

# Build arguments for version information
ARG GITHUB_SHA

# Set build argument as environment variable for build.rs
ENV GITHUB_SHA=${GITHUB_SHA}

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache \
    musl-dev \
    pkgconfig \
    openssl-dev \
    openssl-libs-static \
    git

# Copy manifests and build script
COPY Cargo.toml Cargo.lock build.rs ./

# Create dummy src directory to cache dependencies
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs

# Build dependencies only (this layer will be cached)
# Reduce parallelism to avoid OOM
RUN CARGO_BUILD_JOBS=2 cargo build --release && \
    rm -rf src target/release/deps/ec2_statuscheck_rebooter* target/release/ec2-statuscheck-rebooter*

# Copy actual source code
COPY src ./src

# Build actual binary with reduced parallelism
# Disable incremental compilation to reduce memory usage
RUN CARGO_INCREMENTAL=0 CARGO_BUILD_JOBS=2 cargo build --release

# Runtime stage
FROM alpine:3.22 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root group and user with explicit GID and UID (Alpine syntax)
RUN addgroup -g 1000 rebooter && \
    adduser -D -u 1000 -G rebooter -s /bin/sh rebooter

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/target/release/ec2-statuscheck-rebooter /app/ec2-statuscheck-rebooter

# Set ownership and executable permission
RUN chown -R 1000:1000 /app && \
    chmod +x /app/ec2-statuscheck-rebooter

USER 1000:1000

ENTRYPOINT ["/app/ec2-statuscheck-rebooter"]
