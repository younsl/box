# Build stage
FROM rust:1.91-alpine3.22 AS builder

# Build arguments for version info
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /app

# Install build dependencies including UPX for binary compression
RUN apk add --no-cache musl-dev git upx

# Copy manifests
COPY Cargo.toml Cargo.lock ./
COPY build.rs ./

# Create dummy main to cache dependencies
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Copy source code
COPY src ./src

# Build for release with version info from build args
RUN touch src/main.rs && \
    VERGEN_GIT_SHA=${GIT_COMMIT} VERGEN_BUILD_TIMESTAMP="${BUILD_DATE}" \
    cargo build --release --locked

# Strip debug symbols and compress with UPX
RUN strip /app/target/release/filesystem-cleaner && \
    upx --best --lzma /app/target/release/filesystem-cleaner

# Runtime stage
FROM alpine:3.22 AS runtime

# Create non-root user (no ca-certificates or tzdata - not needed for filesystem operations)
RUN addgroup -g 1000 cleaner && \
    adduser -D -u 1000 -G cleaner cleaner

USER cleaner
WORKDIR /app

# Copy compressed binary with ownership set (no separate chown layer needed)
COPY --from=builder --chown=cleaner:cleaner /app/target/release/filesystem-cleaner /app/filesystem-cleaner

ENTRYPOINT ["/app/filesystem-cleaner"]
