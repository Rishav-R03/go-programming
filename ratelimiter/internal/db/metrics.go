package db

import (
	"context"
	"fmt"
	"ratelimiter/internal/ratelimiter"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricRow struct {
	ClientID  string
	Timestamp time.Time
	Status    string
	LatencyMs float64
}

func BatchInsertMetrics(ctx context.Context, pool *pgxpool.Pool, rows []MetricRow) error {
	if len(rows) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	const stmt = `INSERT INTO request_metrics (client_id, ts, status, latency_ms) VALUES ($1, $2, $3, $4)`
	for _, r := range rows {
		batch.Queue(stmt, r.ClientID, r.Timestamp, r.Status, r.LatencyMs)
	}
	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert error: %v", err)
		}
	}
	return nil
}

func GetPolicyByClientID(ctx context.Context, pool *pgxpool.Pool, clientID string) (ratelimiter.Policy, error) {
	var limit int
	var windowSecond int

	err := pool.QueryRow(ctx, `SELECT req_limit, window_seconds FROM client_policies WHERE client_id = $1`, clientID).Scan(&limit, &windowSecond)
	if err == pgx.ErrNoRows {
		return ratelimiter.DefaultPolicy, nil
	}
	if err != nil {
		return ratelimiter.Policy{}, fmt.Errorf("querying policy failed: %v", err)
	}

	return ratelimiter.Policy{
		Limit:  limit,
		Window: time.Duration(windowSecond) * time.Second,
	}, nil
}
