package service

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update menu business logic for PostgreSQL database operations
// 1. Replace file-based menu validation with database constraints
// 2. Implement proper transaction handling for menu item operations
// 3. Add database relationship management for menu items and ingredients
// 4. Update availability logic to use database triggers and views
// 5. Implement database-based menu categorization and search features

import (
	"fmt"
	"strings"

	"frappuccino/internal/repositories"
	"frappuccino/models"
	"frappuccino/pkg/logger"
)

type CreateMenuItemRequest struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Category    models.MenuCategory         `json:"category"`
	Price       float64                     `json:"price"`
	Available   bool                        `json:"available"`
	Ingredients []models.MenuItemIngredient `json:"ingredients"`
}

type UpdateMenuItemRequest struct {
	Name        *string                      `json:"name"`
	Description *string                      `json:"description"`
	Category    *models.MenuCategory         `json:"category"`
	Price       *float64                     `json:"price"`
	Available   *bool                        `json:"available"`
	Ingredients *[]models.MenuItemIngredient `json:"ingredients"`
}

type MenuServiceInterface interface {
	GetAllMenuItems() ([]*models.MenuItem, error)
	GetMenuItem(id string) (*models.MenuItem, error)
	CreateMenuItem(req CreateMenuItemRequest) (*models.MenuItem, error)
	UpdateMenuItem(id string, req UpdateMenuItemRequest) error
	DeleteMenuItem(id string) error
}

type MenuService struct {
	menuRepo      repositories.MenuRepositoryInterface
	inventoryRepo repositories.InventoryRepositoryInterface
	orderRepo     repositories.OrderRepositoryInterface
	logger        *logger.Logger
}

func NewMenuService(inventoryRepo repositories.InventoryRepositoryInterface, menuRepo repositories.MenuRepositoryInterface, orderRepo repositories.OrderRepositoryInterface, logger *logger.Logger) *MenuService {
	return &MenuService{
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
		orderRepo:     orderRepo,
		logger:        logger.WithComponent("menu_service"),
	}
}

// GetAllMenuItems retrieves all menu items
func (s *MenuService) GetAllMenuItems() ([]*models.MenuItem, error) {
	s.logger.Info("Fetching all menu items from repository")

	items, err := s.menuRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to get menu items from repository", "error", err)
		return nil, err
	}

	s.logger.Info("Fetched menu items", "count", len(items))
	return items, nil
}

// CreateMenuItem creates new menu item
func (s *MenuService) CreateMenuItem(req CreateMenuItemRequest) (*models.MenuItem, error) {
	s.logger.Info("Creating menu item", "name", req.Name, "price", req.Price)

	if err := s.validateCreateMenuItemData(req); err != nil {
		s.logger.Warn("Create failed: invalid data", "error", err)
		return nil, err
	}
	if err := s.validateIngredients(req.Ingredients); err != nil {
		s.logger.Warn("")
		return nil, err
	}

	newID := s.generateMenuItemID(req.Name)

	item := &models.MenuItem{
		ID:          newID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		Available:   req.Available,
		Ingredients: req.Ingredients,
	}

	if err := s.menuRepo.Create(item); err != nil {
		s.logger.Error("Failed to create menu item in repository", "id", newID, "error", err)
		return nil, err
	}

	s.logger.Info("Menu item created successfully", "id", newID, "name", req.Name)
	return item, nil
}

// UpdateMenuItem updates existing menu item
func (s *MenuService) UpdateMenuItem(id string, req UpdateMenuItemRequest) error {
	s.logger.Info("Updating menu item", "id", id, "name", req.Name, "price", req.Price)

	if err := s.validateUpdateMenuItemData(req); err != nil {
		s.logger.Warn("Update failed: invalid data", "id", id, "error", err)
		return err
	}

	if err := s.checkMenuItemUsageInOrders(id); err != nil {
		s.logger.Warn("Cannot update menu item: used in orders", "id", id, "error", err)
		return err
	}

	existingItem, err := s.menuRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get existing menu item", "id", id, "error", err)
		return err
	}
	if req.Ingredients != nil {
		if err := s.validateIngredients(*req.Ingredients); err != nil {
			return err
		}
	}
	if err := s.validateUpdateMenuItemData(req); err != nil {
		s.logger.Warn("Update failed: invalid data", "id", id, "error", err)
		return err
	}

	updatedItem := &models.MenuItem{
		ID:          id,
		Name:        existingItem.Name,
		Description: existingItem.Description,
		Category:    existingItem.Category,
		Price:       existingItem.Price,
		Available:   existingItem.Available,
		Ingredients: existingItem.Ingredients,
	}

	if req.Name != nil {
		updatedItem.Name = *req.Name
	}
	if req.Description != nil {
		updatedItem.Description = *req.Description
	}
	if req.Category != nil {
		updatedItem.Category = *req.Category
	}
	if req.Price != nil {
		updatedItem.Price = *req.Price
	}
	if req.Available != nil {
		updatedItem.Available = *req.Available
	}
	if req.Ingredients != nil {
		updatedItem.Ingredients = *req.Ingredients
	}

	if s.hasMenuItemChanged(existingItem, updatedItem) {
		if err := s.menuRepo.Update(id, updatedItem); err != nil {
			s.logger.Error("Failed to update menu item", "id", id, "error", err)
			return err
		}
		s.logger.Info("Menu item updated successfully", "id", id)
	} else {
		s.logger.Warn("Update canceled: no changes detected", "id", id)
		return fmt.Errorf("no changes detected for menu item with ID %s", id)
	}

	s.logger.Info("Menu item updated successfully", "id", id, "name", req.Name)
	return nil
}

// DeleteMenuItem deletes menu item
func (s *MenuService) DeleteMenuItem(id string) error {
	s.logger.Info("Deleting menu item", "id", id)

	if _, err := s.menuRepo.GetByID(id); err != nil {
		s.logger.Warn("Menu item not found for deletion", "id", id, "error", err)
		return err
	}

	if err := s.checkMenuItemUsageInOrders(id); err != nil {
		s.logger.Warn("Cannot delete menu item: used in orders", "id", id, "error", err)
		return err
	}

	if err := s.menuRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete menu item from repository", "id", id, "error", err)
		return err
	}

	s.logger.Info("Menu item deleted successfully", "id", id)
	return nil
}

// GetMenuItem retrieves menu item by ID
func (s *MenuService) GetMenuItem(id string) (*models.MenuItem, error) {
	item, err := s.menuRepo.GetByID(id)
	if err != nil {
		s.logger.Warn("Menu item not found", "id", id, "error", err)
		return nil, err
	}

	s.logger.Info("Fetched menu item successfully", "id", id, "name", item.Name)
	return item, nil
}

// checkMenuItemUsageInOrders checks if a menu item is used in any existing orders
func (s *MenuService) validateCreateMenuItemData(req CreateMenuItemRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	if len(req.Ingredients) == 0 {
		return fmt.Errorf("menu item must have at least 1 ingredient")
	}

	for i, ingredient := range req.Ingredients {
		if ingredient.IngredientID == "" {
			return fmt.Errorf("ingredient %d: ID is required", i+1)
		}
		if ingredient.Quantity <= 0 {
			return fmt.Errorf("ingredient %d: quantity must be positive", i+1)
		}
	}

	return nil
}

// validateUpdateMenuItemData validates the update request for a menu item
func (s *MenuService) validateUpdateMenuItemData(req UpdateMenuItemRequest) error {
	if req.Name != nil && *req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Price != nil && *req.Price < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	if req.Category != nil {
		if err := s.validateMenuCategory(*req.Category); err != nil {
			return err
		}
	}
	if req.Ingredients != nil {
		if len(*req.Ingredients) == 0 {
			return fmt.Errorf("menu item must have at least 1 ingredient")
		}
	}

	for i, ingredient := range *req.Ingredients {
		if ingredient.IngredientID == "" {
			return fmt.Errorf("ingredient %d: ID is required", i+1)
		}
		if ingredient.Quantity <= 0 {
			return fmt.Errorf("ingredient %d: quantity must be positive", i+1)
		}
	}

	return nil
}

// validateMenuCategory checks if the category is valid
func (s *MenuService) validateMenuCategory(category models.MenuCategory) error {
	switch category {
	case models.CategoryCoffee, models.CategoryDrink, models.CategoryPastry, models.CategorySandwich, models.CategoryTea:
		return nil
	default:
		return fmt.Errorf("invalid menu category: %s", category)
	}
}

// checkMenuItemUsageInOrders checks if a menu item is used in any existing orders
func (s *MenuService) checkMenuItemUsageInOrders(menuItemID string) error {
	orders, err := s.orderRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to check orders: %v", err)
	}

	for _, order := range orders {
		// Only check open orders (not closed orders)
		if order.Status != "closed" {
			for _, orderItem := range order.Items {
				if orderItem.ProductID == menuItemID {
					return fmt.Errorf("menu item '%s' is used in open order '%s'",
						menuItemID, order.ID)
				}
			}
		}
	}
	return nil
}

func (s *MenuService) validateIngredients(ingredients []models.MenuItemIngredient) error {
	for _, requiredIng := range ingredients {
		_, err := s.inventoryRepo.GetByID(requiredIng.IngredientID)
		if err != nil {
			s.logger.Warn("Validation failed: ingredient not found in inventory", "ingredient_id", requiredIng.IngredientID)
			return fmt.Errorf("ingredoent with ID %s not found", requiredIng.IngredientID)
		}

		// if inventoryItem.Quantity < requiredIng.Quantity {
		// 	s.logger.Warn("Validation failed: lack of ingredient quantity", "ingredient_id", requiredIng.IngredientID, "required", requiredIng.Quantity, "available", inventoryItem.Quantity)
		// 	return fmt.Errorf("insufficient quantity for ingredient %s", inventoryItem.Name)
		// }
	}
	return nil
}

// hasMenuItemChanged checks if any changes were made to menu item
func (s *MenuService) hasMenuItemChanged(existing, updated *models.MenuItem) bool {
	if existing.Name != updated.Name || existing.Description != updated.Description ||
		existing.Category != updated.Category || existing.Price != updated.Price || existing.Available != updated.Available {
		return true
	}

	if len(existing.Ingredients) != len(updated.Ingredients) {
		return true
	}

	existingIngredients := make(map[string]float64)
	for _, ing := range existing.Ingredients {
		existingIngredients[ing.IngredientID] = ing.Quantity
	}

	for _, ing := range updated.Ingredients {
		if qty, exists := existingIngredients[ing.IngredientID]; !exists || qty != ing.Quantity {
			return true
		}
	}

	return false
}

// generateMenuItemID generates menu item ID based on the name
func (s *MenuService) generateMenuItemID(name string) string {
	cleaned := strings.ToLower(strings.TrimSpace(name))
	cleaned = strings.ReplaceAll(cleaned, " ", "_")

	result := ""
	for _, char := range cleaned {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result += string(char)
		}
	}

	if result == "" {
		result = "menu_item"
	}

	return result
}
