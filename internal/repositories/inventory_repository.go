package repositories

// TODO: Add imports when implementing:
// import (
//     "sync"
//     "hot-coffee/models"
//     "hot-coffee/pkg/logger"
// )

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

// TODO: Implement InventoryRepository interface
type InventoryRepositoryInterface interface {
	GetAll() ([]*models.InventoryItem, error)
	Update(id string, item *models.InventoryItem) error
	Add(item *models.InventoryItem) error
	GetByID(id string) (*models.InventoryItem, error)
	Delete(id string) error
}

// Add adds a new inventory item
func (r *InventoryRepository) Add(item *models.InventoryItem) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load inventory from file", "error", err)
			return err
		}
	}

	if _, exists := r.items[item.IngredientID]; exists {
		r.logger.Warn("Attempted to add duplicate inventory item", "item_id", item.IngredientID)
		return fmt.Errorf("inventory item with id %s already exists", item.IngredientID)
	}

	if err := r.validateInventoryItem(item); err != nil {
		r.logger.Error("Failed to validate inventory item", "error", err, "item_id", item.IngredientID)
		return err
	}

	r.items[item.IngredientID] = item
	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save inventory after add", "error", err)
		return err
	}
	r.logger.Info("Added new inventory item", "item_id", item.IngredientID, "name", item.Name)
	return nil
}

// GetByID retrieves a single inventory item by ID
func (r *InventoryRepository) GetByID(id string) (*models.InventoryItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load inventory from file", "error", err)
			return nil, err
		}
	}

	item, exists := r.items[id]
	if !exists {
		r.logger.Warn("Inventory item not found", "item_id", id)
		return nil, fmt.Errorf("inventory item with id %s not found", id)
	}
	itemCopy := *item
	return &itemCopy, nil
}

// Delete removes an inventory item by ID
func (r *InventoryRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load inventory from file", "error", err)
			return err
		}
	}

	if _, exists := r.items[id]; !exists {
		r.logger.Warn("Attempted to delete non-existent inventory item", "item_id", id)
		return fmt.Errorf("inventory item with id %s not found", id)
	}

	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup before delete", "error", err)
	}

	delete(r.items, id)
	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save inventory after delete", "error", err)
		return err
	}
	r.logger.Info("Deleted inventory item", "item_id", id)
	return nil
}

type InventoryRepository struct {
	items        map[string]*models.InventoryItem
	mutex        sync.RWMutex
	logger       *logger.Logger
	dataFilePath string
	loaded       bool
}

func NewInventoryRepository(logger *logger.Logger, dataDir string) *InventoryRepository {
	return &InventoryRepository{
		items:        make(map[string]*models.InventoryItem),
		logger:       logger.WithComponent("inventory_repository"),
		dataFilePath: filepath.Join(dataDir, "inventory.json"),
		loaded:       false,
	}
}

func (r *InventoryRepository) GetAll() ([]*models.InventoryItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		err := r.loadFromFile()
		if err != nil {
			r.logger.Error("Failed to load inventory from file", "error", err)
			return nil, err
		}
	}

	items := make([]*models.InventoryItem, 0, len(r.items))
	for _, item := range r.items {
		itemCopy := *item
		items = append(items, &itemCopy)
	}

	r.logger.Info("Retrieved all inventory items", "count", len(items))
	return items, nil
}

func (r *InventoryRepository) Update(id string, item *models.InventoryItem) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		err := r.loadFromFile()
		if err != nil {
			r.logger.Error("Failed to load inventory from file", "error", err)
			return fmt.Errorf("failed to load inventory: %v", err)
		}
	}

	_, exists := r.items[id]
	if !exists {
		r.logger.Warn("Attempted to update non existing inventory item", "item_id", id)
		return fmt.Errorf("inventory item with id %s not found", id)
	}

	err := r.validateInventoryItem(item)
	if err != nil {
		r.logger.Error("Failed to validate inventory item", "error", err, "item_id", id)
		return fmt.Errorf("invalid inventory item: %v", err)
	}

	err = r.backupFile()
	if err != nil {
		r.logger.Warn("Failed to create backup", "error", err)
	}

	item.IngredientID = id
	r.items[id] = item

	err = r.saveToFile()
	if err != nil {
		r.logger.Error("Failed to save inventory to file after update", "error", err, "item_id", id)
		return fmt.Errorf("failed to save inventory: %v", err)
	}

	r.logger.Info("Updated inventory item", "item_id", id, "name", item.Name)
	return nil
}

func (r *InventoryRepository) loadFromFile() error {
	err := os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755)
	if err != nil {
		return err
	}

	_, err = os.Stat(r.dataFilePath)
	if err != nil {
		r.items = make(map[string]*models.InventoryItem)
		r.loaded = true
		return r.saveToFile()
	}

	file, err := os.Open(r.dataFilePath)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		r.items = make(map[string]*models.InventoryItem)
		r.loaded = true
		return err
	}

	items := []*models.InventoryItem{}
	err = json.Unmarshal(data, &items)
	if err != nil {
		return err
	}

	r.items = make(map[string]*models.InventoryItem)
	for _, item := range items {
		r.items[item.IngredientID] = item
	}

	r.loaded = true
	return nil
}

func (r *InventoryRepository) saveToFile() error {
	items := make([]*models.InventoryItem, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal inventory data: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	tempFile := r.dataFilePath + ".tmp"
	err = os.WriteFile(tempFile, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write temporary inventory file: %v", err)
	}

	err = os.Rename(tempFile, r.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename file path: %v", err)
	}

	return nil
}

func (r *InventoryRepository) validateInventoryItem(item *models.InventoryItem) error {
	if item == nil {
		return errors.New("inventory item cannot be nil")
	}
	if item.IngredientID == "" {
		return errors.New("ingridient ID cannot be empty")
	}
	if item.Name == "" {
		return errors.New("ingridient name cannot be empty")
	}
	if item.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	if item.Unit == "" {
		return errors.New("unit cannot be empty")
	}

	return nil
}

func (r *InventoryRepository) backupFile() error {
	_, err := os.Stat(r.dataFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	backupPath := r.dataFilePath + ".backup." + time.Now().Format("20060102_150405")

	data, err := os.ReadFile(r.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %v", err)
	}

	err = os.WriteFile(backupPath, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create backup file, %v", err)
	}

	return nil
}
