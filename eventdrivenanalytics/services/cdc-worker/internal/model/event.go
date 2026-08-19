package model

import "time"

type OrderEvent struct {
	OrderID      int64     `json:"order_id"`
	CustomerID   int       `json:"customer_id"`
	RestaurantID int       `json:"restaurant_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
