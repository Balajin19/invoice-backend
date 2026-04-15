package repository

import "errors"

var (
	ErrDuplicateCategory = errors.New("category already exists")
	ErrDuplicateProduct  = errors.New("product already exists")
	ErrDuplicateUnit     = errors.New("unit already exists")
	ErrUnitNotFound      = errors.New("unit not found")
	ErrUnitInUse         = errors.New("unit is in use by products")
	ErrDuplicateCustomer = errors.New("customer already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrSettingsAlreadyExists = errors.New("settings already exist")
)
