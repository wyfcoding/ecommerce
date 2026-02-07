package domain

import "time"

const (
	FlashsaleCreatedEventType        = "flashsale.created"
	FlashsaleOrderCreatedEventType   = "flashsale.order.created"
	FlashsaleOrderPaidEventType      = "flashsale.order.paid"
	FlashsaleOrderCancelledEventType = "flashsale.order.cancelled"
	FlashsaleStockExhaustedEventType = "flashsale.stock.exhausted"
)

// FlashSaleEventCreatedEvent 秒杀活动创建事件
type FlashSaleEventCreatedEvent struct {
	EventID   uint      `json:"event_id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timestamp time.Time `json:"timestamp"`
}

// FlashSaleOrderCreatedEvent 秒杀订单创建事件
type FlashSaleOrderCreatedEvent struct {
	OrderID     uint64    `json:"order_id"`
	FlashsaleID uint64    `json:"flashsale_id"`
	UserID      uint64    `json:"user_id"`
	ProductID   uint64    `json:"product_id"`
	SkuID       uint64    `json:"sku_id"`
	Quantity    int32     `json:"quantity"`
	Price       int64     `json:"price"`
	Timestamp   time.Time `json:"timestamp"`
}

// FlashSaleOrderPaidEvent 秒杀订单支付事件
type FlashSaleOrderPaidEvent struct {
	OrderID   uint64    `json:"order_id"`
	Timestamp time.Time `json:"timestamp"`
}

// FlashSaleOrderCancelledEvent 秒杀订单取消事件
type FlashSaleOrderCancelledEvent struct {
	OrderID     uint64    `json:"order_id"`
	FlashsaleID uint64    `json:"flashsale_id"`
	UserID      uint64    `json:"user_id"`
	Quantity    int32     `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
}

// FlashSaleStockExhaustedEvent 秒杀库存售罄事件
type FlashSaleStockExhaustedEvent struct {
	FlashsaleID uint64    `json:"flashsale_id"`
	Timestamp   time.Time `json:"timestamp"`
}
