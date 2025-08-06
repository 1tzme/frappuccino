package models

// TODO: Transition State: JSON → PostgreSQL
// ✅ COMPLETED: Repository now uses PostgreSQL inventory table

type InventoryItem struct {
	IngredientID string  `json:"ingredient_id"` // Maps to inventory.id (UUID)
	Name         string  `json:"name"`          // Maps to inventory.name (VARCHAR)
	Quantity     float64 `json:"quantity"`      // Maps to inventory.quantity (DECIMAL)
	Unit         string  `json:"unit"`          // Maps to inventory.unit (unit_type ENUM)
	MinThreshold float64 `json:"min_threshold"` // Maps to inventory.min_threshold (DECIMAL)
}
