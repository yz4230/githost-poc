package entity

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrInvalid   = errors.New("invalid entity")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
	ErrInternal  = errors.New("internal error")
)
