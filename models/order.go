package models

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update models to support database relationships and constraints
// 1. Add database column mappings with `db:"column_name"` tags
// 2. Convert Items slice to separate OrderItems table with foreign key relationship
// 3. Add proper timestamp fields using time.Time instead of string
// 4. Implement database constraints for order status validation
// 5. Add foreign key relationships to customer and menu_item tables

type Order struct {
	ID           string      `json:"order_id"`      // TODO: Add `db:"id"` and use UUID type
	CustomerName string      `json:"customer_name"` // TODO: Add `db:"customer_name"` or foreign key to customers table
	Items        []OrderItem `json:"items"`         // TODO: Replace with separate order_items table relationship
	Status       string      `json:"status"`        // TODO: Add `db:"status"` and enum constraint
	CreatedAt    string      `json:"created_at"`    // TODO: Change to time.Time with `db:"created_at"`
	// TODO: Add database timestamp fields:
	// UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	// TotalAmount  float64   `json:"total_amount" db:"total_amount"`
}

type OrderItem struct {
	ProductID string `json:"product_id"` // TODO: Add `db:"product_id"` with foreign key to menu_items
	Quantity  int    `json:"quantity"`   // TODO: Add `db:"quantity"`
	// TODO: Add database fields for order_items table:
	// ID       string  `json:"id" db:"id"`
	// OrderID  string  `json:"order_id" db:"order_id"` // Foreign key to orders table
	// Price    float64 `json:"price" db:"price"`       // Price at time of order
}
