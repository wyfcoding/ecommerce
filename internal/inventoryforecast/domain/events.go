package domain

import "time"

// ForecastGeneratedEvent 预测生成事件。
type ForecastGeneratedEvent struct {
	ForecastID uint64    `json:"forecast_id"`
	ProductID  uint64    `json:"product_id"`
	SkuID      uint64    `json:"sku_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// ThresholdAlertEvent 库存阈值预警事件。
type ThresholdAlertEvent struct {
	ProductID     uint64    `json:"product_id"`
	SkuID         uint64    `json:"sku_id"`
	CurrentStock  int32     `json:"current_stock"`
	ForecastStock int32     `json:"forecast_stock"`
	Timestamp     time.Time `json:"timestamp"`
}
