package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	fmt.Println("Connected to database")

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse pgxpool config")
	}

	config.MinConns = 5
	config.MaxConns = 25
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 1 * time.Hour

	pool, err = pgxpool.NewWithConfig(ctx, config)
}
