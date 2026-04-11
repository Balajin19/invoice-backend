package repository

import (
	"database/sql"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/pkg/crypto"
)

const bankSelect = `
	SELECT
		id, bank_name,
		COALESCE(account_name, ''), COALESCE(account_number, ''),
		COALESCE(ifsc, ''), COALESCE(upi, ''), COALESCE(branch_name, ''),
		COALESCE(is_primary, false),
		created_at::text, updated_at::text,
		COALESCE(created_at_epoch, 0), COALESCE(updated_at_epoch, 0),
		COALESCE(created_by, ''), COALESCE(updated_by, '')
	FROM banks
`

func scanBank(row interface{ Scan(...any) error }) (models.Banks, error) {
	var bank models.Banks
	err := row.Scan(
		&bank.ID, &bank.BankName,
		&bank.AccountName, &bank.AccountNumber,
		&bank.IFSC, &bank.UPI, &bank.BranchName,
		&bank.IsPrimary,
		&bank.CreatedAt, &bank.UpdatedAt,
		&bank.CreatedAtEpoch, &bank.UpdatedAtEpoch,
		&bank.CreatedBy, &bank.UpdatedBy,
	)
	if err != nil {
		return bank, err
	}
	if bank.AccountNumber != "" {
		decrypted, decErr := crypto.Decrypt(bank.AccountNumber)
		if decErr == nil {
			bank.AccountNumber = decrypted
		}
	}
	return bank, nil
}

func GetBankSettings() ([]models.Banks, error) {
	rows, err := config.DB.Query(bankSelect + ` ORDER BY bank_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	banks := make([]models.Banks, 0)
	for rows.Next() {
		bank, err := scanBank(rows)
		if err != nil {
			return nil, err
		}
		banks = append(banks, bank)
	}
	return banks, rows.Err()
}

func GetBankSettingsByID(id string) (*models.Banks, error) {
	bank, err := scanBank(config.DB.QueryRow(bankSelect+` WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &bank, nil
}

func CreateBankSettings(userEmail string, payload models.Banks) (*models.Banks, error) {
	encryptedAccount, err := crypto.Encrypt(payload.AccountNumber)
	if err != nil {
		return nil, err
	}
	var id string
	err = config.DB.QueryRow(`
		INSERT INTO banks (
			bank_name, account_name, account_number, ifsc, upi, branch_name, is_primary,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id`,
		payload.BankName, payload.AccountName, encryptedAccount,
		payload.IFSC, payload.UPI, payload.BranchName, payload.IsPrimary, userEmail,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetBankSettingsByID(id)
}

func UpdateBankSettingsByID(id, userEmail string, payload models.Banks) (*models.Banks, error) {
	encryptedAccount, err := crypto.Encrypt(payload.AccountNumber)
	if err != nil {
		return nil, err
	}
	result, err := config.DB.Exec(`
		UPDATE banks
		SET
			bank_name = $1, account_name = $2, account_number = $3,
			ifsc = $4, upi = $5, branch_name = $6, is_primary = $7,
			updated_by = $8
		WHERE id = $9`,
		payload.BankName, payload.AccountName, encryptedAccount,
		payload.IFSC, payload.UPI, payload.BranchName, payload.IsPrimary, userEmail, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return GetBankSettingsByID(id)
}

func DeleteBankSettingsByID(id string) error {
	result, err := config.DB.Exec(`DELETE FROM banks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
