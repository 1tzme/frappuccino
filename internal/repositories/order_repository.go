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
// DEPRECATED: Replace with PostgreSQL-backed repository interface
// Interface should remain the same but implementation will change from JSON files to SQL operations
type OrderRepositoryInterface interface {
	GetAll() ([]*models.Order, error)
	GetByID(id string) (*models.Order, error)
	Add(order *models.Order) error
	Update(id string, order *models.Order) error
	Delete(id string) error
	CloseOrder(id string) error
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// Signature: NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository
// Temporarily keeping file operations while transitioning to database
type OrderRepository struct {
	orders       map[string]*models.Order // TODO: Replace with database queries
	mutex        sync.RWMutex             // TODO: Remove (database handles this)
	logger       *logger.Logger
	db           *database.DB // NEW: Database connection
	dataFilePath string       // TODO: Remove (no more file operations)
	loaded       bool         // TODO: Remove (no more file loading)
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now accepts database connection instead of dataDir
// New signature: NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository
// Temporarily falls back to in-memory storage during transition
func NewOrderRepository(logger *logger.Logger, db *database.DB) *OrderRepository {
	return &OrderRepository{
		orders:       make(map[string]*models.Order), // TODO: Replace with database queries
		logger:       logger.WithComponent("order_repository"),
		db:           db,   // NEW: Store database connection
		dataFilePath: "",   // TODO: Remove completely
		loaded:       true, // Skip file loading during transition
	}
}

// Add adds a new order
func (r *OrderRepository) Add(order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return err
		}
	}

	if _, exists := r.orders[order.ID]; exists {
		r.logger.Warn("Attempted to add duplicate order", "order_id", order.ID)
		return fmt.Errorf("order with id %s already exists", order.ID)
	}

	if err := r.validateOrder(order); err != nil {
		r.logger.Error("Failed to validate order", "error", err, "order_id", order.ID)
		return err
	}

	r.orders[order.ID] = order
	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save orders after add", "error", err)
		return err
	}
	r.logger.Info("Added new order", "order_id", order.ID, "customer", order.CustomerName)
	return nil
}

// GetByID retrieves a single order by ID
func (r *OrderRepository) GetByID(id string) (*models.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		r.mutex.RUnlock()
		r.mutex.Lock()
		defer r.mutex.Unlock()
		defer r.mutex.RLock()
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return nil, err
		}
	}

	order, exists := r.orders[id]
	if !exists {
		r.logger.Warn("Order not found", "order_id", id)
		return nil, fmt.Errorf("order with id %s not found", id)
	}
	orderCopy := *order
	return &orderCopy, nil
}

// GetAll retrieves all orders
func (r *OrderRepository) GetAll() ([]*models.Order, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return nil, err
		}
	}

	orders := make([]*models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orderCopy := *order
		orders = append(orders, &orderCopy)
	}

	r.logger.Info("Retrieved all orders", "count", len(orders))
	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(id string, order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return fmt.Errorf("failed to load orders: %v", err)
		}
	}

	_, exists := r.orders[id]
	if !exists {
		r.logger.Warn("Attempted to update non-existent order", "order_id", id)
		return fmt.Errorf("order with id %s not found", id)
	}

	if err := r.validateOrder(order); err != nil {
		r.logger.Error("Failed to validate order", "error", err, "order_id", id)
		return fmt.Errorf("invalid order: %v", err)
	}

	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup", "error", err)
	}

	order.ID = id
	r.orders[id] = order

	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save orders after update", "error", err, "order_id", id)
		return fmt.Errorf("failed to save orders: %v", err)
	}

	r.logger.Info("Updated order", "order_id", id, "customer", order.CustomerName)
	return nil
}

// Delete removes an order by ID
func (r *OrderRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return err
		}
	}

	if _, exists := r.orders[id]; !exists {
		r.logger.Warn("Attempted to delete non-existent order", "order_id", id)
		return fmt.Errorf("order with id %s not found", id)
	}

	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup before delete", "error", err)
	}

	delete(r.orders, id)
	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save orders after delete", "error", err)
		return err
	}
	r.logger.Info("Deleted order", "order_id", id)
	return nil
}

// CloseOrder closes an order by setting status to closed
func (r *OrderRepository) CloseOrder(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.loaded {
		if err := r.loadFromFile(); err != nil {
			r.logger.Error("Failed to load orders from file", "error", err)
			return err
		}
	}

	order, exists := r.orders[id]
	if !exists {
		r.logger.Warn("Attempted to close non-existent order", "order_id", id)
		return fmt.Errorf("order with id %s not found", id)
	}

	if err := r.backupFile(); err != nil {
		r.logger.Warn("Failed to create backup before close", "error", err)
	}

	order.Status = "closed"
	if err := r.saveToFile(); err != nil {
		r.logger.Error("Failed to save orders after close", "error", err)
		return err
	}
	r.logger.Info("Closed order", "order_id", id)
	return nil
}

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: All file operations below should be removed and replaced with SQL queries
// - loadFromFile() → SELECT queries with proper table joins
// - saveToFile() → INSERT/UPDATE queries with transactions
// - backupFile() → Database backup strategies (pg_dump, WAL archiving)
// - validateOrder() → Database constraints and triggers

// loadFromFile loads orders from JSON file
func (r *OrderRepository) loadFromFile() error {
	err := os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755)
	if err != nil {
		return err
	}

	_, err = os.Stat(r.dataFilePath)
	if err != nil {
		r.orders = make(map[string]*models.Order)
		r.loaded = true
		return r.saveToFile()
	}

	file, err := os.Open(r.dataFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		r.orders = make(map[string]*models.Order)
		r.loaded = true
		return nil
	}

	orders := []*models.Order{}
	err = json.Unmarshal(data, &orders)
	if err != nil {
		return err
	}

	r.orders = make(map[string]*models.Order)
	for _, order := range orders {
		r.orders[order.ID] = order
	}

	r.loaded = true
	return nil
}

// saveToFile saves orders to JSON file atomically
func (r *OrderRepository) saveToFile() error {
	orders := make([]*models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}

	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal order data: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(r.dataFilePath), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	tempFile := r.dataFilePath + ".tmp"
	err = os.WriteFile(tempFile, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write temporary order file: %v", err)
	}

	err = os.Rename(tempFile, r.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename file path: %v", err)
	}

	return nil
}

// validateOrder validates order data
func (r *OrderRepository) validateOrder(order *models.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.ID == "" {
		return errors.New("order ID cannot be empty")
	}
	if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	}
	if len(order.Items) == 0 {
		return errors.New("order must have at least one item")
	}

	for i, item := range order.Items {
		if item.ProductID == "" {
			return fmt.Errorf("item %d: product ID cannot be empty", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i)
		}
	}

	return nil
}

// backupFile creates a backup of the current data file
func (r *OrderRepository) backupFile() error {
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
		return fmt.Errorf("failed to create backup file: %v", err)
	}

	return nil
}
