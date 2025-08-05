package repositories

import (
	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

type AggregationRepositoryInterface interface {
	GetAggregationData() (orders []*models.Order, menuItems []*models.MenuItem, err error)
}

type AggregationRepository struct {
	orderRepo OrderRepositoryInterface
	menuRepo  MenuRepositoryInterface
	logger    *logger.Logger
}

func NewAggregationRepository(orderRepo OrderRepositoryInterface, menuRepo MenuRepositoryInterface, log *logger.Logger) *AggregationRepository {
	return &AggregationRepository{
		orderRepo: orderRepo,
		menuRepo:  menuRepo,
		logger:    log.WithComponent("aggregation_repository"),
	}
}

func (r *AggregationRepository) GetAggregationData() (orders []*models.Order, menuItems []*models.MenuItem, err error) {
	r.logger.Info("Fetching data for aggregation reports")

	orders, err = r.orderRepo.GetAll()
	if err != nil {
		r.logger.Error("Failed to get orders for aggregation", "error", err)
		return nil, nil, err
	}

	menuItems, err = r.menuRepo.GetAll()
	if err != nil {
		r.logger.Error("Failed to get menu items for aggregation", "error", err)
		return nil, nil, err
	}

	return orders, menuItems, nil
}
