package models

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update menu models for database relationships and constraints
// 1. Add database column mappings with `db:"column_name"` tags  
// 2. Implement proper relationship between menu_items and ingredients tables
// 3. Add database constraints for price validation and category enum
// 4. Update timestamp fields to use time.Time with proper database mapping
// 5. Add foreign key relationships and junction tables for ingredients

import "time"

type MenuItem struct {
	ID                   string               `json:"product_id" db:"id"`
	Name                 string               `json:"name" db:"name"`
	Description          string               `json:"description" db:"description"`
	Category             MenuCategory         `json:"category" db:"category"`
	Price                float64              `json:"price" db:"price"`
	Available            bool                 `json:"available" db:"available"`
	Tags                 []string             `json:"tags" db:"tags"`
	Allergens            []string             `json:"allergens" db:"allergens"`
	CustomizationOptions []byte               `json:"customization_options" db:"customization_options"`
	Ingredients          []MenuItemIngredient `json:"ingredients"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at" db:"updated_at"`
}

// TODO: Add additional fields based on README spec:
// Category    MenuCategory `json:"category"`
// Available   bool         `json:"available"`
// CreatedAt   time.Time    `json:"created_at"`
// UpdatedAt   time.Time    `json:"updated_at"`

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace with junction table for menu_item_ingredients
// Create separate table: menu_item_ingredients (menu_item_id, ingredient_id, quantity)
// This will normalize the relationship and allow for better database constraints
type MenuItemIngredient struct {
	IngredientID string  `json:"ingredient_id" db:"ingredient_id"`
	Quantity     float64 `json:"quantity" db:"quantity"`
}

// TODO: Add MenuCategory enum based on README spec
type MenuCategory string

const (
	CategoryCoffee   MenuCategory = "coffee"
	CategoryTea      MenuCategory = "tea"
	CategoryPastry   MenuCategory = "pastry"
	CategorySandwich MenuCategory = "sandwich"
	CategoryDrink    MenuCategory = "drink"
)

// TODO: Add aggregation models based on README spec
type PopularItemAggregation struct {
	ItemID       string    `json:"item_id"`
	ItemName     string    `json:"item_name"`
	OrderCount   int       `json:"order_count"`
	TotalRevenue float64   `json:"total_revenue"`
	Rank         int       `json:"rank"`
	LastOrdered  time.Time `json:"last_ordered"`
}
