FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# Best practice: Copy dependency files first for Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o cdc-worker \
    ./cmd/worker

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/cdc-worker .
COPY --from=builder /app/internal/config/worker.yaml ./config/worker.yaml
CMD ["./cdc-worker"]