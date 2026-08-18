package service

import (
	"context"
	"orderservice/internal/model"
	"orderservice/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService struct {
	pool *pgxpool.Pool
	repo *repository.OrderRepository
}

func NewOrderService(pool *pgxpool.Pool, repo *repository.OrderRepository) *OrderService {
	return &OrderService{
		pool: pool,
		repo: repo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (int64, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	orderID, err := s.repo.CreateOrder(ctx, tx, req.CustomerID, req.RestaurantID)
	if err != nil {
		return 0, err
	}
	err = s.repo.CreateOrderItem(ctx, tx, orderID, req.ItemID, req.Quantity, req.Price)
	if err != nil {
		return 0, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}
