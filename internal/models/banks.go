package models

type Banks struct {
	ID            string `json:"id" db:"id"`
	BankName      string `json:"bankName" db:"bank_name"`
	AccountName   string `json:"accountName" db:"account_name"`
	AccountNumber string `json:"accountNumber" db:"account_number"`
	IFSC          string `json:"ifsc" db:"ifsc"`
	UPI           string `json:"upi" db:"upi"`
	BranchName    string `json:"branch" db:"branch_name"`
	IsPrimary     bool   `json:"isPrimary" db:"is_primary"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
	UpdatedAt     string `json:"updatedAt" db:"updated_at"`
	CreatedAtEpoch int64 `json:"createdAtEpoch" db:"created_at_epoch"`
	UpdatedAtEpoch int64 `json:"updatedAtEpoch" db:"updated_at_epoch"`
	CreatedBy     string `json:"createdBy" db:"created_by"`
	UpdatedBy     string `json:"updatedBy" db:"updated_by"`
}
