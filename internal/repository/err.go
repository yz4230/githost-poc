package repository

import "gorm.io/gorm"

var (
	ErrNotFound  = gorm.ErrRecordNotFound
	ErrDuplicate = gorm.ErrDuplicatedKey
)
