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
	"strings"

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
	r.logger.Debug("Retrieving all menu items from database")

	query := `
        SELECT m.id, m.name, m.description, m.category, m.price, m.available,
               COALESCE(
                   json_agg(
                       json_build_object(
                           'ingredient_id', mi.ingredient_id,
                           'quantity', mi.required_quantity
                       )
                   ) FILTER (WHERE mi.ingredient_id IS NOT NULL), '[]'::json
               ) as ingredients
        FROM menu_items m
        LEFT JOIN menu_item_ingredients mi ON m.id = mi.menu_item_id
        GROUP BY m.id, m.name, m.description, m.category, m.price, m.available
        ORDER BY m.name
    `

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Failed to query menu items", "error", err)
		return nil, fmt.Errorf("failed to query menu items: %v", err)
	}
	defer rows.Close()

	items := []*models.MenuItem{}
	for rows.Next() {
		item := &models.MenuItem{}
		ingredientsJSON := ""

		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.Price, &item.Available, &ingredientsJSON)
		if err != nil {
			r.logger.Error("Failed to scan menu items", "error", err)
			return nil, fmt.Errorf("failed to scan menu item: %v", err)
		}

		err = r.parseIngredients(ingredientsJSON, &item.Ingredients)
		if err != nil {
			r.logger.Error("Failed to parse ingredients", "error", err, "item_id", item.ID)
			return nil, fmt.Errorf("failed to parse ingredients for item %s: %v", item.ID, err)
		}

		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		r.logger.Error("Error iterating menu rows", "error", err)
		return nil, fmt.Errorf("error iterating menu rows: %v", err)
	}

	r.logger.Info("Retrieved all menu items", "count", len(items))
	return items, nil
}

// Create - creates a new menu item
func (r *MenuRepository) Create(item *models.MenuItem) error {
	r.logger.Debug("Adding new menu item", "item_name", item.Name)

	err := r.validateMenuItem(item)
	if err != nil {
		r.logger.Error("Failed to validate menu item", "error", err, "item_name", item.Name)
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
        INSERT INTO menu_items (id, name, description, category, price, available)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err = tx.Exec(query, item.ID, item.Name, item.Description, item.Category, item.Price, item.Available)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "violates unique constraint") {
			r.logger.Warn("Attempted to add duplicate menu item", "item_id", item.ID, "error", err)
			return fmt.Errorf("menu item with ID %s already exists", item.ID)
		}
		r.logger.Error("Failed to add menu item", "error", err, "item_id", item.ID)
		return fmt.Errorf("failed to add menu item: %v", err)
	}

	err = r.insertIngredients(tx, item.ID, item.Ingredients)
	if err != nil {
		r.logger.Error("Failed to add menu item ingredients", "error", err, "item_id", item.ID)
		return fmt.Errorf("failed to add menu item ingredients: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("Added new menu item", "item_id", item.ID, "name", item.Name)
	return nil
}

// Update - updates existing menu item
func (r *MenuRepository) Update(id string, item *models.MenuItem) error {
	r.logger.Debug("Updating menu item in database", "item_id", id)

	if err := r.validateMenuItemForUpdate(item, id); err != nil {
		r.logger.Error("Failed to validate menu item", "error", err, "item_id", id)
		return fmt.Errorf("invalid menu item: %v", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
        UPDATE menu_items
        SET name = $1, description = $2, category = $3, price = $4, available = $5
        WHERE id = $6
    `

    result, err := tx.Exec(query, item.Name, item.Description, item.Category, item.Price, item.Available, id)
	if err != nil {
		r.logger.Error("Failed to update menu item", "error", err, "item_id", item.ID)
		return fmt.Errorf("failed to update menu item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err, "item_id", item)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Attempted to update non-existent menu item", "item_id", id)
		return fmt.Errorf("menu item with id %s not found", id)
	}

	err = r.deleteIngredients(tx, id)
	if err != nil {
		r.logger.Error("Failed to delete existing ingredients", "error", err, "item_id", id)
		return fmt.Errorf("failed to delete existing ingredients: %v", err)
	}

	err = r.insertIngredients(tx, id, item.Ingredients)
	if err != nil {
		r.logger.Error("Failed to update menu item ingredients", "error", err, "item_id", id)
		return fmt.Errorf("failed to update menu item ingredients: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err, "item_id", id)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("Updated menu item", "item_id", id, "name", item.Name)
	return nil
}

// Delete - removes menu item by ID
func (r *MenuRepository) Delete(id string) error {
	r.logger.Debug("Deleting menu item from database", "item_id", id)

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	err = r.deleteIngredients(tx, id)
	if err != nil {
		r.logger.Error("Failed to delete menu item ingredients", "error", err, "item_id", id)
		return fmt.Errorf("failed to delete menu item ingredients: %v", err)
	}

	query := `DELETE FROM menu_items WHERE id = $1`
	result, err := tx.Exec(query, id)

	if err != nil {
        r.logger.Error("Failed to delete menu item", "error", err, "item_id", id)
        return fmt.Errorf("failed to delete menu item: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        r.logger.Error("Failed to get rows affected", "error", err, "item_id", id)
        return fmt.Errorf("failed to get rows affected: %v", err)
    }
    if rowsAffected == 0 {
        r.logger.Warn("Attempted to delete non-existent menu item", "item_id", id)
        return fmt.Errorf("menu item with id %s not found", id)
    }

    if err := tx.Commit(); err != nil {
        r.logger.Error("Failed to commit transaction", "error", err, "item_id", id)
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    r.logger.Info("Deleted menu item", "item_id", id)
    return nil
}

// GetByID - retrieves menu item by ID
func (r *MenuRepository) GetByID(id string) (*models.MenuItem, error) {
	r.logger.Debug("Retrieving menu item from database", "item_id", id)

    query := `
        SELECT m.id, m.name, m.description, m.category, m.price, m.available,
               COALESCE(
                   json_agg(
                       json_build_object(
                           'ingredient_id', mi.ingredient_id,
                           'quantity', mi.required_quantity
                       )
                   ) FILTER (WHERE mi.ingredient_id IS NOT NULL), '[]'::json
               ) as ingredients
        FROM menu_items m
        LEFT JOIN menu_item_ingredients mi ON m.id = mi.menu_item_id
        WHERE m.id = $1
        GROUP BY m.id, m.name, m.description, m.category, m.price, m.available
    `

	row := r.db.QueryRow(query, id)

    item := &models.MenuItem{}
    var ingredientsJSON string

    err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.Price, &item.Available, &ingredientsJSON)

    if err != nil {
        if err == sql.ErrNoRows {
            r.logger.Warn("Menu item not found", "item_id", id)
            return nil, fmt.Errorf("menu item with id %s not found", id)
        }
        r.logger.Error("Failed to retrieve menu item", "error", err, "item_id", id)
        return nil, fmt.Errorf("failed to retrieve menu item: %v", err)
    }

    if err := r.parseIngredients(ingredientsJSON, &item.Ingredients); err != nil {
        r.logger.Error("Failed to parse ingredients", "error", err, "item_id", item.ID)
        return nil, fmt.Errorf("failed to parse ingredients for item %s: %v", item.ID, err)
    }

    r.logger.Debug("Retrieved menu item", "item_id", id, "name", item.Name)
    return item, nil
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
		return fmt.Errorf("failed to delete ingredient: %v", err)
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

func (r *MenuRepository) validateMenuItemForUpdate(item *models.MenuItem, id string) error {
	if id == "" {
		return errors.New("menu item ID cannot be empty for updates")
	}
	return r.validateMenuItem(item)
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
