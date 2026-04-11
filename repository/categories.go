package repository

import (
	"database/sql"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

func GetAllCategories() ([]models.Category, error) {
	rows, err := config.DB.Query(`SELECT category_id, category_name, created_by, updated_by FROM categories ORDER BY category_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.CategoryId, &c.CategoryName, &c.CreatedBy, &c.UpdatedBy); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func GetCategoryByID(categoryID string) (*models.Category, error) {
	var c models.Category
	err := config.DB.QueryRow(
		`SELECT category_id, category_name, created_by, updated_by FROM categories WHERE category_id = $1`,
		categoryID,
	).Scan(&c.CategoryId, &c.CategoryName, &c.CreatedBy, &c.UpdatedBy)
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
		`INSERT INTO categories (category_id, category_name, created_by, updated_by, created_at_epoch, updated_at_epoch) VALUES (gen_random_uuid(), $1, $2, $2, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT) RETURNING category_id, created_by, updated_by`,
		category.CategoryName, category.CreatedBy,
	).Scan(&category.CategoryId, &category.CreatedBy, &category.UpdatedBy)
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
		`UPDATE categories SET category_name = $1, updated_by = $2, updated_at = CURRENT_TIMESTAMP, updated_at_epoch = EXTRACT(EPOCH FROM NOW())::BIGINT WHERE category_id = $3`,
		category.CategoryName, category.UpdatedBy, categoryID,
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
