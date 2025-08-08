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

type InventoryUpdateResult struct {
	IngredientID string  `json:"ingredient_id"`
	Name         string  `json:"name"`
	QuantityUsed float64 `json:"quantity_used"`
	Remaining    float64 `json:"remaining"`
}

type BatchOrderRequest struct {
	Orders []BatchOrderItem `json:"orders"`
}

type BatchOrderItem struct {
	CustomerName string             `json:"customer_name"`
	Items        []BatchOrderItemDetail `json:"items"`
}

type BatchOrderItemDetail struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
}

type BatchProcessResult struct {
	OrderID      string  `json:"order_id"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Total        float64 `json:"total,omitempty"`
	Reason       string  `json:"reason,omitempty"`
}

type BatchProcessSummary struct {
	TotalOrders       int                     `json:"total_orders"`
	Accepted          int                     `json:"accepted"`
	Rejected          int                     `json:"rejected"`
	TotalRevenue      float64                 `json:"total_revenue"`
	InventoryUpdates  []InventoryUpdateResult `json:"inventory_updates"`
}

type BatchProcessResponse struct {
	ProcessedOrders []BatchProcessResult `json:"processed_orders"`
	Summary         BatchProcessSummary  `json:"summary"`
}