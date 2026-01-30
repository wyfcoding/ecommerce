package domain

import "time"

// PointItemExchangedEvent 积分商品兑换事件。
type PointItemExchangedEvent struct {
	ExchangeID uint64    `json:"exchange_id"`
	UserID     uint64    `json:"user_id"`
	ItemID     uint64    `json:"item_id"`
	Points     int32     `json:"points"`
	Timestamp  time.Time `json:"timestamp"`
}

// PointsStockUpdatedEvent 积分商品库存更新事件。
type PointsStockUpdatedEvent struct {
	ItemID    uint64    `json:"item_id"`
	NewStock  int32     `json:"new_stock"`
	Timestamp time.Time `json:"timestamp"`
}
