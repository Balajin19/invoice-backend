package repository

import (
	"database/sql"

	"invoice-generator-backend/config"
	"invoice-generator-backend/internal/models"
)

func GetAllUnits() ([]models.Unit, error) {
	rows, err := config.DB.Query(`
		SELECT unit_id, unit_name, COALESCE(created_by, ''), COALESCE(updated_by, '')
		FROM units
		ORDER BY unit_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make([]models.Unit, 0)
	for rows.Next() {
		var unit models.Unit
		if err := rows.Scan(&unit.UnitID, &unit.UnitName, &unit.CreatedBy, &unit.UpdatedBy); err != nil {
			return nil, err
		}
		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return units, nil
}

func GetUnitByID(unitID string) (*models.Unit, error) {
	var unit models.Unit
	err := config.DB.QueryRow(`
		SELECT unit_id, unit_name, COALESCE(created_by, ''), COALESCE(updated_by, '')
		FROM units
		WHERE unit_id = $1
	`, unitID).Scan(&unit.UnitID, &unit.UnitName, &unit.CreatedBy, &unit.UpdatedBy)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func CreateUnit(unit models.Unit, actorEmail string) (*models.Unit, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM units
			WHERE lower(trim(unit_name)) = lower(trim($1))
		)`,
		unit.UnitName,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUnit
	}

	err = config.DB.QueryRow(`
		INSERT INTO units (unit_id, unit_name, created_by, updated_by)
		VALUES (gen_random_uuid(), $1, $2, $2)
		RETURNING unit_id, unit_name, COALESCE(created_by, ''), COALESCE(updated_by, '')
	`, unit.UnitName, actorEmail).Scan(&unit.UnitID, &unit.UnitName, &unit.CreatedBy, &unit.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return &unit, nil
}

func UpdateUnit(unitID string, unit models.Unit, actorEmail string) (*models.Unit, error) {
	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM units
			WHERE lower(trim(unit_name)) = lower(trim($1))
			  AND unit_id <> $2
		)`,
		unit.UnitName,
		unitID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUnit
	}

	result, err := config.DB.Exec(`
		UPDATE units
		SET unit_name = $1,
		    updated_by = $2
		WHERE unit_id = $3
	`, unit.UnitName, actorEmail, unitID)
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

	return GetUnitByID(unitID)
}

func DeleteUnit(unitID string) error {
	var inUse bool
	err := config.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM products WHERE unit_id = $1
		)
	`, unitID).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse {
		return ErrUnitInUse
	}

	result, err := config.DB.Exec(`DELETE FROM units WHERE unit_id = $1`, unitID)
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
