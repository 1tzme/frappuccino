package repositories

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: This entire file-based repository implementation should be replaced
// with PostgreSQL-backed repository. Key changes needed:
// 1. Replace map[string]*models.Order with database connection
// 2. Replace JSON file operations with SQL queries (SELECT, INSERT, UPDATE, DELETE)
// 3. Remove file I/O operations (loadFromFile, saveToFile, backupFile)
// 4. Replace sync.RWMutex with database transaction handling
// 5. Convert dataFilePath to database connection dependency
// 6. Implement proper SQL schema for orders table with relationships

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"frappuccino/models"
	"frappuccino/pkg/database"
	"frappuccino/pkg/logger"
)

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace with PostgreSQL-backed repository interface
// Interface should remain the same but implementation will change from JSON files to SQL operations
type OrderRepositoryInterface interface {
	GetAll() ([]*models.Order, error)
	GetByID(id string) (*models.Order, error)
	Add(order *models.Order) error
	Update(id string, order *models.Order) error
	Delete(id string) error
	CloseOrder(id string) error
	GetNumberOfOrderedItems(startDate, endDate *time.Time) (map[string]int, error)
	BatchProcessOrders(orders []*models.Order) ([]*models.Order, error)
	GetInventoryRequirements(orders []*models.Order) (map[string]float64, error)
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// Signature: NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository
// Temporarily keeping file operations while transitioning to database
type OrderRepository struct {
	logger *logger.Logger
	db     *database.DB
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// New signature: NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository
// Temporarily falls back to in-memory storage during transition
func NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository {
	return &OrderRepository{
		logger: logger.WithComponent("order_repository"),
		db:     db,
	}
}

// Add adds a new order
func (r *OrderRepository) Add(order *models.Order) error {
	r.logger.Debug("Adding new order to database", "customer_name", order.CustomerName)

	err := r.validateOrder(order)
	if err != nil {
		r.logger.Error("Failed to validate order", "error", err, "order_id", order.ID)
		return fmt.Errorf("failed to validate order: %v", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			r.logger.Warn("Rolling back order creation transaction due to error", "error", err, "customer_name", order.CustomerName)
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO orders (customer_name, status, total_amount, special_instructions)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	generatedID := ""
	var createdAt, updatedAt time.Time

	err = tx.QueryRow(query, order.CustomerName, order.Status, order.TotalAmount, order.SpecialInstructions).Scan(&generatedID, &createdAt, &updatedAt)
	if err != nil {
		r.logger.Error("Failed to insert order", "error", err, "customer_name", order.CustomerName)
		return fmt.Errorf("failed to insert order: %v", err)
	}

	order.ID = generatedID
	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt

	if len(order.Items) > 0 {
		itemQuery := `
			INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`

		for i, item := range order.Items {
			itemID := ""
			err := tx.QueryRow(itemQuery, order.ID, item.MenuItemID, item.Quantity, item.PriceAtTime, item.Customizations).Scan(&itemID)
			if err != nil {
				r.logger.Error("Failed to insert order item", "error", err, "order_id", order.ID, "menu_item_id", item.MenuItemID)
				return fmt.Errorf("failed to insert order item: %v", err)
			}
			order.Items[i].ID = itemID
			order.Items[i].OrderID = order.ID
		}
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err, "order_id", order.ID)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("Successfully committed order creation transaction", "order_id", order.ID, "customer_name", order.CustomerName, "items_count", len(order.Items))
	r.logger.Info("Added new order", "order_id", order.ID, "customer_name", order.CustomerName)
	return nil
}

// GetByID retrieves a single order by ID
func (r *OrderRepository) GetByID(id string) (*models.Order, error) {
	r.logger.Debug("Retrieving order from database", "order_id", id)

	query := `
		SELECT id, customer_name, status, total_amount, special_instructions, created_at, updated_at
		FROM orders
		WHERE id = $1`

	order := &models.Order{}
	var specialInstructions string
	err := r.db.QueryRow(query, id).Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructions, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		r.logger.Error("Failed to retrieve order", "error", err, "order_id", id)
		return nil, fmt.Errorf("failed to retrieve order: %v", err)
	}

	itemsQuery := `
		SELECT id, menu_item_id, quantity, price_at_time, customizations
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`

	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		r.logger.Error("Failed to query order items", "error", err, "order_id", id)
		return nil, fmt.Errorf("failed to query order items: %v", err)
	}
	defer rows.Close()

	items := []models.OrderItem{}
	for rows.Next() {
		item := models.OrderItem{OrderID: id}
		customizations := ""
		err := rows.Scan(&item.ID, &item.MenuItemID, &item.Quantity, &item.PriceAtTime, &customizations)
		if err != nil {
			r.logger.Error("Failed to scan order item", "error", err, "order_id", id)
			return nil, fmt.Errorf("failed to scan order item: %v", err)
		}
		item.ProductID = item.MenuItemID
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		r.logger.Error("Error iterating order items", "error", err, "order_id", id)
		return nil, fmt.Errorf("error iterating order items: %v", err)
	}

	order.Items = items
	r.logger.Debug("Retrieved order with items", "order_id", id, "items_count", len(items))
	return order, nil
}

// GetAll retrieves all orders
func (r *OrderRepository) GetAll() ([]*models.Order, error) {
	r.logger.Debug("Retrieving all orders from database")

	query := `
		SELECT id, customer_name, status, total_amount, special_instructions, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Failed to query orders", "error", err)
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	orders := []*models.Order{}
	orderMap := make(map[string]*models.Order)

	for rows.Next() {
		order := &models.Order{}
		var specialInstructions string
		err := rows.Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructions, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			r.logger.Error("Failed to scan order", "error", err)
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}
		order.Items = []models.OrderItem{}
		orders = append(orders, order)
		orderMap[order.ID] = order
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating orders", "error", err)
		return nil, fmt.Errorf("error iterating orders: %v", err)
	}

	if len(orders) > 0 {
		itemsQuery := `
			SELECT order_id, id, menu_item_id, quantity, price_at_time, customizations
			FROM order_items
			WHERE order_id = ANY($1)
			ORDER BY order_id, id`

		orderIDs := make([]string, len(orders))
		for i, order := range orders {
			orderIDs[i] = order.ID
		}

		itemRows, err := r.db.Query(itemsQuery, "{"+strings.Join(orderIDs, ",")+"}")
		if err != nil {
			r.logger.Error("Failed to query order items", "error", err)
			return nil, fmt.Errorf("failed to query order items: %v", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			item := models.OrderItem{}
			var customizations string
			err := itemRows.Scan(&item.OrderID, &item.ID, &item.MenuItemID, &item.Quantity, &item.PriceAtTime, &customizations)
			if err != nil {
				r.logger.Error("Failed to scan order item", "error", err)
				return nil, fmt.Errorf("failed to scan order item: %v", err)
			}
			item.ProductID = item.MenuItemID

			if order, exists := orderMap[item.OrderID]; exists {
				order.Items = append(order.Items, item)
			}
		}

		if err = itemRows.Err(); err != nil {
			r.logger.Error("Error iterating order items", "error", err)
			return nil, fmt.Errorf("error iterating order items: %v", err)
		}
	}

	r.logger.Info("Retrieved all orders", "count", len(orders))
	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(id string, order *models.Order) error {
	r.logger.Debug("Updating order in database", "order_id", id)

	if err := r.validateOrderForUpdate(order, id); err != nil {
		r.logger.Error("Failed to validate order", "error", err, "order_id", id)
		return fmt.Errorf("invalid order: %v", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			r.logger.Warn("Rolling back order update transaction due to error", "error", err, "order_id", id)
			tx.Rollback()
		}
	}()

	query := `
		UPDATE orders
		SET customer_name = $1, status = $2, total_amount = $3, special_instructions = $4
		WHERE id = $5`

	result, err := tx.Exec(query, order.CustomerName, order.Status, order.TotalAmount, "{}", id)
	if err != nil {
		r.logger.Error("Failed to update order", "error", err, "order_id", id)
		return fmt.Errorf("failed to update order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "order_id", id)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Attempted to update non-existent order", "order_id", id)
		return fmt.Errorf("order with id %s not found", id)
	}

	deleteItemsQuery := `DELETE FROM order_items WHERE order_id = $1`
	_, err = tx.Exec(deleteItemsQuery, id)
	if err != nil {
		r.logger.Error("Failed to delete existing order items", "error", err, "order_id", id)
		return fmt.Errorf("failed to delete existing order items: %v", err)
	}

	if len(order.Items) > 0 {
		itemQuery := `
			INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
			VALUES ($1, $2, $3, $4, $5)`

		for _, item := range order.Items {
			_, err = tx.Exec(itemQuery, id, item.MenuItemID, item.Quantity, item.PriceAtTime, "{}")
			if err != nil {
				r.logger.Error("Failed to insert updated order item", "error", err, "order_id", id, "menu_item_id", item.MenuItemID)
				return fmt.Errorf("failed to insert updated order item: %v", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", "error", err, "order_id", id)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("Successfully committed order update transaction", "order_id", id, "customer_name", order.CustomerName, "items_count", len(order.Items))
	r.logger.Info("Updated order", "order_id", id, "customer_name", order.CustomerName)
	return nil
}

// Delete removes an order by ID
func (r *OrderRepository) Delete(id string) error {
	r.logger.Debug("Deleting order from database", "order_id", id)

	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Error("Failed to delete order", "error", err, "order_id", id)
		return fmt.Errorf("failed to delete order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "order_id", id)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Attempted to delete non-existent order", "order_id", id)
		return fmt.Errorf("order with id %s not found", id)
	}

	r.logger.Info("Deleted order", "order_id", id)
	return nil
}

// CloseOrder closes an order by setting status to closed
func (r *OrderRepository) CloseOrder(id string) error {
	r.logger.Debug("Closing order in database", "order_id", id)

	query := `
		UPDATE orders
		SET status = 'closed'
		WHERE id = $1 AND status != 'closed'`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Error("Failed to close order", "error", err, "order_id", id)
		return fmt.Errorf("failed to close order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "order_id", id)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Attempted to close non-existent or already closed order", "order_id", id)
		return fmt.Errorf("order with id %s not found or already closed", id)
	}

	r.logger.Info("Closed order", "order_id", id)
	return nil
}

// GetNumberOfOrderedItems retrieves number of ordered items count by date interval
func (r *OrderRepository) GetNumberOfOrderedItems(startDate, endDate *time.Time) (map[string]int, error) {
	r.logger.Debug("Retrieving number of ordered items", "startDate", startDate, "endDate", endDate)

	query := `SELECT mi.name, SUM(oi.quantity) as total_quantity
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE ($1::timestamp IS NULL OR o.created_at >= $1)
		AND ($2::timestamp IS NULL OR o.created_at <= $2)
		GROUP BY mi.name
		ORDER BY mi.name`

	var startDateParam, endDateParam interface{}
	if startDate != nil {
		startDateParam = *startDate
	}
	if endDate != nil {
		endDateParam = *endDate
	}

	rows, err := r.db.Query(query, startDateParam, endDateParam)
	if err != nil {
		r.logger.Error("Failed to query ordered items", "error", err)
		return nil, fmt.Errorf("failed to query ordered items: %v", err)
	}
	defer rows.Close()

	result := make(map[string]int)

	for rows.Next() {
		itemName := ""
		quantity := 0
		err := rows.Scan(&itemName, &quantity)
		if err != nil {
			r.logger.Error("Failed to scan ordered item", "error", err)
			return nil, fmt.Errorf("failed to scan ordered item: %v", err)
		}
		result[itemName] = quantity
	}

	err = rows.Err()
	if err != nil {
		r.logger.Error("Error iterating ordered items", "error", err)
		return nil, fmt.Errorf("error iterating ordered items: %v", err)
	}

	r.logger.Info("Retrieved ordered items", "count", len(result))
	return result, nil
}

// BatchProcessOrders processes multiple orders in a single transaction
func (r *OrderRepository) BatchProcessOrders(orders []*models.Order) ([]*models.Order, error) {
	r.logger.Debug("Batch processing orders", "count", len(orders))

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin batch transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	processedOrders := make([]*models.Order, len(orders))

	for i, order := range orders {
		if err := r.validateOrder(order); err != nil {
			r.logger.Error("Failed to validate order in batch", "error", err, "customer", order.CustomerName)
			return nil, fmt.Errorf("order %d validation failed: %v", i, err)
		}

		query := `
			INSERT INTO orders (customer_name, status, total_amount, special_instructions)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at`

		var generatedID string
		var createdAt, updatedAt time.Time

		err = tx.QueryRow(query, order.CustomerName, order.Status, order.TotalAmount, order.SpecialInstructions).Scan(&generatedID, &createdAt, &updatedAt)
		if err != nil {
			r.logger.Error("Failed to insert order in batch", "error", err, "customer", order.CustomerName)
			return nil, fmt.Errorf("failed to insert order %d: %v", i, err)
		}

		order.ID = generatedID
		order.CreatedAt = createdAt
		order.UpdatedAt = updatedAt

		if len(order.Items) > 0 {
			itemQuery := `
				INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id`

			for j, item := range order.Items {
				itemID := ""
				err := tx.QueryRow(itemQuery, order.ID, item.MenuItemID, item.Quantity, item.PriceAtTime, item.Customizations).Scan(&itemID)
				if err != nil {
					r.logger.Error("Failed to insert order item in batch", "error", err, "order_id", order.ID, "menu_item_id", item.MenuItemID)
					return nil, fmt.Errorf("failed to insert order item for order %d: %v", i, err)
				}
				order.Items[j].ID = itemID
				order.Items[j].OrderID = order.ID
			}
		}

		processedOrders[i] = order
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit batch transaction", "error", err)
		return nil, fmt.Errorf("failed to commit batch transaction: %v", err)
	}

	r.logger.Info("Batch processed orders", "count", len(processedOrders))
	return processedOrders, nil
}

// GetInventoryRequirements calculates total ingredient requirements for multiple orders
func (r *OrderRepository) GetInventoryRequirements(orders []*models.Order) (map[string]float64, error) {
	r.logger.Debug("Calculating inventory requirements for batch orders", "count", len(orders))

	requirements := make(map[string]float64)

	for _, order := range orders {
		for _, item := range order.Items {
			query := `
				SELECT mi.ingredient_id, mi.required_quantity
				FROM menu_item_ingredients mi
				WHERE mi.menu_item_id = $1`

			rows, err := r.db.Query(query, item.MenuItemID)
			if err != nil {
				r.logger.Error("Failed to query ingredient requirements", "error", err, "menu_item_id", item.MenuItemID)
				return nil, fmt.Errorf("failed to query ingredient requirements: %v", err)
			}
			defer rows.Close()

			for rows.Next() {
				var ingredientID string
				var quantity float64

				err := rows.Scan(&ingredientID, &quantity)
				if err != nil {
					r.logger.Error("Failed to scan ingredient requirement", "error", err)
					return nil, fmt.Errorf("failed to scan ingredient requirement: %v", err)
				}

				totalNeeded := quantity * float64(item.Quantity)
				requirements[ingredientID] += totalNeeded
			}

			if err = rows.Err(); err != nil {
				r.logger.Error("Error iterating ingredient requirements", "error", err)
				return nil, fmt.Errorf("error iterating ingredient requirements: %v", err)
			}
		}
	}

	r.logger.Info("Calculated inventory requirements", "ingredient_count", len(requirements))
	return requirements, nil
}

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: All file operations below should be removed and replaced with SQL queries
// - validateOrder() → Database constraints and triggers

// validateOrder validates order data
func (r *OrderRepository) validateOrder(order *models.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	}
	if order.Status == "" {
		order.Status = "Pending"
	}

	for i, item := range order.Items {
		if item.MenuItemID == "" && item.ProductID == "" {
			return fmt.Errorf("item %d: menu item ID cannot be empty", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i)
		}
		if item.PriceAtTime < 0 {
			return fmt.Errorf("item %d: price cannot be negative", i)
		}

		if item.MenuItemID == "" && item.ProductID != "" {
			order.Items[i].MenuItemID = item.ProductID
		}
	}

	return nil
}

func (r *OrderRepository) validateOrderForUpdate(order *models.Order, id string) error {
	if id == "" {
		return errors.New("order ID cannot be empty")
	}
	return r.validateOrder(order)
}
