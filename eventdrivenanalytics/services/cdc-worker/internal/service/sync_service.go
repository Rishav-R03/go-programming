package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cdcworker/internal/metrics"
)

type SyncService struct {
	source  *pgxpool.Pool
	target  *pgxpool.Pool
	metrics *metrics.Metrics
}

type OrderAggregate struct {
	OrderID          int64
	DateKey          int32
	RestaurantID     int64
	CustomerID       int64
	RestaurantName   string
	RestaurantCity   string
	CustomerName     string
	TotalItemCount   int64
	TotalOrderAmount float64
	OrderStatus      string
	CreatedAt        time.Time
}

func NewSyncService(source *pgxpool.Pool, target *pgxpool.Pool, m *metrics.Metrics) *SyncService {
	return &SyncService{
		source:  source,
		target:  target,
		metrics: m,
	}
}

func (s *SyncService) SyncOrder(ctx context.Context, orderID int64) error {
	start := time.Now()
	eventType := "order_sync"

	defer func() {
		if s.metrics != nil {
			duration := time.Since(start).Seconds()
			s.metrics.CDCSyncDurationSeconds.WithLabelValues(eventType).Observe(duration)
		}
	}()

	// 1. SELECT from Source DB (oltp-db)
	selectQuery := `
        SELECT
            o.order_id,
            TO_CHAR(o.created_at, 'YYYYMMDD')::INTEGER,
            r.restaurant_id,
            c.customer_id,
            r.name,
            r.city,
            c.name,
            COALESCE(SUM(oi.quantity), 0),
            COALESCE(SUM(oi.quantity * oi.price), 0),
            o.status,
            o.created_at
        FROM orders o
        JOIN restaurants r ON o.restaurant_id = r.restaurant_id
        JOIN customers c ON o.customer_id = c.customer_id
        LEFT JOIN order_items oi ON o.order_id = oi.order_id
        WHERE o.order_id = $1
        GROUP BY o.order_id, r.restaurant_id, c.customer_id;
    `

	var agg OrderAggregate
	err := s.source.QueryRow(ctx, selectQuery, orderID).Scan(
		&agg.OrderID,
		&agg.DateKey,
		&agg.RestaurantID,
		&agg.CustomerID,
		&agg.RestaurantName,
		&agg.RestaurantCity,
		&agg.CustomerName,
		&agg.TotalItemCount,
		&agg.TotalOrderAmount,
		&agg.OrderStatus,
		&agg.CreatedAt,
	)
	if err != nil {
		if s.metrics != nil {
			s.metrics.CDCSyncFailuresTotal.WithLabelValues(eventType, "source_fetch_failed").Inc()
		}
		return fmt.Errorf("failed to fetch order from source: %w", err)
	}

	// 2. INSERT into Target DB (olap-db)
	insertQuery := `
        INSERT INTO fact_order_sales (
            order_id, date_key, restaurant_id, customer_id,
            restaurant_name, restaurant_city, customer_name,
            total_item_count, total_order_amount, order_status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_id) DO UPDATE SET
            order_status = EXCLUDED.order_status,
            total_item_count = EXCLUDED.total_item_count,
            total_order_amount = EXCLUDED.total_order_amount;
    `

	_, err = s.target.Exec(
		ctx,
		insertQuery,
		agg.OrderID,
		agg.DateKey,
		agg.RestaurantID,
		agg.CustomerID,
		agg.RestaurantName,
		agg.RestaurantCity,
		agg.CustomerName,
		agg.TotalItemCount,
		agg.TotalOrderAmount,
		agg.OrderStatus,
		agg.CreatedAt,
	)
	if err != nil {
		if s.metrics != nil {
			s.metrics.CDCSyncFailuresTotal.WithLabelValues(eventType, "target_insert_failed").Inc()
		}
		return fmt.Errorf("failed to insert fact into target: %w", err)
	}

	if s.metrics != nil {
		s.metrics.CDCEventsProcessedTotal.WithLabelValues(eventType).Inc()
	}

	return nil
}
