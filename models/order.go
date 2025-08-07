package models

import "time"

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update models to support database relationships and constraints
// 1. Add database column mappings with `db:"column_name"` tags
// 2. Convert Items slice to separate OrderItems table with foreign key relationship
// 3. Add proper timestamp fields using time.Time instead of string
// 4. Implement database constraints for order status validation
// 5. Add foreign key relationships to customer and menu_item tables

type Order struct {
	ID                  string      `json:"order_id" db:"id"`
	CustomerName        string      `json:"customer_name" db:"customer_name"`
	Items               []OrderItem `json:"items"`
	Status              string      `json:"status" db:"status"`
	TotalAmount         float64     `json:"total_amount" db:"total_amount"`
	SpecialInstructions string      `json:"special_instructions"`
	CreatedAt           time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID             string  `json:"id" db:"id"`
	OrderID        string  `json:"order_id" db:"order_id"`
	MenuItemID     string  `json:"menu_item_id" db:"menu_item_id"`
	ProductID      string  `json:"product_id" db:"product_id"`
	Quantity       int     `json:"quantity" db:"quantity"`
	PriceAtTime    float64 `json:"price_at_time" db:"price_at_time"`
	Customizations string  `json:"customizations"`
}
