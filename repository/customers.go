package repository

import (
	"database/sql"
	"encoding/json"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

func replaceCustomerProducts(tx *sql.Tx, customerID string, products []models.CustomerProduct) error {
	if _, err := tx.Exec(`DELETE FROM customer_products WHERE customer_id = $1`, customerID); err != nil {
		return err
	}

	for _, p := range products {
		if _, err := tx.Exec(
			`INSERT INTO customer_products (customer_id, product_id, unit_id, price) VALUES ($1, $2, $3, $4)`,
			customerID,
			p.ProductId,
			p.UnitID,
			p.Price,
		); err != nil {
			return err
		}
	}

	return nil
}

func GetAllCustomers() ([]models.Customer, error) {
	query := `
	SELECT
		c.customer_id,
		c.customer_name,
		c.building_number,
		c.street,
		c.city,
		c.district,
		c.state,
		c.pincode,
		c.gstin,
		COALESCE(
			(
				SELECT json_agg(
					json_build_object(
						'productId', cp.product_id,
						'unitId', cp.unit_id,
						'productName', p.product_name,
						'unit', COALESCE(u.unit_name, ''),
						'price', cp.price
					)
				)
				FROM customer_products cp
				LEFT JOIN products p ON p.product_id = cp.product_id
				LEFT JOIN units u ON u.unit_id = cp.unit_id
				WHERE cp.customer_id = c.customer_id
			),
			'[]'::json
		) AS products
	FROM customers c
	`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]models.Customer, 0)

	for rows.Next() {
		var customer models.Customer
		var street sql.NullString
		var productsJSON []byte

		err := rows.Scan(
			&customer.CustomerId,
			&customer.CustomerName,
			&customer.Address.BuildingNo,
			&street,
			&customer.Address.City,
			&customer.Address.District,
			&customer.Address.State,
			&customer.Address.Pincode,
			&customer.GSTIN,
			&productsJSON,
		)
		if err != nil {
			return nil, err
		}

		if street.Valid {
			customer.Address.Street = &street.String
		}

		if err := json.Unmarshal(productsJSON, &customer.Products); err != nil {
			return nil, err
		}

		results = append(results, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
func GetCustomerByID(customerId string) (*models.Customer, error) {
	query := `
	SELECT
		c.customer_id,
		c.customer_name,
		c.building_number,
		c.street,
		c.city,
		c.district,
		c.state,
		c.pincode,
		c.gstin,
		COALESCE(
			(
				SELECT json_agg(
					json_build_object(
						'productId', cp.product_id,
						'unitId', cp.unit_id,
						'productName', p.product_name,
						'unit', COALESCE(u.unit_name, ''),
						'price', cp.price
					)
				)
				FROM customer_products cp
				LEFT JOIN products p ON p.product_id = cp.product_id
				LEFT JOIN units u ON u.unit_id = cp.unit_id
				WHERE cp.customer_id = c.customer_id
			),
			'[]'::json
		) AS products
	FROM customers c
	WHERE c.customer_id = $1
	`

	row := config.DB.QueryRow(query, customerId)

	var customer models.Customer
	var street sql.NullString
	var productsJSON []byte

	err := row.Scan(
		&customer.CustomerId,
		&customer.CustomerName,
		&customer.Address.BuildingNo,
		&street,
		&customer.Address.City,
		&customer.Address.District,
		&customer.Address.State,
		&customer.Address.Pincode,
		&customer.GSTIN,
		&productsJSON,
	)
	if err != nil {
		return nil, err
	}

	if street.Valid {
		customer.Address.Street = &street.String
	}

	if err := json.Unmarshal(productsJSON, &customer.Products); err != nil {
		return nil, err
	}

	return &customer, nil
}

func CreateCustomer(customer models.Customer) (*models.Customer, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM customers
			WHERE lower(trim(customer_name)) = lower(trim($1))
			  AND lower(trim(building_number)) = lower(trim($2))
			  AND lower(trim(coalesce(street, ''))) = lower(trim(coalesce($3, '')))
			  AND lower(trim(city)) = lower(trim($4))
			  AND lower(trim(district)) = lower(trim($5))
			  AND lower(trim(state)) = lower(trim($6))
			  AND lower(trim(pincode)) = lower(trim($7))
		)`,
		customer.CustomerName,
		customer.Address.BuildingNo,
		customer.Address.Street,
		customer.Address.City,
		customer.Address.District,
		customer.Address.State,
		customer.Address.Pincode,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateCustomer
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO customers (customer_id, customer_name, building_number, street, city, district, state, pincode, gstin)
	VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING customer_id
	`
	err = tx.QueryRow(query,
		customer.CustomerName,
		customer.Address.BuildingNo,
		customer.Address.Street,
		customer.Address.City,
		customer.Address.District,
		customer.Address.State,
		customer.Address.Pincode,
		customer.GSTIN,
	).Scan(&customer.CustomerId)
	if err != nil {
		return nil, err
	}

	if err := replaceCustomerProducts(tx, customer.CustomerId, customer.Products); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return GetCustomerByID(customer.CustomerId)
}

func UpdateCustomer(customerId string, customer models.Customer) (*models.Customer, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM customers
			WHERE lower(trim(customer_name)) = lower(trim($1))
			  AND lower(trim(building_number)) = lower(trim($2))
			  AND lower(trim(coalesce(street, ''))) = lower(trim(coalesce($3, '')))
			  AND lower(trim(city)) = lower(trim($4))
			  AND lower(trim(district)) = lower(trim($5))
			  AND lower(trim(state)) = lower(trim($6))
			  AND lower(trim(pincode)) = lower(trim($7))
			  AND customer_id <> $8
		)`,
		customer.CustomerName,
		customer.Address.BuildingNo,
		customer.Address.Street,
		customer.Address.City,
		customer.Address.District,
		customer.Address.State,
		customer.Address.Pincode,
		customerId,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateCustomer
	}

	query := `
	UPDATE customers
	SET customer_name   = $1,
	    building_number = $2,
	    street          = $3,
	    city            = $4,
	    district        = $5,
	    state           = $6,
	    pincode         = $7,
	    gstin           = $8
	WHERE customer_id = $9
	`
	tx, err := config.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(query,
		customer.CustomerName,
		customer.Address.BuildingNo,
		customer.Address.Street,
		customer.Address.City,
		customer.Address.District,
		customer.Address.State,
		customer.Address.Pincode,
		customer.GSTIN,
		customerId,
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

	if err := replaceCustomerProducts(tx, customerId, customer.Products); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return GetCustomerByID(customerId)
}

func DeleteCustomer(customerId string) error {
	result, err := config.DB.Exec(`DELETE FROM customers WHERE customer_id = $1`, customerId)
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