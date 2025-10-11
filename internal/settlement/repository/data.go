package data

import (
	"gorm.io/gorm"
)

// Data struct contains the database connection.
type Data struct {
	db *gorm.DB
}

// NewData creates a new Data struct.
func NewData(db *gorm.DB) (*Data, func(), error) {
	cleanup := func() {
		// No-op cleanup
	}
	return &Data{db: db}, cleanup, nil
}
