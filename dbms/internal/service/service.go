package service

import (
	"context"
	"dbms/internal/domain"
	"dbms/internal/repository"
	"errors"
)

type OrderService struct {
	repo repository.PostgresOrderRepository
}

func NewOrderService(repo repository.PostgresOrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) PlaceOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("Order must be contain at least one item")
	}
	var total float64
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, errors.New("Invalid quantity")
		}
		total += item.UnitPrice * float64(item.Quantity)
	}

	order := &domain.Order{
		UserID:      userID,
		TotalAmount: total,
		Status:      "PAID",
	}

	err := s.repo.CreateOrderWithTx(ctx, order, items)
	if err != nil {
		return nil, err
	}
	return order, nil
}
