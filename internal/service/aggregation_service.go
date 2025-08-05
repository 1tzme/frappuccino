package service

import (
	"sort"

	"hot-coffee/internal/repositories"
	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

type TotalSales struct {
	TotalRevenue float64    `json:"total_revenue"`
	ItemSales    []ItemSale `json:"item_sales"`
}

type ItemSale struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	QuantitySold int     `json:"quantity_sold"`
	TotalValue   float64 `json:"total_value"`
}

type PopularItem struct {
	models.MenuItem
	SalesCount int `json:"sales_count"`
}

type AggregationServiceInterface interface {
	GetTotalSales() (*TotalSales, error)
	GetPopularItems() ([]PopularItem, error)
}

type AggregationService struct {
	aggregationRepo repositories.AggregationRepositoryInterface
	logger          *logger.Logger
}

func NewAggregationService(aggregationRepo repositories.AggregationRepositoryInterface, log *logger.Logger) *AggregationService {
	return &AggregationService{
		aggregationRepo: aggregationRepo,
		logger:          log.WithComponent("aggregation_service"),
	}
}

func (s *AggregationService) GetTotalSales() (*TotalSales, error) {
	s.logger.Info("Calculating total sales report")

	orders, menuItems, err := s.aggregationRepo.GetAggregationData()
	if err != nil {
		s.logger.Error("Failed to get aggregation data for sales report", "error", err)
		return nil, err
	}

	menuMap := make(map[string]*models.MenuItem)
	for _, item := range menuItems {
		menuMap[item.ID] = item
	}

	report := &TotalSales{
		ItemSales: make([]ItemSale, 0),
	}
	itemSalesMap := make(map[string]*ItemSale)

	for _, order := range orders {
		if order.Status != "closed" {
			continue
		}
		for _, orderItem := range order.Items {
			menuItem, ok := menuMap[orderItem.ProductID]
			if !ok {
				s.logger.Warn("Product ID from an order not found in menu", "product_id", orderItem.ProductID, "order_id", order.ID)
				continue
			}

			itemValue := menuItem.Price * float64(orderItem.Quantity)
			report.TotalRevenue += itemValue

			if sale, exists := itemSalesMap[orderItem.ProductID]; exists {
				sale.QuantitySold += orderItem.Quantity
				sale.TotalValue += itemValue
			} else {
				itemSalesMap[orderItem.ProductID] = &ItemSale{
					ProductID:    orderItem.ProductID,
					ProductName:  menuItem.Name,
					QuantitySold: orderItem.Quantity,
					TotalValue:   itemValue,
				}
			}
		}
	}

	for _, sale := range itemSalesMap {
		report.ItemSales = append(report.ItemSales, *sale)
	}

	sort.Slice(report.ItemSales, func(i, j int) bool {
		return report.ItemSales[i].ProductName < report.ItemSales[j].ProductName
	})

	s.logger.Info("Total sales report calculated successfully", "total_revenue", report.TotalRevenue)
	return report, nil
}

func (s *AggregationService) GetPopularItems() ([]PopularItem, error) {
	s.logger.Info("Calculating popular items report")

	orders, menuItems, err := s.aggregationRepo.GetAggregationData()
	if err != nil {
		s.logger.Error("Failed to get aggregation data for popular items report", "error", err)
		return nil, err
	}

	salesCount := make(map[string]int)
	for _, order := range orders {
		if order.Status != "closed" {
			continue
		}
		for _, item := range order.Items {
			salesCount[item.ProductID] += item.Quantity
		}
	}

	var popularItems []PopularItem
	for _, menuItem := range menuItems {
		popularItems = append(popularItems, PopularItem{
			MenuItem:   *menuItem,
			SalesCount: salesCount[menuItem.ID],
		})
	}

	sort.Slice(popularItems, func(i, j int) bool {
		return popularItems[i].SalesCount > popularItems[j].SalesCount
	})

	s.logger.Info("Popular items report calculated successfully", "item_count", len(popularItems))
	return popularItems, nil
}
