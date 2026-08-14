package repository

import (
	"context"
	"database/sql"
	"dbms/internal/domain"
	"fmt"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrderWithTx(ctx context.Context, order *domain.Order, items []domain.OrderItem) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orderQuery := `INSERT INTO orders (user_id, total_amount, status) VALUES ($1, $2, $3) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, orderQuery, order.UserID, order.TotalAmount, order.Status).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range items {
		// FIXED: Replaced stock-quantity with stock_quantity
		stockQuery := `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND stock_quantity >= $1`
		res, err := tx.ExecContext(ctx, stockQuery, item.Quantity, item.ProductID)
		if err != nil {
			return err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("insufficient stock for productID : %d", item.ProductID)
		}

		itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, $4)`
		_, err = tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresOrderRepository) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.total_amount, o.status, o.created_at,
		       oi.product_id, oi.quantity, oi.unit_price
		FROM orders o
		LEFT JOIN order_items oi ON o.id = oi.order_id
		WHERE o.id = $1`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var order domain.Order
	var items []domain.OrderItem

	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.CreatedAt,
			&item.ProductID, &item.Quantity, &item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		item.OrderID = order.ID
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	order.Items = items
	return &order, nil
}
