# Build stage
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with version info
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o promdrop ./cmd/promdrop

# Runtime stage
FROM scratch

# Add metadata labels for GitHub Container Registry
LABEL org.opencontainers.image.source=https://github.com/younsl/promdrop
LABEL org.opencontainers.image.description="A tool to generate prometheus metric-drop config"
LABEL org.opencontainers.image.licenses=MIT

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/promdrop /promdrop

# Run as non-root user
USER 65534:65534

ENTRYPOINT ["/promdrop"]
CMD ["--help"]
