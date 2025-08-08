package service

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace in-memory aggregation logic with SQL-based reporting
// 1. Implement SQL aggregate functions (SUM, COUNT, GROUP BY) for sales reports
// 2. Create database views for complex reporting queries
// 3. Add database-based analytics and business intelligence features
// 4. Replace manual data processing with stored procedures
// 5. Implement efficient pagination and filtering for large datasets

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"frappuccino/internal/repositories"
	"frappuccino/models"
	"frappuccino/pkg/logger"
)

type AggregationServiceInterface interface {
	GetTotalSales() (*TotalSales, error)
	GetPopularItems() ([]PopularItem, error)
	SearchFullText(req SearchRequest) (*repositories.SearchResult, error)
	GetOrderedItemsByPeriod(req OrderedItemsByPeriodRequest) (*repositories.OrderedItemsByPeriodResult, error)
}

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

type SearchRequest struct {
	Query    string   `json:"query"`
	Filters  []string `json:"filters"`
	MinPrice *float64 `json:"min_price"`
	MaxPrice *float64 `json:"max_price"`
}

type OrderedItemsByPeriodRequest struct {
	Period string `json:"period"`
	Month  string `json:"month"`
	Year   string `json:"year"`
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

func (s *AggregationService) SearchFullText(req SearchRequest) (*repositories.SearchResult, error) {
	s.logger.Info("Processing full text search request", "query", req.Query, "filters", req.Filters)

	if err := s.validateSearchRequest(req); err != nil {
		s.logger.Warn("Invalid search request", "error", err)
		return nil, err
	}

	result, err := s.aggregationRepo.SearchFullText(req.Query, req.Filters, req.MinPrice, req.MaxPrice)
	if err != nil {
		s.logger.Error("Failed to perform full text search", "error", err)
		return nil, err
	}

	s.logger.Info("Full text search completed successfully", "total_matches", result.TotalMatches)
	return result, nil
}

func (s *AggregationService) GetOrderedItemsByPeriod(req OrderedItemsByPeriodRequest) (*repositories.OrderedItemsByPeriodResult, error) {
	s.logger.Info("Processing ordered items by period request", "period", req.Period, "month", req.Month, "year", req.Year)

	if err := s.validatePeriodRequest(req); err != nil {
		s.logger.Warn("Invalid period request", "error", err)
		return nil, err
	}

	result, err := s.aggregationRepo.GetOrderedItemsByPeriod(req.Period, req.Month, req.Year)
	if err != nil {
		s.logger.Error("Failed to get ordered items by period", "error", err)
		return nil, err
	}

	s.logger.Info("Ordered items by period retrieved successfully", "period", req.Period, "items_count", len(result.OrderedItems))
	return result, nil
}

// validation functions
func (s *AggregationService) validateSearchRequest(req SearchRequest) error {
	if strings.TrimSpace(req.Query) == "" {
		return errors.New("search query cannot be empty")
	}

	if len(req.Query) > 255 {
		return errors.New("search query is too long (max 255 characters)")
	}

	validFilters := []string{"orders", "menu", "all"}
	for _, filter := range req.Filters {
		if !contains(validFilters, filter) {
			return errors.New("invalid search filter (allowed: orders, menu, all)")
		}
	}

	if req.MinPrice != nil && *req.MinPrice < 0 {
		return errors.New("minimum price cannot be negative")
	}

	if req.MaxPrice != nil && *req.MaxPrice < 0 {
		return errors.New("maximum price cannot be negative")
	}

	if req.MinPrice != nil && req.MaxPrice != nil && *req.MinPrice > *req.MaxPrice {
		return errors.New("minimum price cannot be greater than maximum price")
	}

	return nil
}

func (s *AggregationService) validatePeriodRequest(req OrderedItemsByPeriodRequest) error {
	if req.Period != "day" && req.Period != "month" {
		return errors.New("invalid period (allowed: day, month)")
	}

	if req.Period == "day" {
		if req.Month == "" {
			return errors.New("month is required when period is 'day'")
		}

		validMonths := []string{
			"january", "february", "march", "april", "may", "june",
			"july", "august", "september", "october", "november", "december",
		}

		if !contains(validMonths, strings.ToLower(req.Month)) {
			return errors.New("invalid month name")
		}
	}

	if req.Period == "month" && req.Year != "" {
		if year, err := strconv.Atoi(req.Year); err != nil || year < 2000 || year > 2100 {
			return errors.New("invalid year (must be between 2000 and 2100)")
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
