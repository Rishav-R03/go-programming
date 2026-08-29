# Stage 1: Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency definition files first (leverages Docker cache)
COPY /go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Final lightweight image
FROM alpine:latest

# Create non-root user for security
RUN adduser -S -D -H -h /app appuser

WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/main .

USER appuser

EXPOSE 3000

CMD ["./main"]