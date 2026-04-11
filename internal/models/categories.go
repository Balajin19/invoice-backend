package models

// Category represents a product category.
type Category struct {
	CategoryId   string `json:"categoryId" db:"category_id"`
	CategoryName string `json:"categoryName" db:"category_name"`
}
