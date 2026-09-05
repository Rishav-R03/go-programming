package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

// Optimistic locking
func BookTicketOptimistic(ctx context.Context, db *sql.DB, eventID int) error {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		var availableTickets, version int
		err := db.QueryRowContext(ctx, "SELECT available_tickets, version FROM events WHERE id = $1", eventID).Scan(&availableTickets, &version)
		if err != nil {
			return err
		}
		if availableTickets <= 0 {
			return errors.New("sold out")
		}

		res, err := db.ExecContext(ctx, `UPDATE events SET available_tickets = available_tickets -1,verion = version + 1 WHERE id = $1 AND version = $2`, eventID, version)
		if err != nil {
			return err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 1 {
			return nil
		}
		// Add random backoff inside the loop before retrying
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
	}
	return errors.New("conflict: maximum retries reached")
}

// Pessimistic locking

func BookTicketPessimistic(ctx context.Context, db *sql.DB, eventID int) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var availabletickets int
	err = tx.QueryRowContext(ctx, "SELECT available_tickets FROM events WHERE id = $1 FOR UPDATE", eventID).Scan(&availabletickets)
	if err != nil {
		return err
	}

	if availabletickets <= 0 {
		return errors.New("sold out")
	}

	_, err = tx.ExecContext(ctx, "UPDATE events SET available_tickets = available_tickets - 1 WHERE id = $1", eventID)
	if err != nil {
		return err
	}
	return tx.Commit()

}

func RunStressTest(bookingFunc func(context.Context, *sql.DB, int) error, db *sql.DB) {
	var wg sync.WaitGroup
	conCurrUser := 50
	successCount := int32(0)
	failureCount := int32(0)

	for i := 0; i < conCurrUser; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bookingFunc(context.Background(), db, 1)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failureCount, 1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("Successful: %d | Failure: %d", successCount, failureCount)
}

func main() {
	connStr := "postgres://admin:admin@localhost:5432/ticketdb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Printf("unable to ping db: %v", err)
	}
	resetDatabase := func() {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS events (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				available_tickets INT NOT NULL CHECK (available_tickets >= 0),
				version INT NOT NULL DEFAULT 1
			);
			TRUNCATE TABLE events RESTART IDENTITY;
			INSERT INTO events (name, available_tickets, version) VALUES ('Rock Concert 2026', 10, 1);
		`)
		if err != nil {
			log.Fatalf("failed to reset table: %v", err)
		}
	}

	fmt.Println("--- Testing Optimistic Locking ---")
	resetDatabase()
	RunStressTest(BookTicketOptimistic, db)

	// 2. Run Pessimistic Concurrency Test
	fmt.Println("\n--- Testing Pessimistic Locking ---")
	resetDatabase()
	RunStressTest(BookTicketPessimistic, db)

}
