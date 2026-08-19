package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	CDCEventsProcessedTotal *prometheus.CounterVec
	CDCSyncFailuresTotal    *prometheus.CounterVec
	CDCSyncDurationSeconds  *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		CDCEventsProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cdc_events_processed_total",
				Help: "Total number of CDC events processed by the worker",
			},
			[]string{"event_type"},
		),
		CDCSyncFailuresTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cdc_sync_failures_total",
			},
			[]string{"event_type", "reason"},
		),
		CDCSyncDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cdc_sync_duration_seconds",
				Help:    "Time taken to process and sync CDC event to OLAP target",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
			},
			[]string{"event_type"},
		),
	}

	prometheus.MustRegister(
		m.CDCEventsProcessedTotal,
		m.CDCSyncFailuresTotal,
		m.CDCSyncDurationSeconds,
	)

	return m
}
