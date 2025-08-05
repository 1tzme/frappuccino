package service

import (
	"fmt"

	"hot-coffee/internal/repositories"
	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

type UpdateInventoryItemRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Quantity     int    `json:"quantity"`
	MinThreshold int    `json:"min_threshold"`
	Unit         string `json:"unit"`
}

type InventoryServiceInterface interface {
	GetAllInventoryItems() ([]*models.InventoryItem, error)
	UpdateInventoryItem(id string, req UpdateInventoryItemRequest) error
	CreateInventoryItem(id string, req UpdateInventoryItemRequest) error
	GetInventoryItem(id string) (*models.InventoryItem, error)
	DeleteInventoryItem(id string) error
}

// CreateInventoryItem adds a new inventory item
func (s *InventoryService) CreateInventoryItem(id string, req UpdateInventoryItemRequest) error {
	s.logger.Info("Creating inventory item", "id", id, "name", req.Name)
	if err := validateCreateInventoryItemData(req); err != nil {
		s.logger.Warn("Create failed: invalid data", "id", id, "error", err)
		return err
	}
	item := &models.InventoryItem{
		IngredientID: id,
		Name:         req.Name,
		Quantity:     float64(req.Quantity),
		MinThreshold: float64(req.MinThreshold),
		Unit:         req.Unit,
	}
	if err := s.inventoryRepo.Add(item); err != nil {
		s.logger.Error("Failed to add inventory item in repository", "id", id, "error", err)
		return err
	}
	s.logger.Info("Inventory item created", "id", id)
	return nil
}

// GetInventoryItem fetches a single inventory item by ID
func (s *InventoryService) GetInventoryItem(id string) (*models.InventoryItem, error) {
	s.logger.Info("Fetching inventory item by id", "id", id)
	item, err := s.inventoryRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Inventory item not found", "id", id, "error", err)
		return nil, err
	}
	s.logger.Info("Fetched inventory item", "id", id)
	return item, nil
}

// DeleteInventoryItem deletes an inventory item by ID
func (s *InventoryService) DeleteInventoryItem(id string) error {
	s.logger.Info("Deleting inventory item", "id", id)

	// Check if ingredient is used in any existing orders
	if err := s.checkIngredientUsageInOrders(id); err != nil {
		s.logger.Warn("Cannot delete ingredient: used in orders", "id", id, "error", err)
		return err
	}

	// Check if ingredient is used in any menu items
	if err := s.checkIngredientUsageInMenu(id); err != nil {
		s.logger.Warn("Cannot delete ingredient: used in menu", "id", id, "error", err)
		return err
	}

	if err := s.inventoryRepo.Delete(id); err != nil {
		s.logger.Warn("Failed to delete inventory item", "id", id, "error", err)
		return err
	}
	s.logger.Info("Inventory item deleted", "id", id)
	return nil
}

type InventoryService struct {
	inventoryRepo repositories.InventoryRepositoryInterface
	orderRepo     repositories.OrderRepositoryInterface
	menuRepo      repositories.MenuRepositoryInterface
	logger        *logger.Logger
}

// NewInventoryService creates a new instance of InventoryService
func NewInventoryService(inventoryRepo repositories.InventoryRepositoryInterface, orderRepo repositories.OrderRepositoryInterface, menuRepo repositories.MenuRepositoryInterface, logger *logger.Logger) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		orderRepo:     orderRepo,
		menuRepo:      menuRepo,
		logger:        logger.WithComponent("inventory_service"),
	}
}

// GetAllInventoryItems returns all inventory items (placeholder implementation)
func (s *InventoryService) GetAllInventoryItems() ([]*models.InventoryItem, error) {
	s.logger.Info("Fetching all inventory items from repository")
	items, err := s.inventoryRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to fetch inventory items from repository", "error", err)
		return nil, err
	}
	s.logger.Info("Fetched inventory items", "count", len(items))
	return items, nil
}

// UpdateInventoryItem updates an existing inventory item
func (s *InventoryService) UpdateInventoryItem(id string, req UpdateInventoryItemRequest) error {
	s.logger.Info("Updating inventory item", "id", id, "name", req.Name)

	existingItem, err := s.inventoryRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Failed to get inventory item for update", "id", id, "error", err)
		return err
	}

	if existingItem.Name == req.Name && existingItem.Quantity == float64(req.Quantity) && existingItem.MinThreshold == float64(req.MinThreshold) && existingItem.Unit == req.Unit {
		s.logger.Warn("Update canceled: no changes detected", "id", id)
		return fmt.Errorf("no changes detected for inventory item with ID %s", id)
	}
	// Validate input
	if err := validateUpdateInventoryItemData(req); err != nil {
		s.logger.Warn("Update failed: invalid data", "id", id, "error", err)
		return err
	}

	// Build item struct for update
	item := &models.InventoryItem{
		IngredientID: id,
		Name:         req.Name,
		Quantity:     float64(req.Quantity),
		MinThreshold: float64(req.MinThreshold),
		Unit:         req.Unit,
	}

	err = s.inventoryRepo.Update(id, item)
	if err != nil {
		s.logger.Error("Failed to update inventory item in repository", "id", id, "error", err)
		return err
	}

	s.logger.Info("Inventory item updated", "id", id)
	return nil
}

// Private business logic methods

// validateCreateInventoryItemData validates data for creation
func validateCreateInventoryItemData(req UpdateInventoryItemRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be non-negative")
	}
	if req.MinThreshold < 0 {
		return fmt.Errorf("minimum threshold must be non-negative")
	}
	if req.Unit == "" {
		return fmt.Errorf("unit is required")
	}
	return nil
}

// validateUpdateInventoryItemData validates data for update
func validateUpdateInventoryItemData(req UpdateInventoryItemRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Quantity < 0 {
		return fmt.Errorf("quantity must be non-negative")
	}
	if req.MinThreshold < 0 {
		return fmt.Errorf("minimum threshold must be non-negative")
	}
	if req.Unit == "" {
		return fmt.Errorf("unit is required")
	}
	return nil
}

// checkIngredientUsageInOrders checks if an ingredient is used in any existing orders
func (s *InventoryService) checkIngredientUsageInOrders(ingredientID string) error {
	orders, err := s.orderRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to check orders: %v", err)
	}

	for _, order := range orders {
		// Only check open orders (not closed orders)
		if order.Status != "closed" {
			for _, orderItem := range order.Items {
				// Get the menu item to check its ingredients
				menuItem, err := s.menuRepo.GetByID(orderItem.ProductID)
				if err != nil {
					// If menu item doesn't exist, skip this order item
					continue
				}

				// Check if this ingredient is used in the menu item
				for _, ingredient := range menuItem.Ingredients {
					if ingredient.IngredientID == ingredientID {
						return fmt.Errorf("ingredient '%s' is used in open order '%s' for product '%s'",
							ingredientID, order.ID, orderItem.ProductID)
					}
				}
			}
		}
	}
	return nil
}

// checkIngredientUsageInMenu checks if an ingredient is used in any menu items
func (s *InventoryService) checkIngredientUsageInMenu(ingredientID string) error {
	menuItems, err := s.menuRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to check menu items: %v", err)
	}

	for _, menuItem := range menuItems {
		for _, ingredient := range menuItem.Ingredients {
			if ingredient.IngredientID == ingredientID {
				return fmt.Errorf("ingredient '%s' is used in menu item '%s' (%s)",
					ingredientID, menuItem.ID, menuItem.Name)
			}
		}
	}
	return nil
}
