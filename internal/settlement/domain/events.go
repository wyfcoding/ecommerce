package domain

import (
	"time"
)

// SettlementCreatedEvent 结算单创建事件
type SettlementCreatedEvent struct {
	SettlementNo string    `json:"settlement_no"`
	MerchantID   uint64    `json:"merchant_id"`
	TotalAmount  uint64    `json:"total_amount"`
	Timestamp    time.Time `json:"timestamp"`
}

// SettlementProcessedEvent 结算单处理中事件
type SettlementProcessedEvent struct {
	SettlementNo string    `json:"settlement_no"`
	MerchantID   uint64    `json:"merchant_id"`
	Amount       uint64    `json:"amount"`
	Timestamp    time.Time `json:"timestamp"`
}

// SettlementCompletedEvent 结算单完成事件
type SettlementCompletedEvent struct {
	SettlementNo string    `json:"settlement_no"`
	MerchantID   uint64    `json:"merchant_id"`
	Amount       uint64    `json:"amount"`
	Timestamp    time.Time `json:"timestamp"`
}

// SettlementFailedEvent 结算单失败事件
type SettlementFailedEvent struct {
	SettlementNo string    `json:"settlement_no"`
	MerchantID   uint64    `json:"merchant_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}
