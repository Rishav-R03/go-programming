package model

type DashboardResponse struct {
	RestaurantName string  `json:"restaurant_name"`
	TotalOrders    int64   `json:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
	AvgOrderValue  float64 `json:"avg_order_value"`
}

type RevenueResponse struct {
	RestaurantID int64   `json:"restaurant_id"`
	Revenue      float64 `json:"revenue"`
}

type OrdersResponse struct {
	RestaurantID int64 `json:"restaurant_id"`
	Orders       int64 `json:"orders"`
}
