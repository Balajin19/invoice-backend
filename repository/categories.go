package repository

import (
	"database/sql"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

func GetAllCategories() ([]models.Category, error) {
	rows, err := config.DB.Query(`SELECT category_id, category_name FROM categories ORDER BY category_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.CategoryId, &c.CategoryName); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func GetCategoryByID(categoryID string) (*models.Category, error) {
	var c models.Category
	err := config.DB.QueryRow(
		`SELECT category_id, category_name FROM categories WHERE category_id = $1`,
		categoryID,
	).Scan(&c.CategoryId, &c.CategoryName)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateCategory(category models.Category) (*models.Category, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE lower(trim(category_name)) = lower(trim($1))
		)`,
		category.CategoryName,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateCategory
	}

	err = config.DB.QueryRow(
		`INSERT INTO categories (category_id, category_name) VALUES (gen_random_uuid(), $1) RETURNING category_id`,
		category.CategoryName,
	).Scan(&category.CategoryId)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func UpdateCategory(categoryID string, category models.Category) (*models.Category, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE lower(trim(category_name)) = lower(trim($1))
			  AND category_id <> $2
		)`,
		category.CategoryName,
		categoryID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateCategory
	}

	result, err := config.DB.Exec(
		`UPDATE categories SET category_name = $1 WHERE category_id = $2`,
		category.CategoryName, categoryID,
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
	category.CategoryId = categoryID
	return &category, nil
}

func DeleteCategory(categoryID string) error {
	result, err := config.DB.Exec(`DELETE FROM categories WHERE category_id = $1`, categoryID)
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
