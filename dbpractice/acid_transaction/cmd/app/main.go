package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

func transferMoney(ctx context.Context, conn *pgx.Conn, fromID, toID int, amount float64, simulateError bool) error {
	//1. Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	//2. Deduct from sender
	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
	if err != nil {
		return fmt.Errorf("failed to deduct balance")
	}

	if simulateError {
		return fmt.Errorf("SIMULATED NETWORK FAILURE BEFORE COMPLETION")
	}

	//3. Add to receiver
	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
	if err != nil {
		return fmt.Errorf("failed to add balance")
	}

	//4. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("Transaction committed successfully!")
	return nil
}

func concurrentDeduction(connStr string, accountID int, amount float64, useLocking bool, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Printf("Tx begin error: %v", err)
	}
	defer tx.Rollback(ctx)

	var currBalance float64
	dml := "SELECT balance FROM accounts WHERE id=$1"
	if useLocking {
		dml += " FOR UPDATE"
	}

	err = tx.QueryRow(ctx, dml, accountID).Scan(&currBalance)
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}
	// race window
	time.Sleep(100 * time.Millisecond)

	newBalance := currBalance - amount
	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = $1 WHERE id=$2", newBalance, accountID)
	if err != nil {
		log.Printf("Update error: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Commit error: %v", err)
		return
	}
}

func runIsolationTest(connStr string) {
	ctx := context.Background()
	conn, _ := pgx.Connect(ctx, connStr)
	defer conn.Close(ctx)

	conn.Exec(ctx, "UPDATE accounts SET balance = 500.00 WHERE id = 1")
	fmt.Println("----Test A: Without lock (RACE CONDITION)----")
	fmt.Println("Starting balance: $500.00. Running 2 concurrent $100 withdrawals...")

	var wg sync.WaitGroup
	wg.Add(2)
	// launch 2 go routines concurrently reading balance before either 1 writes
	go concurrentDeduction(connStr, 1, 100.00, false, &wg)
	go concurrentDeduction(connStr, 1, 100.00, false, &wg)
	wg.Wait()

	var balance float64
	conn.QueryRow(ctx, "SELECT balance FROM accounts WHERE id=1").Scan(&balance)
	fmt.Printf("Final balance without lock: $%.2f (Expected $300.00 - Lost Update!)\n\n", balance)

	// Reset balance for locked test
	conn.Exec(ctx, "UPDATE accounts SET balance = 500.00 WHERE id = 1")

	fmt.Println("--- Test B: With FOR UPDATE Lock ---")
	fmt.Println("Starting balance: $500.00. Running 2 concurrent $100 withdrawals with row locking...")

	wg.Add(2)

	go concurrentDeduction(connStr, 1, 100.00, true, &wg)
	go concurrentDeduction(connStr, 1, 100.00, true, &wg)
	wg.Wait()

	conn.QueryRow(ctx, "SELECT balance FROM accounts WHERE id=1").Scan(&balance)
	fmt.Printf("Final balance with FOR update: $%.2f (Correct!)\n", balance)
}

func main() {
	ctx := context.Background()
	connStr := "postgres://admin:admin@localhost:5432/acid_demo?sslmode=disable" // Adjust credentials
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("--- Test 1: Simulating Failure (Atomicity Check) ---")
	err = transferMoney(ctx, conn, 1, 2, 100.00, true)
	if err != nil {
		fmt.Printf("Transaction aborted: %v\n", err)
	}

	fmt.Println("\n--- Test 2: Successful Transfer ---")
	err = transferMoney(ctx, conn, 1, 2, 100.00, false)
	if err != nil {
		fmt.Printf("Transaction failed: %v\n", err)
	}

	runIsolationTest(connStr)

}
