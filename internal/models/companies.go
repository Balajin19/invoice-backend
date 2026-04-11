package models

type Companies struct {
	CompanyID      string `json:"companyId" db:"company_id"`
	CompanyName    string `json:"companyName" db:"company_name"`
	OwnerName      string `json:"ownerName" db:"owner_name"`
	BuildingNumber string `json:"buildingNumber" db:"building_number"`
	Street         string `json:"street" db:"street"`
	City           string `json:"city" db:"city"`
	District       string `json:"district" db:"district"`
	State          string `json:"state" db:"state"`
	Pincode        string `json:"pincode" db:"pincode"`
	GSTIN          string `json:"gstin" db:"gstin"`
	CGSTRate       float64 `json:"cgstRate" db:"cgst_rate"`
	SGSTRate       float64 `json:"sgstRate" db:"sgst_rate"`
	IGSTRate       float64 `json:"igstRate" db:"igst_rate"`
	Email          string `json:"email" db:"email"`
	MobileNumber   string `json:"mobileNumber" db:"mobile_number"`
	IsPrimary      bool   `json:"isPrimary" db:"is_primary"`
	CreatedAt      string `json:"createdAt" db:"created_at"`
	UpdatedAt      string `json:"updatedAt" db:"updated_at"`
	CreatedAtEpoch int64  `json:"createdAtEpoch" db:"created_at_epoch"`
	UpdatedAtEpoch int64  `json:"updatedAtEpoch" db:"updated_at_epoch"`
	CreatedBy      string `json:"createdBy" db:"created_by"`
	UpdatedBy      string `json:"updatedBy" db:"updated_by"`
}
