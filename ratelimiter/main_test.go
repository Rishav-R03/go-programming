package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"ratelimiter/internal/analytics"
	"ratelimiter/internal/middleware"
	"ratelimiter/internal/ratelimiter"
)

// setupTestServer builds the same limiter/policy/middleware stack as
// main() minus Postgres — the collector is given a nil pool and a very
// generous buffer so Submit() never blocks; we don't assert on metrics
// here, only on rate-limit behavior and throughput. Skips if Redis isn't
// reachable, same convention as limiter_test.go.
func setupTestServer(t testing.TB, limit int, window time.Duration) (*httptest.Server, func()) {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not reachable at %s, skipping: %v", addr, err)
	}

	scriptBytes, err := os.ReadFile("scripts/sliding_window.lua")
	if err != nil {
		t.Fatalf("reading lua script: %v", err)
	}

	limiter := ratelimiter.NewLimiter(rdb, string(scriptBytes))
	policies := ratelimiter.NewPolicyStore()
	policies.Set("bench-client", ratelimiter.Policy{Limit: limit, Window: window})

	// nil pool is fine: Collector.Submit only ever touches the channel;
	// the pool is only dereferenced inside flushWorker's batch insert,
	// which we never wait on in these tests.
	collector := analytics.NewCollector(10_000, 2, nil)
	// Intentionally NOT calling collector.Start() — without Postgres
	// there's nothing to flush to, and Submit() is non-blocking regardless.

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.RateLimit(limiter, policies, collector)(target)

	srv := httptest.NewServer(handler)

	cleanup := func() {
		srv.Close()
		rdb.Del(context.Background(), "rate:bench-client")
		rdb.Close()
	}
	return srv, cleanup
}

// TestRateLimit_EndToEnd is a smoke test: N requests under the limit all
// succeed, and the next one is rejected with 429.
func TestRateLimit_EndToEnd(t *testing.T) {
	limit := 10
	srv, cleanup := setupTestServer(t, limit, 5*time.Second)
	defer cleanup()

	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < limit; i++ {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
		req.Header.Set("X-API-KEY", "bench-client")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	req.Header.Set("X-API-KEY", "bench-client")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("final request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 once over limit, got %d", resp.StatusCode)
	}
}

// BenchmarkRateLimitMiddleware measures end-to-end request latency through
// the full middleware chain (Redis round trip included). Run with:
//
//	go test -bench=. -benchmem ./... -run=^$
//
// A high, generous limit avoids the benchmark itself tripping 429s and
// skewing the numbers toward error handling instead of the hot path.
func BenchmarkRateLimitMiddleware(b *testing.B) {
	srv, cleanup := setupTestServer(b, 10_000_000, time.Minute)
	defer cleanup()

	client := &http.Client{Timeout: 2 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
			req.Header.Set("X-API-KEY", "bench-client")
			resp, err := client.Do(req)
			if err != nil {
				b.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}
