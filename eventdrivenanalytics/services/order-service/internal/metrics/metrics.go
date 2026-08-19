package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	OrdersCreatedTotal         prometheus.Counter
	HTTPRequestsTotal          *prometheus.CounterVec
	HTTPRequestDurationSeconds *prometheus.HistogramVec
	DBQueryDurationSeconds     *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		OrdersCreatedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "orders_created_total",
				Help: "Total number of orders successfully created",
			},
		),
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
		DBQueryDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
			},
			[]string{"operation"},
		),
	}

	prometheus.MustRegister(
		m.OrdersCreatedTotal,
		m.HTTPRequestsTotal,
		m.HTTPRequestDurationSeconds,
		m.DBQueryDurationSeconds,
	)

	return m
}
