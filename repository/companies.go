package repository

import (
	"database/sql"
	"strings"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

const companySelect = `
	SELECT
		company_id, company_name, COALESCE(owner_name, ''),
		COALESCE(building_number, ''), COALESCE(street, ''), COALESCE(city, ''),
		COALESCE(district, ''), COALESCE(state, ''), COALESCE(pincode, ''),
		COALESCE(gstin, ''), COALESCE(cgst_rate, 0), COALESCE(sgst_rate, 0), COALESCE(igst_rate, 0), COALESCE(email, ''), COALESCE(mobile_number, ''),
		COALESCE(is_primary, false),
		created_at::text, updated_at::text,
		COALESCE(created_at_epoch, 0), COALESCE(updated_at_epoch, 0),
		COALESCE(created_by, ''), COALESCE(updated_by, '')
	FROM companies
`

func scanCompany(row interface{ Scan(...any) error }) (models.Companies, error) {
	var company models.Companies
	err := row.Scan(
		&company.CompanyID, &company.CompanyName, &company.OwnerName,
		&company.BuildingNumber, &company.Street, &company.City,
		&company.District, &company.State, &company.Pincode,
		&company.GSTIN, &company.CGSTRate, &company.SGSTRate, &company.IGSTRate, &company.Email, &company.MobileNumber,
		&company.IsPrimary,
		&company.CreatedAt, &company.UpdatedAt,
		&company.CreatedAtEpoch, &company.UpdatedAtEpoch,
		&company.CreatedBy, &company.UpdatedBy,
	)
	return company, err
}

func GetCompanySettings() ([]models.Companies, error) {
	rows, err := config.DB.Query(companySelect + ` ORDER BY company_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	companies := make([]models.Companies, 0)
	for rows.Next() {
		company, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, rows.Err()
}

func GetCompanySettingsByID(id string) (*models.Companies, error) {
	company, err := scanCompany(config.DB.QueryRow(companySelect+` WHERE company_id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func defaultInvoicePrefix(companyName string) string {
	firstWord := strings.TrimSpace(companyName)
	if parts := strings.Fields(companyName); len(parts) > 0 {
		firstWord = parts[0]
	}
	firstWord = strings.ToUpper(firstWord)
	if firstWord == "" {
		return "INV"
	}
	if len(firstWord) > 3 {
		return firstWord[:3]
	}
	return firstWord
}

func CreateCompanySettings(userEmail string, payload models.Companies) (*models.Companies, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRow(`
		INSERT INTO companies (
			company_name, owner_name, building_number, street, city, district,
			state, pincode, gstin, cgst_rate, sgst_rate, igst_rate, email, mobile_number, is_primary,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16)
		RETURNING company_id`,
		payload.CompanyName, payload.OwnerName, payload.BuildingNumber, payload.Street,
		payload.City, payload.District, payload.State, payload.Pincode,
		payload.GSTIN, payload.CGSTRate, payload.SGSTRate, payload.IGSTRate, payload.Email, payload.MobileNumber, payload.IsPrimary, userEmail,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO invoice_settings (
			company_id, invoice_prefix, start_number, current_number, pad_length, terms_conditions,
			created_by, updated_by
		) VALUES (
			$1, $2, 1, 0, 3, $3,
			$4, $4
		)`,
		id,
		defaultInvoicePrefix(payload.CompanyName),
		"1. Dispute if any shall be subject to Chennai Jurisdiction\n2. Goods once sold will not be taken back\n3. Payment Terms : {{payment_terms}}",
		userEmail,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetCompanySettingsByID(id)
}

func UpdateCompanySettingsByID(id, userEmail string, payload models.Companies) (*models.Companies, error) {
	result, err := config.DB.Exec(`
		UPDATE companies
		SET
			company_name = $1, owner_name = $2, building_number = $3, street = $4,
			city = $5, district = $6, state = $7, pincode = $8,
			gstin = $9, cgst_rate = $10, sgst_rate = $11, igst_rate = $12, email = $13, mobile_number = $14, is_primary = $15,
			updated_by = $16
		WHERE company_id = $17`,
		payload.CompanyName, payload.OwnerName, payload.BuildingNumber, payload.Street,
		payload.City, payload.District, payload.State, payload.Pincode,
		payload.GSTIN, payload.CGSTRate, payload.SGSTRate, payload.IGSTRate, payload.Email, payload.MobileNumber, payload.IsPrimary, userEmail, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return GetCompanySettingsByID(id)
}

func DeleteCompanySettingsByID(id string) error {
	result, err := config.DB.Exec(`DELETE FROM companies WHERE company_id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
