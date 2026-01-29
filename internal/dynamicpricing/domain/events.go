package domain

import "time"

// PriceCalculatedEvent 动态价格计算完成事件。
type PriceCalculatedEvent struct {
	SKUID      uint64    `json:"sku_id"`
	BasePrice  int64     `json:"base_price"`
	FinalPrice int64     `json:"final_price"`
	Factors    string    `json:"factors"` // JSON 格式的因子详情
	Timestamp  time.Time `json:"timestamp"`
}

// StrategyCreatedEvent 定价策略创建事件。
type StrategyCreatedEvent struct {
	StrategyID   uint64    `json:"strategy_id"`
	SKUID        uint64    `json:"sku_id"`
	StrategyType string    `json:"strategy_type"`
	Timestamp    time.Time `json:"timestamp"`
}

// StrategyUpdatedEvent 定价策略更新事件。
type StrategyUpdatedEvent struct {
	StrategyID   uint64    `json:"strategy_id"`
	SKUID        uint64    `json:"sku_id"`
	StrategyType string    `json:"strategy_type"`
	Enabled      bool      `json:"enabled"`
	Timestamp    time.Time `json:"timestamp"`
}

// PriceElasticityAnalyzedEvent 价格弹性分析完成事件。
type PriceElasticityAnalyzedEvent struct {
	SKUID      uint64    `json:"sku_id"`
	Elasticity float64   `json:"elasticity"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
}

// CompetitorPriceUpdatedEvent 竞品价格更新事件。
type CompetitorPriceUpdatedEvent struct {
	SKUID          uint64    `json:"sku_id"`
	CompetitorName string    `json:"competitor_name"`
	OldPrice       int64     `json:"old_price"`
	NewPrice       int64     `json:"new_price"`
	Timestamp      time.Time `json:"timestamp"`
}
