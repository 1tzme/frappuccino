package models

// TODO: Add import when implementing time-based fields:
import "time"

type MenuItem struct {
	ID          string               `json:"product_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    MenuCategory         `json:"category"`
	Price       float64              `json:"price"`
	Available   bool                 `json:"available"`
	Ingredients []MenuItemIngredient `json:"ingredients"`
}

// TODO: Add additional fields based on README spec:
// Category    MenuCategory `json:"category"`
// Available   bool         `json:"available"`
// CreatedAt   time.Time    `json:"created_at"`
// UpdatedAt   time.Time    `json:"updated_at"`

type MenuItemIngredient struct {
	IngredientID string  `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
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
