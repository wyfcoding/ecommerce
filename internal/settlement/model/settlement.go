package model

import "time"

// SettlementRecord represents a settlement record in the business logic layer.
type SettlementRecord struct {
	ID               uint
	RecordID         string
	OrderID          uint64
	MerchantID       uint64
	TotalAmount      uint64
	PlatformFee      uint64
	SettlementAmount uint64
	Status           string // PENDING, COMPLETED, FAILED
	CreatedAt        time.Time
	SettledAt        *time.Time
}
