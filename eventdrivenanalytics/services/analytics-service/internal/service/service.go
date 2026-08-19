package service

import (
	"context"

	"analyticsservice/internal/model"
	"analyticsservice/internal/repository"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsService(
	repo *repository.AnalyticsRepository,
) *AnalyticsService {

	return &AnalyticsService{
		repo: repo,
	}
}

func (s *AnalyticsService) Dashboard(
	ctx context.Context,
	restaurantID int64,
) (*model.DashboardResponse, error) {

	return s.repo.Dashboard(
		ctx,
		restaurantID,
	)
}
