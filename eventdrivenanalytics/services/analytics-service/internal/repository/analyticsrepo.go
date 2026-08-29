package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"analyticsservice/internal/metrics"
	"analyticsservice/internal/model"
)

type AnalyticsRepository struct {
	pool    *pgxpool.Pool
	metrics *metrics.Metrics
}

func NewAnalyticsRepository(
	pool *pgxpool.Pool, m *metrics.Metrics,
) *AnalyticsRepository {
	return &AnalyticsRepository{
		pool:    pool,
		metrics: m,
	}
}

func (r *AnalyticsRepository) Dashboard(
	ctx context.Context,
	restaurantID int64,
) (*model.DashboardResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		if r.metrics != nil {
			r.metrics.AnalyticsQueryDuration.WithLabelValues("dashboard").Observe(duration)
		}
	}()

	query := `
    SELECT
        restaurant_name,
        COUNT(order_id),
        COALESCE(
            SUM(total_order_amount),
            0
        ),
        COALESCE(
            AVG(total_order_amount),
            0
        )
    FROM fact_order_sales
    WHERE restaurant_id = $1
    GROUP BY restaurant_name
    `

	var resp model.DashboardResponse

	err := r.pool.QueryRow(
		ctx,
		query,
		restaurantID,
	).Scan(
		&resp.RestaurantName,
		&resp.TotalOrders,
		&resp.TotalRevenue,
		&resp.AvgOrderValue,
	)

	return &resp, err
}
