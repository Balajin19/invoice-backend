package models

// Product represents a product in the system.
type Product struct {
	ProductId    string `json:"productId" db:"product_id"`
	ProductName  string `json:"productName" db:"product_name"`
	HSNSAC       string `json:"hsnSac" db:"hsn_sac"`
	Unit         string `json:"unit" db:"unit"`
	CategoryId   string `json:"categoryId" db:"category_id"`
	CategoryName string `json:"categoryName" db:"category_name"`
}
