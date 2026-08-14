package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stock_quantity"`
}

type OrderItem struct {
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type Order struct {
	ID          int64       `json:"id"`
	UserID      int64       `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	Items       []OrderItem `json:"items,omitempty"`
}
type OrderRepository interface {
	CreateOrderWithTx(ctx context.Context, order *Order, items []OrderItem) error
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
}
