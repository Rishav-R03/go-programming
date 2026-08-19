FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# 1. Optimizes caching by downloading dependencies before copying full source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 2. Builds a statically linked CGO-disabled binary (recommended for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o order-service \
    ./cmd/server

FROM alpine:latest

WORKDIR /app

# 3. Copy BOTH the compiled binary AND the config file
COPY --from=builder /app/order-service .
COPY --from=builder /app/internal/config/order-service.yaml ./config/order-service.yaml

EXPOSE 8000

CMD ["./order-service"]