package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"analyticsservice/internal/handler"
	"analyticsservice/internal/metrics"
	"analyticsservice/internal/middleware"
	"analyticsservice/internal/repository"
	"analyticsservice/internal/service"
)

func main() {
	// 1. Initialize Prometheus Metrics
	m := metrics.NewMetrics()

	pool, err := pgxpool.New(
		context.Background(),
		"postgres://postgres:postgres@olap-db:5432/analytics",
	)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Pass metrics instance into repository
	repo := repository.NewAnalyticsRepository(pool, m)
	svc := service.NewAnalyticsService(repo)
	h := handler.NewAnalyticsHandler(svc)

	r := chi.NewRouter()

	// 3. Register global HTTP request metrics middleware
	r.Use(middleware.MetricsMiddleware(m))

	// 4. Expose Prometheus metrics scrape route
	r.Handle("/metrics", promhttp.Handler())

	r.Get(
		"/analytics/dashboard",
		h.Dashboard,
	)

	log.Println(
		"analytics service started :8081",
	)

	log.Fatal(
		http.ListenAndServe(
			":8081",
			r,
		),
	)
}
