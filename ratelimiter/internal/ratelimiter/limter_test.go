package ratelimiter

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestLimiter_ConcurrentAllow fires many concurrent requests at the same
// clientID and asserts that no more than policy.Limit are ever allowed.
// Run with: go test -race ./internal/ratelimiter/...
// Requires Redis reachable at REDIS_ADDR (defaults to localhost:6379).
func TestLimiter_ConcurrentAllow(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not reachable at %s, skipping: %v", addr, err)
	}

	script, err := os.ReadFile("../../scripts/sliding_window.lua")
	if err != nil {
		t.Fatalf("reading lua script: %v", err)
	}

	limiter := NewLimiter(rdb, string(script))
	clientID := "test-client-concurrent"
	rdb.Del(ctx, "rate:"+clientID)

	policy := Policy{Limit: 50, Window: 10 * time.Second}

	var allowedCount int64
	var wg sync.WaitGroup
	requests := 200

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _, err := limiter.Allow(ctx, clientID, policy)
			if err != nil {
				t.Errorf("Allow error: %v", err)
				return
			}
			if allowed {
				atomic.AddInt64(&allowedCount, 1)
			}
		}()
	}
	wg.Wait()

	if allowedCount != int64(policy.Limit) {
		t.Errorf("expected exactly %d allowed requests under concurrency, got %d", policy.Limit, allowedCount)
	}

	rdb.Del(ctx, "rate:"+clientID)
}
