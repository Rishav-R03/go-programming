package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequestsTotal          *prometheus.CounterVec
	HTTPRequestDurationSeconds *prometheus.HistogramVec
	AnalyticsQueryDuration     *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "handler", "status"},
		),
		HTTPRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "handler"},
		),
		AnalyticsQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "analytics_query_duration_seconds",
				Help:    "Duration of heavy OLAP analytics queries in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"query_type"},
		),
	}

	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDurationSeconds,
		m.AnalyticsQueryDuration,
	)

	return m
}
