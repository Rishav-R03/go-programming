package main

import (
	"context"
	"database/sql"
	"dbms/internal/domain"
	"dbms/internal/repository"
	"dbms/internal/service"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	conStr := "host=localhost port=5432 user=postgres password=postgres dbname=ecommerce sslmode=disable"
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer db.Close()

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(*orderRepo)
	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 2, UnitPrice: 20.99},
	}
	order, err := orderService.PlaceOrder(ctx, 1, items)
	if err != nil {
		log.Fatalf("Failed to place order: %v", err)
	}
	fmt.Printf("Order placed successfully! Order ID :%d, Total: $%.2f\n", order.ID, order.TotalAmount)

}
