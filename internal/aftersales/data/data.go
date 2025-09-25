package data

import (
	"ecommerce/internal/aftersales/biz"
	"log"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
// Using Google's 'wire' tool for dependency injection is a common practice in Go projects.
var ProviderSet = wire.NewSet(NewData, NewAftersalesRepo)

// Data struct holds the database client.
type Data struct {
	db *gorm.DB
}

// NewData creates a new Data struct with a database connection.
func NewData(db *gorm.DB) (*Data, func(), error) {
	// This cleanup function will be called when the service shuts down.
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("failed to get underlying sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
		log.Println("database connection closed")
	}

	d := &Data{
		db: db,
	}
	return d, cleanup, nil
}

// NewAftersalesRepo is a provider function that creates a new aftersales repository.
// It depends on the Data struct (which has the db connection).
func NewAftersalesRepo(data *Data) biz.AftersalesRepo {
	// The actual implementation is in aftersales.go
	return &aftersalesRepo{
		data: data,
	}
}
