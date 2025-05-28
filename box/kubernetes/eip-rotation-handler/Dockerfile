FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o eip-rotation-handler ./cmd/eip-rotation-handler

FROM alpine:1.24.0 AS runner
RUN apk --no-cache \
        add \
        ca-certificates \
        tzdata \
    && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/eip-rotation-handler .
EXPOSE 8080
CMD ["/app/eip-rotation-handler"]