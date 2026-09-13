package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type OrderResponse struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
type Service struct {
	db *pgxpool.Pool
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/inventory_db?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse DB config: %v\n", err)
	}
	config.MaxConns = 25
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	svc := &Service{
		db: pool,
	}
	http.HandleFunc("/orders", svc.handleCreateOrder)
}

func (s *Service) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, OrderResponse{Status: "FAILED", Message: "Invalid JSON payload"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orderID, err := s.processOrderTransaction(ctx, req.ProductID, req.Quantity)
	if err != nil {
		respondJSON(w, http.StatusConflict, OrderResponse{Status: "FAILED", Message: err.Error()})
		return
	}
	respondJSON(w, http.StatusCreated, OrderResponse{
		OrderID: orderID,
		Status:  "COMPLETED",
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Service) processOrderTransaction(ctx context.Context, productID int64, qty int) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w\n", err)
	}
	defer tx.Rollback(ctx) // doesn't execute if committed

	//Lock the specific product row for UPDATE (Pessimistic locking)
	var currStock int
	queryLock := `SELECT stock FROM products WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryLock, productID).Scan(&currStock)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("product not found")
		}
		return 0, fmt.Errorf("failed to lock row: %w", err)
	}

	// business validation
	if currStock < qty {
		return 0, fmt.Errorf("insufficient stock")
	}
	//deduct inventory
	queryDeduct := `UPDATE products SET stock = stock - $1 WHERE id = $2`
	_, err = tx.Exec(ctx, queryDeduct, qty, productID)
	if err != nil {
		return 0, fmt.Errorf("failed to update stock: %w", err)
	}
	// Record order
	var orderID int64
	queryOrder := `INSERT INTO orders (product_id,quantity,status) VALUE ($1,$2,'COMPLETED') RETURNING id`
	err = tx.QueryRow(ctx, queryOrder, productID, qty).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return orderID, nil
}
