package models

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update struct tags to include database column mappings
// 1. Add `db:"column_name"` tags for SQL column mapping
// 2. Consider adding validation tags for database constraints
// 3. Add time.Time fields for created_at and updated_at timestamps
// 4. Implement database-specific field types (UUID for IDs, DECIMAL for quantities)

type InventoryItem struct {
	IngredientID string  `json:"ingredient_id"` // TODO: Add `db:"ingredient_id"`
	Name         string  `json:"name"`          // TODO: Add `db:"name"`
	Quantity     float64 `json:"quantity"`      // TODO: Add `db:"quantity"`
	Unit         string  `json:"unit"`          // TODO: Add `db:"unit"`
	MinThreshold float64 `json:"min_threshold"` // TODO: Add `db:"min_threshold"`
	// TODO: Add database timestamp fields:
	// CreatedAt    time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
