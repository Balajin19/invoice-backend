package repository

import (
	"database/sql"
	"strings"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

func resolveUnitNameByID(unitID string) (string, error) {
	trimmed := strings.TrimSpace(unitID)
	if trimmed == "" {
		return "", ErrUnitNotFound
	}

	var unitName string
	err := config.DB.QueryRow(`SELECT COALESCE(unit_name, '') FROM units WHERE unit_id = $1`, trimmed).Scan(&unitName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUnitNotFound
		}
		return "", err
	}

	unitName = strings.TrimSpace(unitName)
	if unitName == "" {
		return "", ErrUnitNotFound
	}

	return unitName, nil
}

func GetAllProducts() ([]models.Product, error) {
	query := `
	SELECT
		p.product_id,
		p.product_name,
		COALESCE(p.hsn_sac, ''),
		p.unit_id,
		COALESCE(u.unit_name, ''),
		p.category_id,
		c.category_name
	FROM products p
	LEFT JOIN categories c ON c.category_id = p.category_id
	LEFT JOIN units u ON u.unit_id = p.unit_id
	ORDER BY p.product_name ASC
	`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		var product models.Product

		err := rows.Scan(
			&product.ProductId,
			&product.ProductName,
			&product.HSNSAC,
			&product.UnitID,
			&product.Unit,
			&product.CategoryId,
			&product.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func GetProductsByCategoryID(categoryID string) ([]models.Product, error) {
	query := `
	SELECT
		p.product_id,
		p.product_name,
		COALESCE(p.hsn_sac, ''),
		p.unit_id,
		COALESCE(u.unit_name, ''),
		p.category_id,
		c.category_name
	FROM products p
	LEFT JOIN categories c ON c.category_id = p.category_id
	LEFT JOIN units u ON u.unit_id = p.unit_id
	WHERE p.category_id = $1
	ORDER BY p.product_name ASC
	`

	rows, err := config.DB.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		var product models.Product

		err := rows.Scan(
			&product.ProductId,
			&product.ProductName,
			&product.HSNSAC,
			&product.UnitID,
			&product.Unit,
			&product.CategoryId,
			&product.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func GetProductByID(productID string) (*models.Product, error) {
	query := `
	SELECT
		p.product_id,
		p.product_name,
		COALESCE(p.hsn_sac, ''),
		p.unit_id,
		COALESCE(u.unit_name, ''),
		p.category_id,
		c.category_name
	FROM products p
	LEFT JOIN categories c ON c.category_id = p.category_id
	LEFT JOIN units u ON u.unit_id = p.unit_id
	WHERE p.product_id = $1
	`

	var product models.Product

	err := config.DB.QueryRow(query, productID).Scan(
		&product.ProductId,
		&product.ProductName,
		&product.HSNSAC,
		&product.UnitID,
		&product.Unit,
		&product.CategoryId,
		&product.CategoryName,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func CreateProduct(product models.Product, actorEmail string) (*models.Product, error) {
	unitName, err := resolveUnitNameByID(product.UnitID)
	if err != nil {
		return nil, err
	}

	var exists bool
	err = config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM products
			WHERE lower(trim(product_name)) = lower(trim($1))
			  AND unit_id = $2
			  AND category_id = $3
		)`,
		product.ProductName,
		product.UnitID,
		product.CategoryId,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateProduct
	}

	query := `
	INSERT INTO products (product_id, product_name, hsn_sac, unit, unit_id, category_id, created_by, updated_by)
	VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $6)
	RETURNING product_id
	`
	err = config.DB.QueryRow(query,
		product.ProductName,
		product.HSNSAC,
		unitName,
		product.UnitID,
		product.CategoryId,
		actorEmail,
	).Scan(&product.ProductId)
	if err != nil {
		return nil, err
	}
	return GetProductByID(product.ProductId)
}

func UpdateProduct(productID string, product models.Product, actorEmail string) (*models.Product, error) {
	unitName, err := resolveUnitNameByID(product.UnitID)
	if err != nil {
		return nil, err
	}

	var exists bool
	err = config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM products
			WHERE lower(trim(product_name)) = lower(trim($1))
			  AND unit_id = $2
			  AND category_id = $3
			  AND product_id <> $4
		)`,
		product.ProductName,
		product.UnitID,
		product.CategoryId,
		productID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateProduct
	}

	query := `
	UPDATE products
	SET product_name = $1,
	    hsn_sac      = $2,
	    unit         = $3,
	    unit_id      = $4,
	    category_id  = $5,
	    updated_by   = $6
	WHERE product_id = $7
	`
	result, err := config.DB.Exec(query,
		product.ProductName,
		product.HSNSAC,
		unitName,
		product.UnitID,
		product.CategoryId,
		actorEmail,
		productID,
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
	return GetProductByID(productID)
}

func DeleteProduct(productID string) error {
	result, err := config.DB.Exec(`DELETE FROM products WHERE product_id = $1`, productID)
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
