# Build stage
FROM golang:1.25-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with build args as environment variables
RUN VERSION=${VERSION} COMMIT=${COMMIT} DATE=${DATE} make build

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 cleaner && \
    adduser -D -u 1000 -G cleaner cleaner

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/filesystem-cleaner /app/filesystem-cleaner

# Set ownership
RUN chown -R cleaner:cleaner /app

USER cleaner

ENTRYPOINT ["/app/filesystem-cleaner"]