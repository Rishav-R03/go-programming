package analytics

import (
	"context"
	"log"
	"ratelimiter/internal/db"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultFlushInterval = 2 * time.Second
	defaultFlushSize     = 500
)

// Collector buffers MetricEvents from the hot request path and flushes
// them to Postgres in batches from background workers, so no API request
// ever waits on a database write.

type Collector struct {
	eventsChan chan MetricEvent
	dbPool     *pgxpool.Pool
	workers    int
	wg         sync.WaitGroup
}

// NewCollector builds a Collector. bufferSize sizes the channel (how many
// in-flight events can queue before Submit starts dropping); workers is
// how many flushWorker goroutines drain it in parallel.

func NewCollector(bufferSize int, workers int, dbPool *pgxpool.Pool) *Collector {
	if bufferSize <= 0 {
		bufferSize = 10_000
	}
	if workers <= 0 {
		workers = 2
	}
	return &Collector{
		eventsChan: make(chan MetricEvent, bufferSize),
		dbPool:     dbPool,
		workers:    workers,
	}
}

// Submit enqueues an event without blocking the caller. If the buffer is
// full, the event is dropped and logged rather than backing up the
// request path — metrics are best-effort, correctness of the rate limit

func (c *Collector) Start(ctx context.Context) {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.flushWorker(ctx, i)
	}
}

func (c *Collector) Stop() {
	close(c.eventsChan)
	c.wg.Wait()
}

// Submit enqueues an event without blocking the caller. If the buffer is
// full, the event is dropped and logged rather than backing up the
// request path — metrics are best-effort, correctness of the rate limit
// itself never depends on them.
func (c *Collector) Submit(event MetricEvent) {
	select {
	case c.eventsChan <- event:
	default:
		log.Printf("analytics: buffer full, dropping metric for client=%s", event.ClientID)
	}
}

func (c *Collector) flushWorker(ctx context.Context, workerID int) {
	defer c.wg.Done()

	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()

	buf := make([]MetricEvent, 0, defaultFlushSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}
		rows := make([]db.MetricRow, len(buf))
		for i, e := range buf {
			rows[i] = db.MetricRow{
				ClientID:  e.ClientID,
				Timestamp: e.Timestamp,
				Status:    string(e.Status),
				LatencyMs: e.LatencyMs,
			}
		}

		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := db.BatchInsertMetrics(flushCtx, c.dbPool, rows); err != nil {
			log.Printf("analytics worker %d: batch insert failed: %v", workerID, err)
		}
		cancel()
		buf = buf[:0]
	}

	for {
		select {
		case e, ok := <-c.eventsChan:
			if !ok {
				flush()
				return
			}
			buf = append(buf, e)
			if len(buf) >= defaultFlushSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}
