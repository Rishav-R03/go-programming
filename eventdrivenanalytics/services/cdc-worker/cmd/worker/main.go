package main

import (
	"context"
	"log"

	"cdcworker/internal/config"
	"cdcworker/internal/consumer"
	"cdcworker/internal/model"
	"cdcworker/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

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

	syncService := service.NewSyncService(
		sourcePool,
		targetPool,
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
