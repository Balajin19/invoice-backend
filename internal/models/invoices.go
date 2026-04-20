package models

import "time"

// InvoiceProduct represents an invoice product row in API responses.
type InvoiceProduct struct {
	ID          int     `json:"id" db:"id"`
	InvoiceId   string  `json:"invoiceId" db:"invoice_id"`
	ProductId   string  `json:"productId" db:"product_id"`
	ProductName string  `json:"productName" db:"product_name"`
	HSN         string  `json:"hsnSac" db:"hsn_sac"`
	Unit        string  `json:"unit" db:"unit"`
	Qty         float64 `json:"qty" db:"qty"`
	Price       float64 `json:"price" db:"price"`
	Discount    float64 `json:"discount" db:"discount"`
	Total       float64 `json:"total" db:"total"`
	CGSTRate    float64 `json:"cgstRate" db:"cgst_rate"`
	SGSTRate    float64 `json:"sgstRate" db:"sgst_rate"`
	IGSTRate    float64 `json:"igstRate" db:"igst_rate"`
}

// Invoice represents an invoice in the system
type Invoice struct {
	InvoiceNumber string    `json:"invoiceNumber" db:"invoice_number"`
	InvoiceDate   time.Time `json:"invoiceDate" db:"invoice_date"`
	PONumber      string    `json:"poNumber" db:"po_number"`
	PODate        *time.Time `json:"poDate" db:"po_date"`
	InvoiceId     string    `json:"invoiceId" db:"invoice_id"`
	CompanyId     string    `json:"companyId" db:"company_id"`
	CustomerId    string    `json:"customerId" db:"customer_id"`
	CustomerName  string    `json:"customerName" db:"customer_name"`
	Address       string    `json:"customerAddress" db:"customer_address"`
	GSTIN         string    `json:"gstIn" db:"gstin"`
	PaymentTerms  string    `json:"paymentTerms" db:"payment_terms"`
	Amount        float64   `json:"subTotal" db:"sub_total"`
	OverallDiscount float64 `json:"overallDiscount" db:"overall_discount"`
	CGST          float64   `json:"cgst" db:"cgst"`
	SGST          float64   `json:"sgst" db:"sgst"`
	IGST          float64   `json:"igst" db:"igst"`
	RoundedOff    float64   `json:"roundedOff" db:"rounded_off"`
	TotalTax      float64   `json:"totalTax" db:"total_tax"`
	Total         float64   `json:"totalAmount" db:"total_amount"`
	TotalInWords  string    `json:"amountInWords" db:"amount_in_words"`
	Products      []InvoiceProduct `json:"products"`
}
