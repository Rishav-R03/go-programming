package middleware

import (
	"net/http"
	"ratelimiter/internal/analytics"
	"ratelimiter/internal/ratelimiter"
	"strconv"
	"time"
)

// PolicySource is anything that can resolve a client's Policy. Both
// ratelimiter.PolicyStore and a Postgres-backed lookup (db.GetPolicyByClientID
// wrapped in a small adapter) satisfy this — the middleware doesn't care
// which, keeping it swappable without touching this file.

type PolicySource interface {
	Get(clientID string) ratelimiter.Policy
}

// RateLimit wraps an http.Handler with the sliding-window check. On every
// request it: resolves the client, asks the limiter, records a metric
// event (non-blocking), sets the X-RateLimit-* headers, and either
// forwards the request or returns 429.

func RateLimit(limiter *ratelimiter.Limiter, policies PolicySource, collector *analytics.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			clientID := extractClientID(r)
			policy := policies.Get(clientID)
			allowed, remaining, err := limiter.Allow(r.Context(), clientID, policy)
			if err != nil {
				http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
				return
			}

			status := analytics.StatusAllowed
			if !allowed {
				status = analytics.StatusBlocked
			}
			collector.Submit(analytics.MetricEvent{
				ClientID:  clientID,
				Timestamp: time.Now(),
				Status:    status,
				LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
			})

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(policy.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractClientID(r *http.Request) string {
	if key := r.Header.Get("X-API-KEY"); key != "" {
		return key
	}
	return r.RemoteAddr
}
