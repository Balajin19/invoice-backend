package repository

import "errors"

var (
	ErrDuplicateCategory = errors.New("category already exists")
	ErrDuplicateProduct  = errors.New("product already exists")
	ErrDuplicateCustomer = errors.New("customer already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrSettingsAlreadyExists = errors.New("settings already exist")
)
