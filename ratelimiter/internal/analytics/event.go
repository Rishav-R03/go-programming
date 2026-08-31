package analytics

import "time"

// Status is the outcome of a rate-limit decision for one request.
type Status string

const (
	StatusAllowed Status = "allowed"
	StatusBlocked Status = "blocked"
)

// MetricEvent is a single rate-limit decision, queued for async
// persistence to Postgres. Keep this struct small — it's copied onto
// a channel on every single request.
type MetricEvent struct {
	ClientID  string
	Timestamp time.Time
	Status    Status
	LatencyMs float64
}
