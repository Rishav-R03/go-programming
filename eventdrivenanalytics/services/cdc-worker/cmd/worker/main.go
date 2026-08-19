package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cdcworker/internal/config"
	"cdcworker/internal/consumer"
	"cdcworker/internal/metrics"
	"cdcworker/internal/model"
	"cdcworker/internal/service"
)

func main() {
	// 1. Initialize Prometheus Metrics
	m := metrics.NewMetrics()

	// 2. Start HTTP Server for Prometheus scraping on port 8082
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("cdc worker metrics server running on :8082/metrics")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	// Works cleanly both in container and locally if run from app root
	cfg, err := config.Load("config/worker.yaml")
	if err != nil {
		log.Fatal(err)
	}

	sourceDSN := config.BuildDSN(
		cfg.SourceDB.Host,
		cfg.SourceDB.Port,
		cfg.SourceDB.User,
		cfg.SourceDB.Password,
		cfg.SourceDB.DBName,
	)

	sourcePool, err := pgxpool.New(
		context.Background(),
		sourceDSN,
	)
	if err != nil {
		log.Fatal(err)
	}

	targetDSN := config.BuildDSN(
		cfg.TargetDB.Host,
		cfg.TargetDB.Port,
		cfg.TargetDB.User,
		cfg.TargetDB.Password,
		cfg.TargetDB.DBName,
	)

	targetPool, err := pgxpool.New(
		context.Background(),
		targetDSN,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Inject Metrics into SyncService
	syncService := service.NewSyncService(
		sourcePool,
		targetPool,
		m,
	)

	err = consumer.StartListener(
		sourceDSN,
		func(event model.OrderEvent) {
			log.Printf(
				"received order %d",
				event.OrderID,
			)

			err := syncService.SyncOrder(
				context.Background(),
				event.OrderID,
			)

			if err != nil {
				log.Println(err)
			}
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
