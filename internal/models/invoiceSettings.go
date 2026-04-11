package models

type InvoiceSettings struct {
	ID               string `json:"id" db:"id"`
	CompanyID        string `json:"companyId" db:"company_id"`
	InvoicePrefix    string `json:"invoicePrefix" db:"invoice_prefix"`
	StartNumber      int    `json:"startNumber" db:"start_number"`
	CurrentNumber    int    `json:"currentNumber" db:"current_number"`
	PadLength        int    `json:"padLength" db:"pad_length"`
	TermsConditions  string `json:"termsConditions" db:"terms_conditions"`
	CreatedAt        string `json:"createdAt" db:"created_at"`
	UpdatedAt        string `json:"updatedAt" db:"updated_at"`
	CreatedAtEpoch   int64  `json:"createdAtEpoch" db:"created_at_epoch"`
	UpdatedAtEpoch   int64  `json:"updatedAtEpoch" db:"updated_at_epoch"`
	CreatedBy        string `json:"createdBy" db:"created_by"`
	UpdatedBy        string `json:"updatedBy" db:"updated_by"`
}
