package model

type CreateOrderRequest struct {
	CustomerID   int     `json:"customer_id" validate:"required,gt=0"`
	RestaurantID int     `json:"restaurant_id" validate:"required,gt=0"`
	ItemID       int     `json:"item_id" validate:"required,gt=0"`
	Quantity     int     `json:"quantity" validate:"required,gt=0"`
	Price        float64 `json:"price" validate:"required,gt=0"`
}

type CreateOrderResponse struct {
	OrderID int64 `json:"order_id"`
}
