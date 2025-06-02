# ========================
# Build Stage
# ========================
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o eip-rotation-handler ./cmd/eip-rotation-handler

# ========================
# Runtime Stage
# ========================
FROM alpine:3.22 AS runner

RUN apk --no-cache \
        add \
        ca-certificates \
        tzdata \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/eip-rotation-handler .

# Create non-root linux user and switch to it
RUN adduser -D -u 10001 appuser
USER appuser

# Expose port 8080
EXPOSE 8080

ENTRYPOINT ["/app/eip-rotation-handler"]