package domain

import "time"

// PriceCalculatedEvent 价格计算完成事件。
type PriceCalculatedEvent struct {
	ProductID uint64    `json:"product_id"`
	SKUID     uint64    `json:"sku_id"`
	UserID    uint64    `json:"user_id"`
	Price     uint64    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// PricingRuleUpdatedEvent 定价规则更新事件。
type PricingRuleUpdatedEvent struct {
	RuleID    uint64    `json:"rule_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
