FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build \
    -o order-service \
    ./cmd/server

FROM alpine

WORKDIR /app

COPY --from=builder \
    /app/order-service .

EXPOSE 8000

CMD ["./order-service"]