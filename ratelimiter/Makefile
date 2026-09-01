.PHONY: up down build test race bench loadtest-vegeta loadtest-k6 tidy

up:
	docker compose up --build -d

down:
	docker compose down -v

tidy:
	go mod tidy

build:
	go build -o bin/rate-limiter ./cmd/api

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -bench=. -benchmem ./... -run=^$$

loadtest-vegeta:
	./scripts/loadtest_vegeta.sh

loadtest-k6:
	k6 run scripts/loadtest_k6.js
vegeta-install:
	go install github.com/tsenart/vegeta@latest
