package repositories

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace in-memory data aggregation with SQL-based queries
// 1. Implement JOIN queries to aggregate data across orders and menu_items tables
// 2. Create database views for common aggregation patterns
// 3. Replace manual data fetching with optimized SQL aggregate functions
// 4. Add proper database indexing for performance optimization
// 5. Implement database-specific reporting features (window functions, CTEs)

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"frappuccino/models"
	"frappuccino/pkg/database"
	"frappuccino/pkg/logger"
)

type AggregationRepositoryInterface interface {
	GetAggregationData() (orders []*models.Order, menuItems []*models.MenuItem, err error)
	SearchFullText(query string, filters []string, minPrice, maxPrice *float64) (*SearchResult, error)
	GetOrderedItemsByPeriod(period, month, year string) (*OrderedItemsByPeriodResult, error)
}

type AggregationRepository struct {
	db     *database.DB
	logger *logger.Logger
}

type SearchResult struct {
	MenuItems    []MenuSearchResult  `json:"menu_items"`
	Orders       []OrderSearchResult `json:"orders"`
	TotalMatches int                 `json:"total_matches"`
}

type MenuSearchResult struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Relevance   float64 `json:"relevance"`
}

type OrderSearchResult struct {
	ID           string   `json:"id"`
	CustomerName string   `json:"customer_name"`
	Items        []string `json:"items"`
	Total        float64  `json:"total"`
	Relevance    float64  `json:"relevance"`
}

type OrderedItemsByPeriodResult struct {
	Period       string           `json:"period"`
	Month        string           `json:"month,omitempty"`
	Year         string           `json:"year,omitempty"`
	OrderedItems []map[string]int `json:"orderedItems"`
}

func NewAggregationRepository(db *database.DB, logger *logger.Logger) *AggregationRepository {
	return &AggregationRepository{
		db:     db,
		logger: logger.WithComponent("aggregation_repository"),
	}
}

func (r *AggregationRepository) GetAggregationData() (orders []*models.Order, menuItems []*models.MenuItem, err error) {
	r.logger.Info("Fetching data for aggregation reports")

	ordersQuery := `
		SELECT o.id, o.customer_name, o.special_instructions, o.status, 
		       o.total_amount, o.created_at, o.updated_at
		FROM orders o
		ORDER BY o.created_at DESC`

	orderRows, err := r.db.Query(ordersQuery)
	if err != nil {
		r.logger.Error("Failed to query orders for aggregation", "error", err)
		return nil, nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer orderRows.Close()

	orderMap := make(map[string]*models.Order)
	for orderRows.Next() {
		order := &models.Order{}
		var specialInstructions sql.NullString

		err := orderRows.Scan(&order.ID, &order.CustomerName, &specialInstructions, &order.Status, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			r.logger.Error("Failed to scan order", "error", err)
			return nil, nil, fmt.Errorf("failed to scan order: %v", err)
		}

		if specialInstructions.Valid {
			order.SpecialInstructions = specialInstructions.String
		}

		order.Items = []models.OrderItem{}
		orderMap[order.ID] = order
		orders = append(orders, order)
	}

	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.menu_item_id, oi.quantity, 
		       oi.price_at_time, oi.customizations
		FROM order_items oi
		WHERE oi.order_id = ANY($1)`

	orderIDs := make([]string, 0, len(orderMap))
	for id := range orderMap {
		orderIDs = append(orderIDs, id)
	}

	if len(orderIDs) > 0 {
		itemRows, err := r.db.Query(itemsQuery, "{"+strings.Join(orderIDs, ",")+"}")
		if err != nil {
			r.logger.Error("Failed to query order items", "error", err)
			return nil, nil, fmt.Errorf("failed to query order items: %v", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			item := models.OrderItem{}
			var customizations sql.NullString

			err := itemRows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.PriceAtTime, &customizations)
			if err != nil {
				r.logger.Error("Failed to scan order item", "error", err)
				return nil, nil, fmt.Errorf("failed to scan order item: %v", err)
			}

			if customizations.Valid {
				item.Customizations = customizations.String
			}

			item.ProductID = item.MenuItemID

			if order, exists := orderMap[item.OrderID]; exists {
				order.Items = append(order.Items, item)
			}
		}
	}

	menuQuery := `
		SELECT m.id, m.name, m.description, m.category, m.price, 
		       m.available, m.metadata, m.tags, m.allergens, 
		       m.available_sizes, m.created_at, m.updated_at
		FROM menu_items m
		ORDER BY m.name`

	menuRows, err := r.db.Query(menuQuery)
	if err != nil {
		r.logger.Error("Failed to query menu items for aggregation", "error", err)
		return nil, nil, fmt.Errorf("failed to query menu items: %v", err)
	}
	defer menuRows.Close()

	menuMap := make(map[string]*models.MenuItem)
	for menuRows.Next() {
		item := &models.MenuItem{}
		var metadata, tags, allergens, availableSizes sql.NullString

		err := menuRows.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.Price, &item.Available, &metadata, &tags, &allergens, &availableSizes, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			r.logger.Error("Failed to scan menu item", "error", err)
			return nil, nil, fmt.Errorf("failed to scan menu item: %v", err)
		}

		if tags.Valid && tags.String != "" {
			item.Tags = parsePostgreSQLArray(tags.String)
		}
		if allergens.Valid && allergens.String != "" {
			item.Allergens = parsePostgreSQLArray(allergens.String)
		}
		if metadata.Valid {
			item.CustomizationOptions = []byte(metadata.String)
		}

		menuMap[item.ID] = item
		menuItems = append(menuItems, item)
	}

	ingredientsQuery := `
		SELECT mii.menu_item_id, mii.ingredient_id, mii.required_quantity
		FROM menu_item_ingredients mii
		WHERE mii.menu_item_id = ANY($1)`

	menuItemIDs := make([]string, 0, len(menuMap))
	for id := range menuMap {
		menuItemIDs = append(menuItemIDs, id)
	}
	if len(menuItemIDs) > 0 {
		ingredientRows, err := r.db.Query(ingredientsQuery, "{"+strings.Join(menuItemIDs, ",")+"}")
		if err != nil {
			r.logger.Error("Failed to query menu item ingredients", "error", err)
			return nil, nil, fmt.Errorf("failed to query menu item ingredients: %v", err)
		}
		defer ingredientRows.Close()

		for ingredientRows.Next() {
			var menuItemID, ingredientID string
			var quantity float64

			err := ingredientRows.Scan(&menuItemID, &ingredientID, &quantity)
			if err != nil {
				r.logger.Error("Failed to scan menu item ingredient", "error", err)
				return nil, nil, fmt.Errorf("failed to scan menu item ingredient: %v", err)
			}

			if menuItem, exists := menuMap[menuItemID]; exists {
				ingredient := models.MenuItemIngredient{
					IngredientID: ingredientID,
					Quantity:     quantity,
				}
				menuItem.Ingredients = append(menuItem.Ingredients, ingredient)
			}
		}
	}

	r.logger.Info("Aggregation data fetched successfully", "orders_count", len(orders), "menu_items_count", len(menuItems))
	return orders, menuItems, nil
}

func (r *AggregationRepository) SearchFullText(query string, filters []string, minPrice, maxPrice *float64) (*SearchResult, error) {
	r.logger.Info("Performing full text search", "query", query, "filters", filters)

	result := &SearchResult{
		MenuItems: []MenuSearchResult{},
		Orders:    []OrderSearchResult{},
	}

	searchAll := len(filters) == 0 || contains(filters, "all")
	searchMenu := searchAll || contains(filters, "menu")
	searchOrders := searchAll || contains(filters, "orders")

	if searchMenu {
		menuQuery := `
			SELECT m.id, m.name, m.description, m.price,
			       similarity(m.name || ' ' || COALESCE(m.description, ''), $1) as relevance
			FROM menu_items m
			WHERE (m.name ILIKE '%' || $1 || '%' OR m.description ILIKE '%' || $1 || '%')`
		args := []interface{}{query}
		argCount := 1

		if minPrice != nil {
			argCount++
			menuQuery += fmt.Sprintf(" AND m.price >= $%d", argCount)
			args = append(args, *minPrice)
		}
		if maxPrice != nil {
			argCount++
			menuQuery += fmt.Sprintf(" AND m.price <= $%d", argCount)
			args = append(args, *maxPrice)
		}

		menuQuery += " ORDER BY relevance DESC, m.name LIMIT 50"

		menuRows, err := r.db.Query(menuQuery, args...)
		if err != nil {
			r.logger.Error("Failed to search menu items", "error", err)
			return nil, fmt.Errorf("failed to search menu items: %v", err)
		}
		defer menuRows.Close()

		for menuRows.Next() {
			var item MenuSearchResult
			var description sql.NullString

			err := menuRows.Scan(&item.ID, &item.Name, &description, &item.Price, &item.Relevance)
			if err != nil {
				r.logger.Error("Failed to scan menu search result", "error", err)
				continue
			}

			if description.Valid {
				item.Description = description.String
			}

			result.MenuItems = append(result.MenuItems, item)
		}
	}

	if searchOrders {
		orderQuery := `
			SELECT DISTINCT o.id, o.customer_name, o.total_amount,
			       similarity(o.customer_name, $1) as relevance
			FROM orders o
			LEFT JOIN order_items oi ON o.id = oi.order_id
			LEFT JOIN menu_items m ON oi.menu_item_id = m.id
			WHERE (o.customer_name ILIKE '%' || $1 || '%' 
			       OR m.name ILIKE '%' || $1 || '%')`
		args := []interface{}{query}
		argCount := 1

		if minPrice != nil {
			argCount++
			orderQuery += fmt.Sprintf(" AND o.total_amount >= $%d", argCount)
			args = append(args, *minPrice)
		}
		if maxPrice != nil {
			argCount++
			orderQuery += fmt.Sprintf(" AND o.total_amount <= $%d", argCount)
			args = append(args, *maxPrice)
		}

		orderQuery += " ORDER BY relevance DESC, o.created_at DESC LIMIT 50"

		orderRows, err := r.db.Query(orderQuery, args...)
		if err != nil {
			r.logger.Error("Failed to search orders", "error", err)
			return nil, fmt.Errorf("failed to search orders: %v", err)
		}
		defer orderRows.Close()

		for orderRows.Next() {
			var order OrderSearchResult
			err := orderRows.Scan(&order.ID, &order.CustomerName, &order.Total, &order.Relevance)
			if err != nil {
				r.logger.Error("Failed to scan order search result", "error", err)
				continue
			}

			itemsQuery := `
				SELECT m.name
				FROM order_items oi
				JOIN menu_items m ON oi.menu_item_id = m.id
				WHERE oi.order_id = $1`
			itemRows, err := r.db.Query(itemsQuery, order.ID)
			if err != nil {
				r.logger.Error("Failed to get order items", "order_id", order.ID, "error", err)
				continue
			}

			items := []string{}
			for itemRows.Next() {
				itemName := ""
				if err := itemRows.Scan(&itemName); err == nil {
					items = append(items, itemName)
				}
			}
			itemRows.Close()

			order.Items = items
			result.Orders = append(result.Orders, order)
		}
	}

	result.TotalMatches = len(result.MenuItems) + len(result.Orders)

	r.logger.Info("Full text search completed", "total_matches", result.TotalMatches)
	return result, nil
}

func (r *AggregationRepository) GetOrderedItemsByPeriod(period, month, year string) (*OrderedItemsByPeriodResult, error) {
	r.logger.Info("Getting ordered items by period", "period", period, "month", month, "year", year)

	result := &OrderedItemsByPeriodResult{
		Period:       period,
		OrderedItems: []map[string]int{},
	}
	if period == "day" {
		result.Month = month
		return r.getOrderedItemsByDay(month, result)
	} else if period == "month" {
		result.Year = year
		return r.getOrderedItemsByMonth(year, result)
	}

	return nil, fmt.Errorf("invalid period: %s", period)
}

func (r *AggregationRepository) getOrderedItemsByDay(month string, result *OrderedItemsByPeriodResult) (*OrderedItemsByPeriodResult, error) {
	currentYear := time.Now().Year()

	monthNum, err := parseMonth(month)
	if err != nil {
		return nil, fmt.Errorf("invalid month: %s", month)
	}

	query := `
		SELECT 
			EXTRACT(DAY FROM o.created_at) as day,
			COUNT(DISTINCT o.id) as order_count
		FROM orders o
		WHERE EXTRACT(MONTH FROM o.created_at) = $1 
		  AND EXTRACT(YEAR FROM o.created_at) = $2
		  AND o.status = 'closed'
		GROUP BY EXTRACT(DAY FROM o.created_at)
		ORDER BY day`

	rows, err := r.db.Query(query, monthNum, currentYear)
	if err != nil {
		r.logger.Error("Failed to get ordered items by day", "error", err)
		return nil, fmt.Errorf("failed to get ordered items by day: %v", err)
	}
	defer rows.Close()

	dayMap := make(map[int]int)
	for rows.Next() {
		var day, orderCount int
		err := rows.Scan(&day, &orderCount)
		if err != nil {
			r.logger.Error("Failed to scan day result", "error", err)
			continue
		}
		dayMap[day] = orderCount
	}

	daysInMonth := getDaysInMonth(monthNum, currentYear)

	for day := 1; day <= daysInMonth; day++ {
		dayData := make(map[string]int)
		dayData[fmt.Sprintf("%d", day)] = dayMap[day]
		result.OrderedItems = append(result.OrderedItems, dayData)
	}

	return result, nil
}

func (r *AggregationRepository) getOrderedItemsByMonth(year string, result *OrderedItemsByPeriodResult) (*OrderedItemsByPeriodResult, error) {
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	query := `
		SELECT 
			EXTRACT(MONTH FROM o.created_at) as month,
			COUNT(DISTINCT o.id) as order_count
		FROM orders o
		WHERE EXTRACT(YEAR FROM o.created_at) = $1
		  AND o.status = 'closed'
		GROUP BY EXTRACT(MONTH FROM o.created_at)
		ORDER BY month`

	rows, err := r.db.Query(query, year)
	if err != nil {
		r.logger.Error("Failed to get ordered items by month", "error", err)
		return nil, fmt.Errorf("failed to get ordered items by month: %v", err)
	}
	defer rows.Close()

	monthMap := make(map[int]int)
	for rows.Next() {
		var month, orderCount int
		err := rows.Scan(&month, &orderCount)
		if err != nil {
			r.logger.Error("Failed to scan month result", "error", err)
			continue
		}
		monthMap[month] = orderCount
	}

	monthNames := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}

	for i, monthName := range monthNames {
		monthData := make(map[string]int)
		monthData[monthName] = monthMap[i+1]
		result.OrderedItems = append(result.OrderedItems, monthData)
	}

	return result, nil
}

func parsePostgreSQLArray(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseMonth(month string) (int, error) {
	monthMap := map[string]int{
		"january": 1, "february": 2, "march": 3, "april": 4,
		"may": 5, "june": 6, "july": 7, "august": 8,
		"september": 9, "october": 10, "november": 11, "december": 12,
	}

	if num, ok := monthMap[strings.ToLower(month)]; ok {
		return num, nil
	}
	return 0, fmt.Errorf("invalid month: %s", month)
}

func getDaysInMonth(month, year int) int {
	t := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}
