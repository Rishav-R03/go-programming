package repository

import (
	"context"
	"orderservice/internal/metrics"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderRepository struct {
	metrics *metrics.Metrics
}

func NewOrderRepository(m *metrics.Metrics) *OrderRepository {
	return &OrderRepository{metrics: m}
}

func (r *OrderRepository) CreateOrder(
	ctx context.Context,
	tx pgx.Tx,
	customerID int,
	restaurantID int,
) (int64, error) {
	var orderID int64
	err := tx.QueryRow(ctx, `INSERT INTO orders (customer_id,restaurant_id,status) VALUES ($1,$2,'PENDING') RETURNING order_id`, customerID, restaurantID).Scan(&orderID)
	return orderID, err
}

func (r *OrderRepository) CreateOrderItem(ctx context.Context, tx pgx.Tx, orderID int64, itemID int, quantity int, price float64) error {
	_, err := tx.Exec(ctx, `INSERT INTO order_items (order_id,item_id,quantity,price) VALUES ($1,$2,$3,$4)`, orderID, itemID, quantity, price)
	return err
}

func (r *OrderRepository) SaveOrder(ctx context.Context, order any) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		r.metrics.DBQueryDurationSeconds.WithLabelValues("insert_order").Observe(duration)
	}()

	// Execute DB query here
	return nil
}
