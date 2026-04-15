package models

// CustomerAddress represents the nested address object
type CustomerAddress struct {
	BuildingNo string  `json:"buildingNumber" db:"building_number"`
	Street     *string `json:"street" db:"street"`
	City       string  `json:"city" db:"city"`
	District   string  `json:"district" db:"district"`
	State      string  `json:"state" db:"state"`
	Pincode    string  `json:"pincode" db:"pincode"`
}

// CustomerProduct represents a product associated with a customer
type CustomerProduct struct {
	ProductId   string  `json:"productId" db:"product_id"`
	UnitID      string  `json:"unitId" db:"unit_id"`
	ProductName string  `json:"productName" db:"product_name"`
	Unit        string  `json:"unit" db:"unit"`
	Price       float64 `json:"price" db:"price"`
}

// Customer represents a customer in the system
type Customer struct {
	CustomerId   string          `json:"customerId" db:"customer_id"`
	CustomerName string          `json:"customerName" db:"customer_name"`
	Address      CustomerAddress `json:"address"`
	GSTIN        string          `json:"gstIn" db:"gstin"`
	Products     []CustomerProduct `json:"products"`
}
