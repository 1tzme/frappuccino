package repositories

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: This entire file-based repository implementation should be replaced
// with PostgreSQL-backed repository. Key changes needed:
// 1. Replace map[string]*models.MenuItem with database connection
// 2. Replace JSON file operations with SQL queries for menu_items table
// 3. Remove file I/O operations (loadFromFile, saveToFile, backupFile)
// 4. Replace sync.RWMutex with database transaction handling
// 5. Convert dataFilePath to database connection dependency
// 6. Implement proper SQL schema for menu_items table with ingredients relationship

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"frappuccino/models"
	"frappuccino/pkg/database"
	"frappuccino/pkg/logger"
)

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Interface remains the same but implementation changes from JSON to SQL
type MenuRepositoryInterface interface {
	GetAll() ([]*models.MenuItem, error)
	Create(item *models.MenuItem) error
	Update(id string, item *models.MenuItem) error
	Delete(id string) error
	GetByID(id string) (*models.MenuItem, error)
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Struct now includes database connection
// New struct contains *database.DB instead of file operations
type MenuRepository struct {
	logger *logger.Logger
	db     *database.DB
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// New signature: NewMenuRepository(logger *logger.Logger, db *database.DB) *MenuRepository
func NewMenuRepository(logger *logger.Logger, db *database.DB) *MenuRepository {
	return &MenuRepository{
		logger: logger.WithComponent("menu_repository"),
		db:     db, // NEW: Store database connection
	}
}

// GetAll - retrieves all menu items
func (r *MenuRepository) GetAll() ([]*models.MenuItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load menu items from file", "error", err)
			return nil, err
		}
	}

	items := make([]*models.MenuItem, 0, len(r.items))
	for _, item := range r.items {
		itemCopy := *item
		items = append(items, &itemCopy)
	}

	r.logger.Info("Retrieved all menu items", "count", len(items))
	return items, nil
}

// Create - creates a new menu item
func (r *MenuRepository) Create(item *models.MenuItem) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load menu items from file", "error", err)
			return err
		}
	}

	_, exists := r.items[item.ID]
	if exists {
		r.logger.Warn("Attempted to create duplicate menu item", "item_id", item.ID)
		return fmt.Errorf("menu item with ID %s already exists", item.ID)
	}

	if err := r.validateMenuItem(item); err != nil {
		r.logger.Error("Failed to validate menu item", "error", err, "item_id", item.ID)
		return err
	}

	r.items[item.ID] = item

	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save menu items after create", "error", err)
		return err
	}

	r.logger.Info("Created new menu item", "item_id", item.ID, "name", item.Name, "price", item.Price)
	return nil
}

// Update - updates existing menu item
func (r *MenuRepository) Update(id string, item *models.MenuItem) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load menu items from file", "error", err)
			return err
		}
	}

	_, exists := r.items[id]
	if !exists {
		r.logger.Warn("Attempted to update non existing menu item", "item_id", id)
		return fmt.Errorf("menu item with id %s not found", id)
	}

	if err := r.validateMenuItem(item); err != nil {
		r.logger.Error("Failed to validate menu item", "error", err, "item_id", id)
		return err
	}
	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup file", "error", err)
	}

	item.ID = id
	r.items[id] = item

	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save menu items after update", "error", err, "item_id", id)
		return err
	}

	r.logger.Info("Updated menu item", "item_id", id, "name", item.Name, "price", item.Price)
	return nil
}

// Delete - removes menu item by ID
func (r *MenuRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load menu items from file", "error", err)
			return err
		}
	}

	item, exists := r.items[id]
	if !exists {
		r.logger.Warn("Attempted to delete non-existent menu item", "item_id", id)
		return fmt.Errorf("menu item with id %s not found", id)
	}
	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup before delete", "error", err)
	}

	delete(r.items, id)

	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save menu items after delete", "error", err)
		return err
	}

	r.logger.Info("Deleted menu item", "item_id", id, "name", item.Name)
	return nil
}

// GetByID - retrieves menu item by ID
func (r *MenuRepository) GetByID(id string) (*models.MenuItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load menu items from file", "error", err)
			return nil, err
		}
	}

	item, exists := r.items[id]
	if !exists {
		r.logger.Warn("Menu item not found", "item_id", id)
		return nil, fmt.Errorf("menu item with id %s not found", id)
	}

	itemCopy := *item
	r.logger.Info("Retrieved menu item", "item_id", id, "name", item.Name)
	return &itemCopy, nil
}

// TODO: Implement GetPopularItems method - Get popular menu items aggregation
// - Analyze order history
// - Count item frequencies
// - Return sorted popular items
// func (r *MenuRepository) GetPopularItems() ([]*models.PopularItemAggregation, error)

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: All file operations below should be removed and replaced with SQL queries
// - loadFromFile() → SELECT queries for menu_items table with ingredient joins
// - saveToFile() → INSERT/UPDATE queries with transactions for menu_items
// - backupFile() → Database backup strategies
// - validateMenuItem() → Database constraints and validation

func (r *MenuRepository) insertIngredients(tx *sql.Tx, menuItemId string, ingredients []models.MenuItemIngredient) error {
	if len(ingredients) == 0 {
		return nil
	}

	query := `
		INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity)
		VALUES ($1, $2, $3)
	`

	for _, ingredient := range ingredients {
		_, err := tx.Exec(query, menuItemId, ingredient.IngredientID, ingredient.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert ingredient %s: %v", ingredient.IngredientID, err)
		}
	}

	return nil
}

func (r *MenuRepository) deleteIngredients(tx *sql.Tx, menuItemId string) error {
	query := `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`
	_, err := tx.Exec(query, menuItemId)
	if err != nil {
		fmt.Errorf("failed to delete ingredient: %v", err)
	}
	return nil
}

func (r *MenuRepository) parseIngredients(ingredientsJSON string, ingredients *[]models.MenuItemIngredient) error {
	if ingredientsJSON == "" || ingredientsJSON == "[]" {
		*ingredients = []models.MenuItemIngredient{}
		return nil
	}

	raw := []models.MenuItemIngredient{}
	err := json.Unmarshal([]byte(ingredientsJSON), &raw)
	if err != nil {
		return fmt.Errorf("invalid JSON format for ingredients: %v", err)
	}

	parsed := make([]models.MenuItemIngredient, 0, len(raw))
	for _, ingredient := range raw {
		parsed = append(parsed, models.MenuItemIngredient{
			IngredientID: ingredient.IngredientID,
			Quantity:     ingredient.Quantity,
		})
	}

	*ingredients = parsed
	return nil
}

func (r *MenuRepository) validateMenuItem(item *models.MenuItem) error {
	if item == nil {
		return errors.New("menu item cannot be nil")
	}
	if item.ID == "" {
		return errors.New("item ID cannot be empty")
	}
	if item.Name == "" {
		return errors.New("item name cannot be empty")
	}
	if item.Price < 0 {
		return errors.New("price cannot be negative")
	}

	if len(item.Ingredients) == 0 {
		return errors.New("menu item must have at least 1 ingredient")
	}
	for i, ingredient := range item.Ingredients {
		if ingredient.IngredientID == "" {
			return fmt.Errorf("ingredient %d: ID cannot be empty", i+1)
		}
		if ingredient.Quantity < 0 {
			return fmt.Errorf("ingredient %d: quantity must be positive", i+1)
		}
	}

	return nil
}
