package main

import (
	"context"
	"log"
	"net/http"
	"orderservice/internal/config"
	"orderservice/internal/database"
	"orderservice/internal/handler"
	"orderservice/internal/metrics"
	"orderservice/internal/middleware"
	"orderservice/internal/repository"
	"orderservice/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load("config/order-service.yaml")
	if err != nil {
		log.Fatal(err)
	}
	pool, err := database.NewPool(
		context.Background(),
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)
	if err != nil {
		log.Fatal(err)
	}
	m := metrics.NewMetrics()
	repo := repository.NewOrderRepository(m)
	orderService := service.NewOrderService(pool, repo)
	orderHandler := handler.NewOrderHandler(orderService, m)

	r := chi.NewRouter()

	// 1. Declare ALL middlewares FIRST
	r.Use(middleware.MetricsMiddleware(m))

	// 2. Register metric scrape routes
	r.Handle("/metrics", promhttp.Handler())

	// 3. Register application business routes
	r.Post("/orders", orderHandler.CreateOrder)
	r.Get("/health", handler.Health)
	log.Println("server started on: " + cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, r))
}
