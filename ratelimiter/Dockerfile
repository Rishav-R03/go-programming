FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 
COPY . . 
RUN go build -o /ratelimiter ./cmd/api

FROM alpine:3.20
COPY --from=builder /ratelimiter /ratelimiter
COPY scripts /scripts
EXPOSE 8000
ENTRYPOINT ["/ratelimiter"]
