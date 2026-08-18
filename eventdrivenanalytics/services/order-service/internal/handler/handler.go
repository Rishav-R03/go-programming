package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"orderservice/internal/model"
	"orderservice/internal/service"
	"orderservice/internal/validator"
	"time"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validator.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	orderID, err := h.service.CreateOrder(ctx, req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			http.Error(w, "transaction timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := model.CreateOrderResponse{OrderID: orderID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
