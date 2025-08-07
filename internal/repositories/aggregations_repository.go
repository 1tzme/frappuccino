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
	"frappuccino/models"
	"frappuccino/pkg/database"
	"frappuccino/pkg/logger"
	"strings"
)

type AggregationRepositoryInterface interface {
	GetAggregationData() (orders []*models.Order, menuItems []*models.MenuItem, err error)
}

type AggregationRepository struct {
	db     *database.DB
	logger *logger.Logger
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

func parsePostgreSQLArray(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}