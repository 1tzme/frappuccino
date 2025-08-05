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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	items        map[string]*models.MenuItem // TODO: Replace with database queries
	mutex        sync.RWMutex                // TODO: Remove (database handles this)
	logger       *logger.Logger
	db           *database.DB // NEW: Database connection
	dataFilePath string       // TODO: Remove (no more file operations)
	loaded       bool         // TODO: Remove (no more file loading)
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// New signature: NewMenuRepository(logger *logger.Logger, db *database.DB) *MenuRepository
func NewMenuRepository(logger *logger.Logger, db *database.DB) *MenuRepository {
	return &MenuRepository{
		items:        make(map[string]*models.MenuItem), // TODO: Replace with database queries
		logger:       logger.WithComponent("menu_repository"),
		db:           db,   // NEW: Store database connection
		dataFilePath: "",   // TODO: Remove completely
		loaded:       true, // Skip file loading during transition
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

func (r *MenuRepository) loadFromFile() error {
	if err := os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	if _, err := os.Stat(r.dataFilePath); err != nil {
		r.items = make(map[string]*models.MenuItem)
		r.loaded = true
		return r.saveToFile()
	}

	file, err := os.Open(r.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to open menu items file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to open menu items file: %v", err)
	}

	if len(data) == 0 {
		r.items = make(map[string]*models.MenuItem)
		r.loaded = true
		return nil
	}

	items := []*models.MenuItem{}
	if err = json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to unmarshal menu items: %v", err)
	}

	r.items = make(map[string]*models.MenuItem)
	for _, item := range items {
		r.items[item.ID] = item
	}

	r.loaded = true
	r.logger.Debug("Loaded menu items from file", "count", len(r.items))
	return nil
}

func (r *MenuRepository) saveToFile() error {
	items := make([]*models.MenuItem, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal menu items: %v", err)
	}
	if err = os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	tempFile := r.dataFilePath + ".tmp"
	if err = os.WriteFile(tempFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temporary menu items file: %v", err)
	}

	if err = os.Rename(tempFile, r.dataFilePath); err != nil {
		return fmt.Errorf("failed to rename menu items file: %v", err)
	}

	r.logger.Debug("Save menu items to file", "count", len(items))
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

func (r *MenuRepository) backupFile() error {
	if _, err := os.Stat(r.dataFilePath); os.IsNotExist(err) {
		return nil
	}

	backupPath := r.dataFilePath + ".backup." + time.Now().Format("20060102_150405")

	data, err := os.ReadFile(r.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %v", err)
	}
	if err = os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to create backup file, %v", err)
	}

	r.logger.Debug("Created backup file", "backup_path", backupPath)
	return nil
}
