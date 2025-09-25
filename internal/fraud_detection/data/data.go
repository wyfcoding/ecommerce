package data

import (
	"ecommerce/internal/fraud_detection/biz"
	"log"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
// Using Google's 'wire' tool for dependency injection is a common practice in Go projects.
var ProviderSet = wire.NewSet(NewData, NewFraudDetectionRepo)

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

// NewFraudDetectionRepo is a provider function that creates a new fraud detection repository.
// It depends on the Data struct (which has the db connection).
func NewFraudDetectionRepo(data *Data) biz.FraudDetectionRepo {
	// The actual implementation is in fraud_detection.go
	return &fraudDetectionRepo{
		data: data,
	}
}
