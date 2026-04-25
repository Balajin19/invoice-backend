package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"

	"github.com/lib/pq"
)

// currentFinancialYear returns the YY-YY string for today using Apr-Mar boundaries.
func currentFinancialYear() string {
	now := time.Now()
	start := now.Year()
	if now.Month() < time.April {
		start--
	}
	return fmt.Sprintf("%02d-%02d", start%100, (start+1)%100)
}

const invoiceSettingsSelect = `
	SELECT
		id, COALESCE(company_id::text, ''), invoice_prefix, COALESCE(financial_year, ''), start_number, COALESCE(current_number, 0), pad_length,
		COALESCE(terms_conditions, ''),
		created_at::text, updated_at::text,
		COALESCE(created_at_epoch, 0), COALESCE(updated_at_epoch, 0),
		COALESCE(created_by, ''), COALESCE(updated_by, '')
	FROM invoice_settings
`

func scanInvoiceSettings(row interface{ Scan(...any) error }) (models.InvoiceSettings, error) {
	var s models.InvoiceSettings
	err := row.Scan(
		&s.ID, &s.CompanyID, &s.InvoicePrefix, &s.FinancialYear, &s.StartNumber, &s.CurrentNumber, &s.PadLength, &s.TermsConditions,
		&s.CreatedAt, &s.UpdatedAt, &s.CreatedAtEpoch, &s.UpdatedAtEpoch,
		&s.CreatedBy, &s.UpdatedBy,
	)
	return s, err
}

func GetInvoiceSettings() ([]models.InvoiceSettings, error) {
	rows, err := config.DB.Query(invoiceSettingsSelect + ` ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make([]models.InvoiceSettings, 0)
	for rows.Next() {
		s, scanErr := scanInvoiceSettings(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		settings = append(settings, s)
	}

	return settings, rows.Err()
}

func GetInvoiceSettingsByCompanyID(companyID string) ([]models.InvoiceSettings, error) {
	rows, err := config.DB.Query(invoiceSettingsSelect+` WHERE company_id = $1 ORDER BY created_at ASC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make([]models.InvoiceSettings, 0)
	for rows.Next() {
		s, scanErr := scanInvoiceSettings(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		settings = append(settings, s)
	}

	return settings, rows.Err()
}

func GetInvoiceSettingsByID(id string) (*models.InvoiceSettings, error) {
	s, err := scanInvoiceSettings(config.DB.QueryRow(invoiceSettingsSelect+` WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetInvoiceSettingsByCompanyIDSingle(companyID string) (*models.InvoiceSettings, error) {
	s, err := scanInvoiceSettings(config.DB.QueryRow(invoiceSettingsSelect+` WHERE company_id = $1 ORDER BY created_at DESC LIMIT 1`, companyID))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateInvoiceSettings(userEmail string, payload models.InvoiceSettings) (*models.InvoiceSettings, error) {
	var id string
	financialYear := strings.TrimSpace(payload.FinancialYear)
	err := config.DB.QueryRow(`
		INSERT INTO invoice_settings (
			company_id, invoice_prefix, financial_year, start_number, current_number, pad_length, terms_conditions,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id`,
		payload.CompanyID, payload.InvoicePrefix, financialYear, payload.StartNumber, payload.CurrentNumber,
		payload.PadLength, payload.TermsConditions, userEmail,
	).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrDuplicateInvoiceSettingsFY
		}
		return nil, err
	}
	return GetInvoiceSettingsByID(id)
}

func UpdateInvoiceSettings(id, userEmail string, payload models.InvoiceSettings) (*models.InvoiceSettings, error) {
	result, err := config.DB.Exec(`
		UPDATE invoice_settings
		SET
			company_id = $1, invoice_prefix = $2, financial_year = $3, start_number = $4, current_number = $5, pad_length = $6,
			terms_conditions = $7, updated_by = $8
		WHERE id = $9`,
		payload.CompanyID, payload.InvoicePrefix, strings.TrimSpace(payload.FinancialYear), payload.StartNumber, payload.CurrentNumber, payload.PadLength,
		payload.TermsConditions, userEmail, id,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrDuplicateInvoiceSettingsFY
		}
		return nil, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return GetInvoiceSettingsByID(id)
}

func UpdateInvoiceSettingsByCompanyID(companyID, userEmail string, payload models.InvoiceSettings) (*models.InvoiceSettings, error) {
	settings, err := GetInvoiceSettingsByCompanyIDSingle(companyID)
	if err != nil {
		return nil, err
	}

	return UpdateInvoiceSettings(settings.ID, userEmail, payload)
}

func DeleteInvoiceSettingsByID(id string) error {
	result, err := config.DB.Exec(`DELETE FROM invoice_settings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
