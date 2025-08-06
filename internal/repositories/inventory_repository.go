package repositories

// TODO: Transition State: JSON → PostgreSQL
// ✅ COMPLETED: PostgreSQL-backed repository implementation
// ✅ 1. Replaced map[string]*models.InventoryItem with database connection
// ✅ 2. Replaced JSON file operations with SQL queries for inventory table
// ✅ 3. Removed file I/O operations (loadFromFile, saveToFile, backupFile)
// ✅ 4. Replaced sync.RWMutex with database transaction handling
// ✅ 5. Converted dataFilePath to database connection dependency
// ✅ 6. COMPLETED: Uses existing PostgreSQL schema with inventory table

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"frappuccino/models"
	"frappuccino/pkg/database"
	"frappuccino/pkg/logger"
)

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Interface remains the same but implementation changes from JSON to SQL
type InventoryRepositoryInterface interface {
	GetAll() ([]*models.InventoryItem, error)
	Update(id string, item *models.InventoryItem) error
	Add(item *models.InventoryItem) error
	GetByID(id string) (*models.InventoryItem, error)
	Delete(id string) error
}

// Add adds a new inventory item
func (r *InventoryRepository) Add(item *models.InventoryItem) error {
	r.logger.Debug("Adding new inventory item to database", "item_name", item.Name)

	if err := r.validateInventoryItem(item); err != nil {
		r.logger.Error("Failed to validate inventory item", "error", err, "item_name", item.Name)
		return err
	}

	query := `
		INSERT INTO inventory (name, quantity, unit, min_threshold) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var generatedID string
	err := r.db.QueryRow(query, item.Name, item.Quantity, item.Unit, item.MinThreshold).Scan(&generatedID)
	if err != nil {
		// Check if this is a duplicate key error (PostgreSQL constraint violation)
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "violates unique constraint") {
			r.logger.Warn("Attempted to add duplicate inventory item", "item_name", item.Name, "error", err)
			return fmt.Errorf("inventory item with name %s already exists", item.Name)
		}
		r.logger.Error("Failed to add inventory item", "error", err, "item_name", item.Name)
		return fmt.Errorf("failed to add inventory item: %v", err)
	}

	// Update the item with the generated ID
	item.IngredientID = generatedID

	r.logger.Info("Added new inventory item", "item_id", item.IngredientID, "name", item.Name)
	return nil
}

// GetByID retrieves a single inventory item by ID
func (r *InventoryRepository) GetByID(id string) (*models.InventoryItem, error) {
	r.logger.Debug("Retrieving inventory item from database", "item_id", id)

	query := `
		SELECT id, name, quantity, unit, min_threshold 
		FROM inventory 
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	item := &models.InventoryItem{}
	err := row.Scan(
		&item.IngredientID,
		&item.Name,
		&item.Quantity,
		&item.Unit,
		&item.MinThreshold,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Inventory item not found", "item_id", id)
			return nil, fmt.Errorf("inventory item with id %s not found", id)
		}
		r.logger.Error("Failed to retrieve inventory item", "error", err, "item_id", id)
		return nil, fmt.Errorf("failed to retrieve inventory item: %v", err)
	}

	r.logger.Debug("Retrieved inventory item", "item_id", id, "name", item.Name)
	return item, nil
}

// Delete removes an inventory item by ID
func (r *InventoryRepository) Delete(id string) error {
	r.logger.Debug("Deleting inventory item from database", "item_id", id)

	query := `DELETE FROM inventory WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Error("Failed to delete inventory item", "error", err, "item_id", id)
		return fmt.Errorf("failed to delete inventory item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "item_id", id)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Attempted to delete non-existent inventory item", "item_id", id)
		return fmt.Errorf("inventory item with id %s not found", id)
	}

	r.logger.Info("Deleted inventory item", "item_id", id)
	return nil
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Struct now uses database connection instead of map
// Removed map, mutex, file operations - using database for storage
type InventoryRepository struct {
	logger *logger.Logger
	db     *database.DB // Database connection for SQL operations
}

// TODO: Transition State: JSON → PostgreSQL
// UPDATED: Constructor now creates database-backed repository
// New signature: NewInventoryRepository(logger *logger.Logger, db *database.DB) *InventoryRepository
func NewInventoryRepository(logger *logger.Logger, db *database.DB) *InventoryRepository {
	return &InventoryRepository{
		logger: logger.WithComponent("inventory_repository"),
		db:     db, // Store database connection for SQL operations
	}
}

func (r *InventoryRepository) GetAll() ([]*models.InventoryItem, error) {
	r.logger.Debug("Retrieving all inventory items from database")

	query := `
		SELECT id, name, quantity, unit, min_threshold 
		FROM inventory 
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Failed to query inventory items", "error", err)
		return nil, fmt.Errorf("failed to query inventory items: %v", err)
	}
	defer rows.Close()

	var items []*models.InventoryItem
	for rows.Next() {
		item := &models.InventoryItem{}
		err := rows.Scan(
			&item.IngredientID,
			&item.Name,
			&item.Quantity,
			&item.Unit,
			&item.MinThreshold,
		)
		if err != nil {
			r.logger.Error("Failed to scan inventory item", "error", err)
			return nil, fmt.Errorf("failed to scan inventory item: %v", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating inventory rows", "error", err)
		return nil, fmt.Errorf("error iterating inventory rows: %v", err)
	}

	r.logger.Info("Retrieved all inventory items", "count", len(items))
	return items, nil
}

func (r *InventoryRepository) Update(id string, item *models.InventoryItem) error {
	r.logger.Debug("Updating inventory item in database", "item_id", id)

	if err := r.validateInventoryItemForUpdate(item, id); err != nil {
		r.logger.Error("Failed to validate inventory item", "error", err, "item_id", id)
		return fmt.Errorf("invalid inventory item: %v", err)
	}

	// Ensure the item ID matches the parameter
	item.IngredientID = id

	query := `
		UPDATE inventory 
		SET name = $1, quantity = $2, unit = $3, min_threshold = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(query, item.Name, item.Quantity, item.Unit, item.MinThreshold, id)
	if err != nil {
		r.logger.Error("Failed to update inventory item", "error", err, "item_id", id)
		return fmt.Errorf("failed to update inventory item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "item_id", id)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Attempted to update non-existent inventory item", "item_id", id)
		return fmt.Errorf("inventory item with id %s not found", id)
	}

	r.logger.Info("Updated inventory item", "item_id", id, "name", item.Name)
	return nil
}

func (r *InventoryRepository) validateInventoryItemForUpdate(item *models.InventoryItem, id string) error {
	if id == "" {
		return errors.New("ingredient ID cannot be empty for updates")
	}
	return r.validateInventoryItem(item)
}

func (r *InventoryRepository) validateInventoryItem(item *models.InventoryItem) error {
	if item == nil {
		return errors.New("inventory item cannot be nil")
	}
	// Note: IngredientID can be empty for new items (will be auto-generated)
	if item.Name == "" {
		return errors.New("ingredient name cannot be empty")
	}
	if item.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	if item.Unit == "" {
		return errors.New("unit cannot be empty")
	}

	return nil
}
