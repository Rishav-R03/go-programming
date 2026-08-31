package ratelimiter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb    *redis.Client
	script *redis.Script
}

func NewLimiter(rdb *redis.Client, luascript string) *Limiter {
	return &Limiter{
		rdb:    rdb,
		script: redis.NewScript(luascript),
	}

}

func (l *Limiter) Allow(ctx context.Context, clientID string, policy Policy) (bool, int, error) {
	key := fmt.Sprintf("rate:%s", clientID)
	now := time.Now().UnixMilli()
	windowMs := policy.Window.Milliseconds()

	reqID, err := uniqueRequestID()
	if err != nil {
		return false, 0, fmt.Errorf("generating reqID: %v", err)
	}
	res, err := l.script.Run(ctx, l.rdb, []string{key}, now, windowMs, policy.Limit, reqID).Result()
	if err != nil {
		return false, 0, fmt.Errorf("running sliding window script: %w", err)
	}
	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("unexpected script result shape: %v", err)
	}

	allowed := toInt64(vals[0]) == 1
	remaining := int(toInt64(vals[1]))
	return allowed, remaining, nil
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	default:
		return 0
	}
}

// uniqueRequestID produces a short random member for the ZSET so
// concurrent requests in the same millisecond don't collide.

func uniqueRequestID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
