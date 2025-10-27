# Build stage
FROM rust:1.90-alpine AS builder

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

# Copy source code
COPY src ./src

# Build release binary with static linking
# Note: Git commit will be extracted from GITHUB_SHA env var in CI builds
RUN cargo build --release

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
