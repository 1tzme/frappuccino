package repositories

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace in-memory data aggregation with SQL-based queries
// 1. Implement JOIN queries to aggregate data across orders and menu_items tables
// 2. Create database views for common aggregation patterns
// 3. Replace manual data fetching with optimized SQL aggregate functions
// 4. Add proper database indexing for performance optimization
// 5. Implement database-specific reporting features (window functions, CTEs)

import (
	"frappuccino/models"
	"frappuccino/pkg/logger"
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
