package main

import (
	"context"
	"log"
	"net/http"
	"orderservice/internal/config"
	"orderservice/internal/database"
	"orderservice/internal/handler"
	"orderservice/internal/repository"
	"orderservice/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load("../../configs/order-service.yaml")
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
	repo := repository.NewOrderRepository()
	orderService := service.NewOrderService(pool, repo)
	orderHandler := handler.NewOrderHandler(orderService)

	r := chi.NewRouter()

	r.Post("/orders", orderHandler.CreateOrder)
	log.Println("server started on: " + cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, r))
}
