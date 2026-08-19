FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o analytics-service \
    ./cmd/server

FROM alpine:latest

WORKDIR /app

# 1. Copy the compiled binary
COPY --from=builder /app/analytics-service .

# 2. Copy the configs directory cleanly
COPY --from=builder /app/configs/ ./configs/

EXPOSE 8081

CMD ["./analytics-service"]