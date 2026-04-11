package models

// Category represents a product category.
type Category struct {
	CategoryId   string `json:"categoryId" db:"category_id"`
	CategoryName string `json:"categoryName" db:"category_name"`
	CreatedBy    string `json:"createdBy" db:"created_by"`
	UpdatedBy    string `json:"updatedBy" db:"updated_by"`
}
