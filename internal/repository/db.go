package repository

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLiteDB(filename string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Repository{}, &Deployment{}); err != nil {
		return nil, err
	}
	return db, nil
}
