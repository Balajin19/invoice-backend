package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
	"strconv"
	"strings"
	"time"
)

func nullableUUID(value string) sql.NullString {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func parseInvoiceNumber(invoiceNumber string) (string, int, error) {
	parts := strings.Split(strings.TrimSpace(invoiceNumber), "/")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid invoice number format: %s", invoiceNumber)
	}

	prefix := strings.TrimSpace(parts[0])
	sequencePart := strings.TrimSpace(parts[1])
	if prefix == "" || sequencePart == "" {
		return "", 0, fmt.Errorf("invalid invoice number format: %s", invoiceNumber)
	}

	sequence, err := strconv.Atoi(sequencePart)
	if err != nil {
		return "", 0, fmt.Errorf("invalid invoice number sequence: %w", err)
	}

	return prefix, sequence, nil
}

func resolveInvoiceProductUnit(tx *sql.Tx, product models.InvoiceProduct) (string, error) {
	unit := strings.TrimSpace(product.Unit)
	productID := strings.TrimSpace(product.ProductId)
	if productID == "" {
		return unit, nil
	}

	var resolvedUnit string
	err := tx.QueryRow(`
		SELECT COALESCE(u.unit_name, '')
		FROM products p
		LEFT JOIN units u ON u.unit_id = p.unit_id
		WHERE p.product_id = $1
	`, productID).Scan(&resolvedUnit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return unit, nil
		}
		return "", err
	}

	resolvedUnit = strings.TrimSpace(resolvedUnit)
	if resolvedUnit == "" {
		return unit, nil
	}

	return resolvedUnit, nil
}

func syncInvoiceSettingsCurrentNumber(tx *sql.Tx, userEmail, companyID, invoiceNumber string, invoiceDate time.Time) error {
	prefix, sequence, err := parseInvoiceNumber(invoiceNumber)
	if err != nil {
		return err
	}
	financialYear := normalizeFinancialYear("", invoiceDate)

	nullCompanyID := nullableUUID(companyID)
	result, err := tx.Exec(`
		WITH target AS (
			SELECT id
			FROM invoice_settings
			WHERE TRIM(LOWER(invoice_prefix)) = TRIM(LOWER($1))
			  AND (
				($2::uuid IS NOT NULL AND company_id = $2::uuid)
				OR
				($2::uuid IS NULL)
			  )
			ORDER BY created_at DESC
			LIMIT 1
		)
		UPDATE invoice_settings AS invoice_setting
		SET
			current_number = CASE
				WHEN TRIM(COALESCE(invoice_setting.financial_year, '')) <> ''
				     AND TRIM(invoice_setting.financial_year) <> TRIM($3)
				THEN GREATEST(COALESCE(invoice_setting.start_number, 1) - 1, $4)
				ELSE GREATEST(COALESCE(invoice_setting.current_number, 0), $4)
			END,
			financial_year = $3,
			updated_by = $5
		FROM target
		WHERE invoice_setting.id = target.id`,
		prefix, nullCompanyID, financialYear, sequence, userEmail,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func replaceInvoiceProducts(tx *sql.Tx, invoiceID string, products []models.InvoiceProduct, userEmail string) ([]models.InvoiceProduct, error) {
	if _, err := tx.Exec(`DELETE FROM invoice_products WHERE invoice_id = $1`, invoiceID); err != nil {
		return nil, err
	}

	updatedProducts := make([]models.InvoiceProduct, len(products))
	copy(updatedProducts, products)

	for i, product := range updatedProducts {
		unitValue, err := resolveInvoiceProductUnit(tx, product)
		if err != nil {
			return nil, err
		}
		err = tx.QueryRow(`
			INSERT INTO invoice_products (invoice_id, product_id, product_name, unit, qty, price, discount, total, cgst_rate, sgst_rate, igst_rate, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
			RETURNING id
		`,
			invoiceID,
			nullableUUID(product.ProductId),
			product.ProductName,
			unitValue,
			product.Qty,
			product.Price,
			product.Discount,
			product.Total,
			product.CGSTRate,
			product.SGSTRate,
			product.IGSTRate,
			userEmail,
		).Scan(&updatedProducts[i].ID)
		if err != nil {
			return nil, err
		}
		updatedProducts[i].InvoiceId = invoiceID
		updatedProducts[i].Unit = unitValue
	}

	return updatedProducts, nil
}

func GetAllInvoices(companyID string) ([]models.Invoice, error) {
	nullCompanyID := nullableUUID(companyID)

	query := `
	SELECT
		i.invoice_number,
		i.invoice_date,
		COALESCE(i.po_number, ''),
		i.po_date,
		i.invoice_id,
		COALESCE(i.company_id::text, ''),
		i.customer_id,
		i.customer_name,
		i.customer_address,
		i.gstin,
		COALESCE(i.payment_terms, ''),
		i.sub_total,
		COALESCE(i.overall_discount, 0),
		i.cgst,
		i.sgst,
		COALESCE(i.igst, 0),
		COALESCE(i.rounded_off, 0),
		COALESCE(i.total_tax, 0),
		i.total_amount,
		i.amount_in_words,
		COALESCE(i.is_active, TRUE),
		COALESCE(
			(
				SELECT json_agg(
					json_build_object(
						'id', ip.id,
						'invoiceId', ip.invoice_id,
						'productId', ip.product_id,
						'productName', ip.product_name,
						'hsnSac', COALESCE(p.hsn_sac, ''),
						'unit', ip.unit,
						'qty', ip.qty,
						'price', ip.price,
						'discount', COALESCE(ip.discount, 0),
						'total', ip.total
					)
					ORDER BY lower(COALESCE(ip.product_name, '')) ASC
				)
				FROM invoice_products ip
				LEFT JOIN products p ON p.product_id = ip.product_id
				WHERE ip.invoice_id = i.invoice_id
			),
			'[]'::json
		) AS products
	FROM invoices i
		WHERE COALESCE(i.is_active, TRUE) = TRUE
		  AND ($1::uuid IS NULL OR i.company_id = $1::uuid)
	ORDER BY invoice_number ASC
	`

	rows, err := config.DB.Query(query, nullCompanyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]models.Invoice, 0)

	for rows.Next() {
		var invoice models.Invoice
		var InvoiceDate time.Time
		var poDate sql.NullTime
		var productsJSON []byte

		err := rows.Scan(
			&invoice.InvoiceNumber,
			&InvoiceDate,
			&invoice.PONumber,
			&poDate,
			&invoice.InvoiceId,
			&invoice.CompanyId,
			&invoice.CustomerId,
			&invoice.CustomerName,
			&invoice.Address,
			&invoice.GSTIN,
			&invoice.PaymentTerms,
			&invoice.Amount,
			&invoice.OverallDiscount,
			&invoice.CGST,
			&invoice.SGST,
			&invoice.IGST,
			&invoice.RoundedOff,
			&invoice.TotalTax,
			&invoice.Total,
			&invoice.TotalInWords,
			&invoice.IsActive,
			&productsJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(productsJSON, &invoice.Products); err != nil {
			return nil, err
		}

		invoice.InvoiceDate = InvoiceDate
		if poDate.Valid {
			parsed := poDate.Time
			invoice.PODate = &parsed
		}
		results = append(results, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func GetInvoiceByID(invoiceID string) (*models.Invoice, error) {
	query := `
	SELECT
		i.invoice_number,
		i.invoice_date,
		COALESCE(i.po_number, ''),
		i.po_date,
		i.invoice_id,
		COALESCE(i.company_id::text, ''),
		i.customer_id,
		i.customer_name,
		i.customer_address,
		i.gstin,
		COALESCE(i.payment_terms, ''),
		i.sub_total,
		COALESCE(i.overall_discount, 0),
		i.cgst,
		i.sgst,
		COALESCE(i.igst, 0),
		COALESCE(i.rounded_off, 0),
		COALESCE(i.total_tax, 0),
		i.total_amount,
		i.amount_in_words,
		COALESCE(i.is_active, TRUE),
		COALESCE(
			(
				SELECT json_agg(
					json_build_object(
						'id', ip.id,
						'invoiceId', ip.invoice_id,
						'productId', ip.product_id,
						'productName', ip.product_name,
						'hsnSac', COALESCE(p.hsn_sac, ''),
						'unit', ip.unit,
						'qty', ip.qty,
						'price', ip.price,
						'discount', COALESCE(ip.discount, 0),
						'total', ip.total,
						'cgstRate', COALESCE(ip.cgst_rate, 0),
						'sgstRate', COALESCE(ip.sgst_rate, 0),
						'igstRate', COALESCE(ip.igst_rate, 0)
					)
					ORDER BY lower(COALESCE(ip.product_name, '')) ASC
				)
				FROM invoice_products ip
				LEFT JOIN products p ON p.product_id = ip.product_id
				WHERE ip.invoice_id = i.invoice_id
			),
			'[]'::json
		) AS products
	FROM invoices i
		WHERE i.invoice_id = $1
		  AND COALESCE(i.is_active, TRUE) = TRUE
	`

	var invoice models.Invoice
	var InvoiceDate time.Time
	var poDate sql.NullTime
	var productsJSON []byte

	err := config.DB.QueryRow(query, invoiceID).Scan(
		&invoice.InvoiceNumber,
		&InvoiceDate,
		&invoice.PONumber,
		&poDate,
		&invoice.InvoiceId,
		&invoice.CompanyId,
		&invoice.CustomerId,
		&invoice.CustomerName,
		&invoice.Address,
		&invoice.GSTIN,
		&invoice.PaymentTerms,
		&invoice.Amount,
		&invoice.OverallDiscount,
		&invoice.CGST,
		&invoice.SGST,
		&invoice.IGST,
		&invoice.RoundedOff,
		&invoice.TotalTax,
		&invoice.Total,
		&invoice.TotalInWords,
		&invoice.IsActive,
		&productsJSON,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(productsJSON, &invoice.Products); err != nil {
		return nil, err
	}

	invoice.InvoiceDate = InvoiceDate
	if poDate.Valid {
		parsed := poDate.Time
		invoice.PODate = &parsed
	}
	return &invoice, nil
}

func CreateInvoice(userEmail string, invoice models.Invoice) (*models.Invoice, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO invoices (
		invoice_id, invoice_number, invoice_date, company_id, customer_id, customer_name,
		customer_address, gstin, payment_terms, po_number, po_date, sub_total, cgst, sgst, igst,
		overall_discount, rounded_off, total_tax, total_amount, amount_in_words, created_by, updated_by
	)
	VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $20)
	RETURNING invoice_id
	`
	err = tx.QueryRow(query,
		invoice.InvoiceNumber,
		invoice.InvoiceDate,
		nullableUUID(invoice.CompanyId),
		invoice.CustomerId,
		invoice.CustomerName,
		invoice.Address,
		invoice.GSTIN,
		invoice.PaymentTerms,
		invoice.PONumber,
		invoice.PODate,
		invoice.Amount,
		invoice.OverallDiscount,
		invoice.CGST,
		invoice.SGST,
		invoice.IGST,
		invoice.RoundedOff,
		invoice.TotalTax,
		invoice.Total,
		invoice.TotalInWords,
		userEmail,
	).Scan(&invoice.InvoiceId)
	if err != nil {
		return nil, err
	}

	for i, product := range invoice.Products {
		unitValue, err := resolveInvoiceProductUnit(tx, product)
		if err != nil {
			return nil, err
		}
		productQuery := `
		INSERT INTO invoice_products (invoice_id, product_id, product_name, unit, qty, price, discount, total, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
		`
		err = tx.QueryRow(productQuery,
			invoice.InvoiceId,
			nullableUUID(product.ProductId),
			product.ProductName,
			unitValue,
			product.Qty,
			product.Price,
			product.Discount,
			product.Total,
			userEmail,
		).Scan(&invoice.Products[i].ID)
		if err != nil {
			return nil, err
		}
		invoice.Products[i].InvoiceId = invoice.InvoiceId
		invoice.Products[i].Unit = unitValue
	}

	if err = syncInvoiceSettingsCurrentNumber(tx, userEmail, invoice.CompanyId, invoice.InvoiceNumber, invoice.InvoiceDate); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &invoice, nil
}

func UpdateInvoice(invoiceID string, invoice models.Invoice, userEmail string) (*models.Invoice, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
	UPDATE invoices
	SET invoice_number   = $1,
	    invoice_date     = $2,
	    company_id       = $3,
	    customer_id      = $4,
	    customer_name    = $5,
	    customer_address = $6,
	    gstin            = $7,
	    payment_terms    = $8,
	    po_number        = $9,
	    po_date          = $10,
	    sub_total        = $11,
	    overall_discount = $12,
	    cgst             = $13,
	    sgst             = $14,
	    igst             = $15,
	    rounded_off      = $16,
	    total_tax        = $17,
	    total_amount     = $18,
	    amount_in_words  = $19,
	    updated_by       = $20
	WHERE invoice_id = $21
	`
	result, err := tx.Exec(query,
		invoice.InvoiceNumber,
		invoice.InvoiceDate,
		nullableUUID(invoice.CompanyId),
		invoice.CustomerId,
		invoice.CustomerName,
		invoice.Address,
		invoice.GSTIN,
		invoice.PaymentTerms,
		invoice.PONumber,
		invoice.PODate,
		invoice.Amount,
		invoice.OverallDiscount,
		invoice.CGST,
		invoice.SGST,
		invoice.IGST,
		invoice.RoundedOff,
		invoice.TotalTax,
		invoice.Total,
		invoice.TotalInWords,
		userEmail,
		invoiceID,
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	invoice.Products, err = replaceInvoiceProducts(tx, invoiceID, invoice.Products, userEmail)
	if err != nil {
		return nil, err
	}

	if err = syncInvoiceSettingsCurrentNumber(tx, userEmail, invoice.CompanyId, invoice.InvoiceNumber, invoice.InvoiceDate); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	invoice.InvoiceId = invoiceID
	return &invoice, nil
}

func DeleteInvoice(invoiceID string) error {
	result, err := config.DB.Exec(`
		UPDATE invoices
		SET is_active = FALSE
		WHERE invoice_id = $1
		  AND COALESCE(is_active, TRUE) = TRUE
	`, invoiceID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}