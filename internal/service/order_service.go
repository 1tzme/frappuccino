package service

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update order business logic for PostgreSQL integration
// 1. Implement database transaction management for multi-table order operations
// 2. Replace in-memory validation with database constraint enforcement
// 3. Add proper handling of database foreign key relationships
// 4. Update order processing to use database triggers and stored procedures
// 5. Implement database-based inventory tracking and order fulfillment

import (
	"fmt"
	"time"

	"frappuccino/internal/repositories"
	"frappuccino/models"
	"frappuccino/pkg/logger"
)

// Define request/response structs
type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name"`
	Items        []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateOrderRequest struct {
	CustomerName string                   `json:"customer_name"`
	Items        []CreateOrderItemRequest `json:"items"`
	Status       string                   `json:"status"`
}

// OrderService interface
type OrderServiceInterface interface {
	CreateOrder(req CreateOrderRequest) (*models.Order, error)
	GetAllOrders() ([]*models.Order, error)
	GetOrderByID(id string) (*models.Order, error)
	UpdateOrder(id string, req UpdateOrderRequest) error
	DeleteOrder(id string) error
	CloseOrder(id string) error
	GetNumberOfOrderedItems(startDate, endDate string) (map[string]int, error)
	BatchProcessOrders(req models.BatchOrderRequest) (*models.BatchProcessResponse, error)
}

// OrderService struct
type OrderService struct {
	orderRepo     repositories.OrderRepositoryInterface
	menuRepo      repositories.MenuRepositoryInterface
	inventoryRepo repositories.InventoryRepositoryInterface
	logger        *logger.Logger
}

// NewOrderService creates a new OrderService with the given repositories and logger
func NewOrderService(orderRepo repositories.OrderRepositoryInterface, menuRepo repositories.MenuRepositoryInterface, inventoryRepo repositories.InventoryRepositoryInterface, logger *logger.Logger) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger.WithComponent("order_service"),
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(req CreateOrderRequest) (*models.Order, error) {
	s.logger.Info("Creating new order", "customer", req.CustomerName)

	if err := s.validateOrderData(req); err != nil {
		s.logger.Warn("Create failed: invalid data", "error", err)
		return nil, err
	}

	// Check inventory availability before creating order
	if err := s.checkInventoryAvailability(req.Items); err != nil {
		s.logger.Warn("Create failed: insufficient inventory", "error", err)
		return nil, err
	}

	totalAmount, err := s.calculateOrderTotal(req.Items)
	if err != nil {
		s.logger.Error("Failed to calculate order total", "error", err)
		return nil, fmt.Errorf("failed to calculate order total: %v", err)
	}

	order := &models.Order{
		CustomerName: req.CustomerName,
		Status:       "pending",
		TotalAmount:  totalAmount,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Items:        make([]models.OrderItem, len(req.Items)),
	}

	for i, item := range req.Items {
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			s.logger.Error("Failed to get menu item", "product_id", item.ProductID, "error", err)
			return nil, fmt.Errorf("menu item %s not found", item.ProductID)
		}

		order.Items[i] = models.OrderItem{
			MenuItemID:  item.ProductID,
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			PriceAtTime: menuItem.Price,
		}
	}

	// Consume inventory for the order
	if err := s.consumeInventory(req.Items); err != nil {
		s.logger.Error("Failed to consume inventory", "error", err)
		return nil, err
	}

	if err := s.orderRepo.Add(order); err != nil {
		// Rollback inventory consumption if order creation fails
		s.logger.Error("Failed to add order, rolling back inventory", "error", err)
		s.restoreInventory(req.Items)
		return nil, err
	}

	s.logger.Info("Order created", "order_id", order.ID, "total_amount", totalAmount)
	return order, nil
}

// GetAllOrders retrieves all orders
func (s *OrderService) GetAllOrders() ([]*models.Order, error) {
	s.logger.Info("Fetching all orders from repository")

	orders, err := s.orderRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to fetch orders from repository", "error", err)
		return nil, err
	}

	s.logger.Info("Fetched orders", "count", len(orders))
	return orders, nil
}

// GetOrderByID retrieves a specific order by ID
func (s *OrderService) GetOrderByID(id string) (*models.Order, error) {
	s.logger.Info("Fetching order by ID", "order_id", id)

	if id == "" {
		s.logger.Warn("Order ID cannot be empty")
		return nil, fmt.Errorf("order ID is required")
	}

	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Order not found", "order_id", id, "error", err)
		return nil, err
	}

	s.logger.Info("Fetched order", "order_id", id)
	return order, nil
}

// UpdateOrder updates an existing order
func (s *OrderService) UpdateOrder(id string, req UpdateOrderRequest) error {
	s.logger.Info("Updating order", "order_id", id, "customer", req.CustomerName)

	if id == "" {
		s.logger.Warn("Order ID cannot be empty")
		return fmt.Errorf("order ID is required")
	}

	if err := s.validateUpdateOrderData(req); err != nil {
		s.logger.Warn("Update failed: invalid data", "order_id", id, "error", err)
		return err
	}

	// Get the existing order to restore its inventory first
	existingOrder, err := s.orderRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Order not found for update", "order_id", id, "error", err)
		return err
	}

	if existingOrder.Status == "closed" {
		s.logger.Warn("Attempted to update a closed order", "order_id", id)
		return fmt.Errorf("cannot update closed order")
	}

	// Convert existing order items to request format for inventory restoration
	existingItems := make([]CreateOrderItemRequest, len(existingOrder.Items))
	for i, item := range existingOrder.Items {
		existingItems[i] = CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Restore inventory from the existing order
	if err := s.restoreInventory(existingItems); err != nil {
		s.logger.Error("Failed to restore inventory from existing order", "order_id", id, "error", err)
		return err
	}

	// Check inventory availability for the new order items
	if err := s.checkInventoryAvailability(req.Items); err != nil {
		s.logger.Warn("Update failed: insufficient inventory", "order_id", id, "error", err)
		// Re-consume the original inventory since update failed
		s.consumeInventory(existingItems)
		return err
	}

	// Consume inventory for the new order items
	if err := s.consumeInventory(req.Items); err != nil {
		s.logger.Error("Failed to consume inventory for updated order", "order_id", id, "error", err)
		// Re-consume the original inventory since update failed
		s.consumeInventory(existingItems)
		return err
	}

	order := &models.Order{
		ID:           id,
		CustomerName: req.CustomerName,
		Items:        make([]models.OrderItem, len(req.Items)),
		Status:       req.Status,
		CreatedAt:    existingOrder.CreatedAt, // Preserve original creation time
	}

	// Convert request items to order items
	for i, item := range req.Items {
		order.Items[i] = models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	if err := s.orderRepo.Update(id, order); err != nil {
		s.logger.Error("Failed to update order in repository", "order_id", id, "error", err)
		// Rollback inventory changes
		s.restoreInventory(req.Items)
		s.consumeInventory(existingItems)
		return err
	}

	s.logger.Info("Order updated with inventory management", "order_id", id)
	return nil
}

// DeleteOrder deletes an order by ID
func (s *OrderService) DeleteOrder(id string) error {
	s.logger.Info("Deleting order", "order_id", id)

	if id == "" {
		s.logger.Warn("Order ID cannot be empty")
		return fmt.Errorf("order ID is required")
	}

	// Get the order before deleting to restore inventory
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Order not found for deletion", "order_id", id, "error", err)
		return err
	}

	// Convert order items to request format for inventory restoration
	items := make([]CreateOrderItemRequest, len(order.Items))
	for i, item := range order.Items {
		items[i] = CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Restore inventory before deleting order
	if err := s.restoreInventory(items); err != nil {
		s.logger.Error("Failed to restore inventory", "order_id", id, "error", err)
		return err
	}

	if err := s.orderRepo.Delete(id); err != nil {
		s.logger.Warn("Failed to delete order", "order_id", id, "error", err)
		// Try to re-consume inventory if delete fails
		s.consumeInventory(items)
		return err
	}

	s.logger.Info("Order deleted and inventory restored", "order_id", id)
	return nil
}

// CloseOrder closes an order by setting status to closed
func (s *OrderService) CloseOrder(id string) error {
	s.logger.Info("Closing order", "order_id", id)

	if id == "" {
		s.logger.Warn("Order ID cannot be empty")
		return fmt.Errorf("order ID is required")
	}

	if err := s.orderRepo.CloseOrder(id); err != nil {
		s.logger.Warn("Failed to close order", "order_id", id, "error", err)
		return err
	}

	s.logger.Info("Order closed", "order_id", id)
	return nil
}

// GetNumberOfOrderedItems returns count of ordered items in date interval
func (s *OrderService) GetNumberOfOrderedItems(startDate, endDate string) (map[string]int, error) {
	s.logger.Info("Getting number of ordered items", "startDate", startDate, "endDate", endDate)

	var parsedStartDate, parsedEndDate *time.Time

	if startDate != "" {
		parsedDate, err := s.parseDate(startDate)
		if err != nil {
			s.logger.Error("Invalid start date format", "startDate", startDate, "error", err)
			return nil, fmt.Errorf("invalid start date format: %v", err)
		}
		parsedStartDate = &parsedDate
	}

	if endDate != "" {
		parsedDate, err := s.parseDate(endDate)
		if err != nil {
			s.logger.Warn("Invalid end date format", "endDate", endDate, "error", err)
			return nil, fmt.Errorf("invalid end date format: %v", err)
		}
		endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, parsedDate.Location())
		parsedEndDate = &endOfDay
	}

	if parsedStartDate != nil && parsedEndDate != nil {
		if parsedStartDate.After(*parsedEndDate) {
			s.logger.Warn("Start date is after end date", "startDate", startDate, "endDate", endDate)
			return nil, fmt.Errorf("start date cannot be after end date")
		}
	}

	result, err := s.orderRepo.GetNumberOfOrderedItems(parsedStartDate, parsedEndDate)
	if err != nil {
		s.logger.Error("Failed to get ordered items from repository", "error", err)
		return nil, fmt.Errorf("failed to get ordered items: %v", err)
	}

	s.logger.Info("Retrieved ordered items count", "items_count", len(result))
	return result, nil
}

// BatchProcessOrders processes multiple orders simultaneously with inventory consistency
func (s *OrderService) BatchProcessOrders(req models.BatchOrderRequest) (*models.BatchProcessResponse, error) {
	s.logger.Info("Starting batch order processing", "order_count", len(req.Orders))

	if len(req.Orders) == 0 {
		return nil, fmt.Errorf("no orders provided for processing")
	}

	// Validate all orders first
	orders := make([]*models.Order, 0, len(req.Orders))
	totalRevenue := 0.0

	for i, orderReq := range req.Orders {
		if err := s.validateBatchOrderRequest(orderReq, i); err != nil {
			s.logger.Warn("Batch order validation failed", "order_index", i, "error", err)
			return nil, fmt.Errorf("order %d validation failed: %v", i, err)
		}

		// Calculate order total and create Order struct
		orderTotal, err := s.calculateBatchOrderTotal(orderReq.Items)
		if err != nil {
			s.logger.Error("Failed to calculate order total in batch", "order_index", i, "error", err)
			return nil, fmt.Errorf("order %d total calculation failed: %v", i, err)
		}

		order := &models.Order{
			CustomerName: orderReq.CustomerName,
			Status:       "pending",
			TotalAmount:  orderTotal,
			Items:        make([]models.OrderItem, len(orderReq.Items)),
		}

		// Convert items
		for j, item := range orderReq.Items {
			menuItem, err := s.menuRepo.GetByID(item.MenuItemID)
			if err != nil {
				s.logger.Error("Menu item not found in batch processing", "menu_item_id", item.MenuItemID, "error", err)
				return nil, fmt.Errorf("order %d: menu item %s not found", i, item.MenuItemID)
			}

			order.Items[j] = models.OrderItem{
				MenuItemID:  item.MenuItemID,
				ProductID:   item.MenuItemID,
				Quantity:    item.Quantity,
				PriceAtTime: menuItem.Price,
			}
		}

		orders = append(orders, order)
		totalRevenue += orderTotal
	}

	// Check inventory availability for all orders
	inventoryRequirements, err := s.orderRepo.GetInventoryRequirements(orders)
	if err != nil {
		s.logger.Error("Failed to calculate inventory requirements", "error", err)
		return nil, fmt.Errorf("failed to calculate inventory requirements: %v", err)
	}

	// Check if we have sufficient inventory
	_, err = s.inventoryRepo.CheckInventoryAvailability(inventoryRequirements)
	if err != nil {
		s.logger.Warn("Insufficient inventory for batch orders", "error", err)

		// Return response with all orders rejected due to insufficient inventory
		response := &models.BatchProcessResponse{
			ProcessedOrders: make([]models.BatchProcessResult, len(orders)),
			Summary: models.BatchProcessSummary{
				TotalOrders:      len(orders),
				Accepted:         0,
				Rejected:         len(orders),
				TotalRevenue:     0,
				InventoryUpdates: []models.InventoryUpdateResult{},
			},
		}

		for i, order := range orders {
			response.ProcessedOrders[i] = models.BatchProcessResult{
				OrderID:      "",
				CustomerName: order.CustomerName,
				Status:       "rejected",
				Reason:       "insufficient_inventory",
			}
		}

		return response, nil
	}

	// Process orders in the repository (with transaction)
	processedOrders, err := s.orderRepo.BatchProcessOrders(orders)
	if err != nil {
		s.logger.Error("Failed to batch process orders in repository", "error", err)

		// Return response with all orders rejected due to processing failure
		response := &models.BatchProcessResponse{
			ProcessedOrders: make([]models.BatchProcessResult, len(orders)),
			Summary: models.BatchProcessSummary{
				TotalOrders:      len(orders),
				Accepted:         0,
				Rejected:         len(orders),
				TotalRevenue:     0,
				InventoryUpdates: []models.InventoryUpdateResult{},
			},
		}

		for i, order := range orders {
			response.ProcessedOrders[i] = models.BatchProcessResult{
				OrderID:      "",
				CustomerName: order.CustomerName,
				Status:       "rejected",
				Reason:       "processing_error",
			}
		}

		return response, nil
	}

	// Update inventory
	inventoryUpdates, err := s.inventoryRepo.BatchUpdateInventory(inventoryRequirements)
	if err != nil {
		s.logger.Error("Failed to update inventory in batch", "error", err)
		// Note: Orders were already created, this is a partial failure scenario
		// In production, you might want to implement compensation logic here
	}

	// Build response
	response := &models.BatchProcessResponse{
		ProcessedOrders: make([]models.BatchProcessResult, len(processedOrders)),
		Summary: models.BatchProcessSummary{
			TotalOrders:      len(processedOrders),
			Accepted:         len(processedOrders),
			Rejected:         0,
			TotalRevenue:     totalRevenue,
			InventoryUpdates: inventoryUpdates,
		},
	}

	for i, order := range processedOrders {
		response.ProcessedOrders[i] = models.BatchProcessResult{
			OrderID:      order.ID,
			CustomerName: order.CustomerName,
			Status:       "accepted",
			Total:        order.TotalAmount,
		}
	}

	s.logger.Info("Completed batch order processing",
		"total_orders", len(processedOrders),
		"accepted", len(processedOrders),
		"total_revenue", totalRevenue)

	return response, nil
}

// Private business logic methods

// validateOrderData validates the data for order creation
func (s *OrderService) validateOrderData(req CreateOrderRequest) error {
	if req.CustomerName == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	return s.validateOrderItems(req.Items)
}

// validateUpdateOrderData validates the data for order updates
func (s *OrderService) validateUpdateOrderData(req UpdateOrderRequest) error {
	if req.CustomerName == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate status values
	validStatuses := []string{"open", "closed"}
	statusValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	return s.validateOrderItems(req.Items)
}

// validateOrderItems validates individual order items
func (s *OrderService) validateOrderItems(items []CreateOrderItemRequest) error {
	for i, item := range items {
		if item.ProductID == "" {
			return fmt.Errorf("item %d: product ID is required", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i+1)
		}

		if _, err := s.menuRepo.GetByID(item.ProductID); err != nil {
			return fmt.Errorf("item %d: product '%s' not found in menu", i+1, item.ProductID)
		}
	}
	return nil
}

// calculateOrderTotal calculates the total amount for an order
func (s *OrderService) calculateOrderTotal(items []CreateOrderItemRequest) (float64, error) {
	var total float64
	for i, item := range items {
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return 0, fmt.Errorf("item %d: product '%s' not found in menu", i+1, item.ProductID)
		}
		total += menuItem.Price * float64(item.Quantity)
	}
	return total, nil
}

// checkInventoryAvailability checks if there's enough inventory for all order items
func (s *OrderService) checkInventoryAvailability(items []CreateOrderItemRequest) error {
	for i, item := range items {
		// Get the menu item to find its ingredients
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("item %d: product '%s' not found in menu", i+1, item.ProductID)
		}

		// Check each ingredient's availability
		for _, ingredient := range menuItem.Ingredients {
			inventoryItem, err := s.inventoryRepo.GetByID(ingredient.IngredientID)
			if err != nil {
				return fmt.Errorf("item %d: ingredient '%s' not found in inventory", i+1, ingredient.IngredientID)
			}

			// Calculate total needed quantity for this ingredient
			totalNeeded := ingredient.Quantity * float64(item.Quantity)

			if inventoryItem.Quantity < totalNeeded {
				return fmt.Errorf("item %d: insufficient inventory for ingredient '%s' (need %.2f, have %.2f)",
					i+1, ingredient.IngredientID, totalNeeded, inventoryItem.Quantity)
			}
		}
	}
	return nil
}

// consumeInventory reduces inventory quantities based on order items
func (s *OrderService) consumeInventory(items []CreateOrderItemRequest) error {
	for i, item := range items {
		// Get the menu item to find its ingredients
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("item %d: product '%s' not found in menu", i+1, item.ProductID)
		}

		// Consume each ingredient
		for _, ingredient := range menuItem.Ingredients {
			inventoryItem, err := s.inventoryRepo.GetByID(ingredient.IngredientID)
			if err != nil {
				return fmt.Errorf("item %d: ingredient '%s' not found in inventory", i+1, ingredient.IngredientID)
			}

			// Calculate consumption amount
			consumeAmount := ingredient.Quantity * float64(item.Quantity)

			// Update inventory item
			inventoryItem.Quantity -= consumeAmount

			if err := s.inventoryRepo.Update(ingredient.IngredientID, inventoryItem); err != nil {
				return fmt.Errorf("item %d: failed to update inventory for ingredient '%s': %v",
					i+1, ingredient.IngredientID, err)
			}

			s.logger.Info("Consumed inventory",
				"ingredient_id", ingredient.IngredientID,
				"amount", consumeAmount,
				"remaining", inventoryItem.Quantity)
		}
	}
	return nil
}

// restoreInventory adds back inventory quantities when order is deleted
func (s *OrderService) restoreInventory(items []CreateOrderItemRequest) error {
	for i, item := range items {
		// Get the menu item to find its ingredients
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("item %d: product '%s' not found in menu", i+1, item.ProductID)
		}

		// Restore each ingredient
		for _, ingredient := range menuItem.Ingredients {
			inventoryItem, err := s.inventoryRepo.GetByID(ingredient.IngredientID)
			if err != nil {
				return fmt.Errorf("item %d: ingredient '%s' not found in inventory", i+1, ingredient.IngredientID)
			}

			// Calculate restoration amount
			restoreAmount := ingredient.Quantity * float64(item.Quantity)

			// Update inventory item
			inventoryItem.Quantity += restoreAmount

			if err := s.inventoryRepo.Update(ingredient.IngredientID, inventoryItem); err != nil {
				return fmt.Errorf("item %d: failed to update inventory for ingredient '%s': %v",
					i+1, ingredient.IngredientID, err)
			}

			s.logger.Info("Restored inventory",
				"ingredient_id", ingredient.IngredientID,
				"amount", restoreAmount,
				"new_total", inventoryItem.Quantity)
		}
	}
	return nil
}

// parseDate parses date string in multiple formats
func (s *OrderService) parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02", // YYYY-MM-DD
		"02.01.2006", // DD.MM.YYYY
		"01/02/2006", // MM/DD/YYYY
		"2006/01/02", // YYYY/MM/DD
		"2-1-2006",   // D-M-YYYY
		"02-01-2006", // DD-MM-YYYY
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, dateStr)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (s *OrderService) validateBatchOrderRequest(orderReq models.BatchOrderItem, index int) error {
	if orderReq.CustomerName == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(orderReq.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	for i, item := range orderReq.Items {
		if item.MenuItemID == "" {
			return fmt.Errorf("item %d: menu item ID is required", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i+1)
		}

		// Check if menu item exists
		if _, err := s.menuRepo.GetByID(item.MenuItemID); err != nil {
			return fmt.Errorf("item %d: menu item '%s' not found", i+1, item.MenuItemID)
		}
	}

	return nil
}

func (s *OrderService) calculateBatchOrderTotal(items []models.BatchOrderItemDetail) (float64, error) {
	var total float64

	for i, item := range items {
		menuItem, err := s.menuRepo.GetByID(item.MenuItemID)
		if err != nil {
			return 0, fmt.Errorf("item %d: menu item '%s' not found", i+1, item.MenuItemID)
		}
		total += menuItem.Price * float64(item.Quantity)
	}

	return total, nil
}
