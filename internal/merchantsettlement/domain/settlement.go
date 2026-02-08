package domain

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SettlementStatus string

const (
	StatusUnpaid SettlementStatus = "UNPAID"
	StatusPaid   SettlementStatus = "PAID"
	StatusFailed SettlementStatus = "FAILED"
)

// Settlement 商家结算单实体
type Settlement struct {
	gorm.Model
	SettlementID   string           `gorm:"column:settlement_id;type:varchar(32);unique_index;not null"`
	MerchantID     string           `gorm:"column:merchant_id;type:varchar(32);index;not null"`
	Amount         decimal.Decimal  `gorm:"column:amount;type:decimal(32,16);not null"`
	Status         SettlementStatus `gorm:"column:status;type:varchar(20);not null;default:'UNPAID'"`
	PeriodStart    time.Time        `gorm:"column:period_start"`
	PeriodEnd      time.Time        `gorm:"column:period_end"`
	TransactionRef string           `gorm:"column:transaction_ref;type:varchar(64)"`
}

func (Settlement) TableName() string { return "merchant_settlements" }
