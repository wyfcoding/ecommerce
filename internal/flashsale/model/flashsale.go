package model

import "time"

// FlashSaleEvent represents a flash sale event in the business layer.
type FlashSaleEvent struct {
	ID          uint
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Status      string // e.g., UPCOMING, ACTIVE, ENDED
	Products    []*FlashSaleProduct
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FlashSaleProduct represents a product within a flash sale event in the business layer.
type FlashSaleProduct struct {
	ID             uint
	EventID        uint
	ProductID      string
	FlashPrice     float64
	TotalStock     int32
	RemainingStock int32
	MaxPerUser     int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}