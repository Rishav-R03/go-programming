package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

func main() {
	const totalQueries = 200
	fmt.Printf("--- PostgreSQL Connection Benchmark (%d queries) ---\n\n", totalQueries)
	// Benchmark 1: Opening a new connection for every query
	durationUnpooled := benchmarkNoPooling(totalQueries)

	// Benchmark 2: Reusing connections with an optimized pool
	durationPooled := benchmarkWithPooling(totalQueries)

	// Summary
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Without Pool (New Conn / Query): %v\n", durationUnpooled)
	fmt.Printf("With Pool (Reused Conns):        %v\n", durationPooled)
	if durationPooled > 0 {
		fmt.Printf("Speedup Factor:                 %.2fx faster\n", float64(durationUnpooled)/float64(durationPooled))
	}
}

func benchmarkNoPooling(iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatalf("Failed to open connection: %v", err)
		}

		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(0)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var num int
		err = db.QueryRowContext(ctx, "SELECT 1").Scan(&num)
		cancel()
		if err != nil {
			log.Fatalf("Open query failed: %v", err)
		}
		db.Close()
	}
	return time.Since(start)
}

func benchmarkWithPooling(iterations int) time.Duration {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection pool: %v", err)
	}
	defer db.Close()

	// Explicit Connection Pool Tuning
	db.SetMaxOpenConns(25)                 // Maximum open connections to the database
	db.SetMaxIdleConns(25)                 // Maximum idle connections in the pool
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum time a connection may be reused
	db.SetConnMaxIdleTime(2 * time.Minute) // Maximum time a connection may remain idle

	// Warm up the pool with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		log.Fatalf("Failed to ping database: %v", err)
	}
	cancel()

	start := time.Now()

	for i := 0; i < iterations; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var num int
		err := db.QueryRowContext(ctx, "SELECT 1").Scan(&num)
		cancel()

		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
	}

	return time.Since(start)
}
