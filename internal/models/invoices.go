package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type fixed2 float64

func (f fixed2) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', 2, 64)), nil
}

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

func (p InvoiceProduct) MarshalJSON() ([]byte, error) {
	type invoiceProductAlias InvoiceProduct
	return json.Marshal(struct {
		invoiceProductAlias
		Qty      fixed2 `json:"qty"`
		Price    fixed2 `json:"price"`
		Discount fixed2 `json:"discount"`
		Total    fixed2 `json:"total"`
		CGSTRate fixed2 `json:"cgstRate"`
		SGSTRate fixed2 `json:"sgstRate"`
		IGSTRate fixed2 `json:"igstRate"`
	}{
		invoiceProductAlias: invoiceProductAlias(p),
		Qty:                 fixed2(p.Qty),
		Price:               fixed2(p.Price),
		Discount:            fixed2(p.Discount),
		Total:               fixed2(p.Total),
		CGSTRate:            fixed2(p.CGSTRate),
		SGSTRate:            fixed2(p.SGSTRate),
		IGSTRate:            fixed2(p.IGSTRate),
	})
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
	IsActive      bool      `json:"isActive" db:"is_active"`
	Products      []InvoiceProduct `json:"products"`
}

func (i Invoice) MarshalJSON() ([]byte, error) {
	type invoiceAlias Invoice
	return json.Marshal(struct {
		invoiceAlias
		Amount          fixed2 `json:"subTotal"`
		OverallDiscount fixed2 `json:"overallDiscount"`
		CGST            fixed2 `json:"cgst"`
		SGST            fixed2 `json:"sgst"`
		IGST            fixed2 `json:"igst"`
		RoundedOff      fixed2 `json:"roundedOff"`
		TotalTax        fixed2 `json:"totalTax"`
		Total           fixed2 `json:"totalAmount"`
	}{
		invoiceAlias:    invoiceAlias(i),
		Amount:          fixed2(i.Amount),
		OverallDiscount: fixed2(i.OverallDiscount),
		CGST:            fixed2(i.CGST),
		SGST:            fixed2(i.SGST),
		IGST:            fixed2(i.IGST),
		RoundedOff:      fixed2(i.RoundedOff),
		TotalTax:        fixed2(i.TotalTax),
		Total:           fixed2(i.Total),
	})
}
