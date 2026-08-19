package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"analyticsservice/internal/handler"
	"analyticsservice/internal/repository"
	"analyticsservice/internal/service"
)

func main() {

	pool, err := pgxpool.New(
		context.Background(),
		"postgres://postgres:postgres@olap-db:5432/analytics",
	)

	if err != nil {
		log.Fatal(err)
	}

	repo :=
		repository.NewAnalyticsRepository(
			pool,
		)

	svc :=
		service.NewAnalyticsService(
			repo,
		)

	h :=
		handler.NewAnalyticsHandler(
			svc,
		)

	r := chi.NewRouter()

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
