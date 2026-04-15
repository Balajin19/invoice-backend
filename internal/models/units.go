package models

// Unit represents a measurable unit option in the system.
type Unit struct {
	UnitID    string `json:"unitId" db:"unit_id"`
	UnitName  string `json:"unitName" db:"unit_name"`
	CreatedBy string `json:"createdBy" db:"created_by"`
	UpdatedBy string `json:"updatedBy" db:"updated_by"`
}
