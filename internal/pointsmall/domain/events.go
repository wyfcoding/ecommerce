package domain

import "time"

const (
	PointItemExchangedEventType = "pointsmall.item.exchanged"
	PointsStockUpdatedEventType = "pointsmall.stock.updated"
	PointsProductCreatedEventType = "pointsmall.product.created"
	PointsOrderCreatedEventType   = "pointsmall.order.created"
	PointsAccountUpdatedEventType = "pointsmall.account.updated"
)

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

// PointsProductCreatedEvent 积分商品创建事件。
type PointsProductCreatedEvent struct {
	ProductID uint64              `json:"product_id"`
	Status    PointsProductStatus `json:"status"`
	Stock     int32               `json:"stock"`
	Timestamp time.Time           `json:"timestamp"`
}

// PointsOrderCreatedEvent 积分订单创建事件。
type PointsOrderCreatedEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	ProductID   uint64    `json:"product_id"`
	Quantity    int32     `json:"quantity"`
	TotalPoints int64     `json:"total_points"`
	Timestamp   time.Time `json:"timestamp"`
}

// PointsAccountUpdatedEvent 积分账户更新事件。
type PointsAccountUpdatedEvent struct {
	UserID      uint64    `json:"user_id"`
	TotalPoints int64     `json:"total_points"`
	UsedPoints  int64     `json:"used_points"`
	Timestamp   time.Time `json:"timestamp"`
}
