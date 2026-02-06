package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口。
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// DailyForecast 结构体定义了未来某一天的销售预测数据。
type DailyForecast struct {
	Date       time.Time `json:"date"`
	Quantity   int32     `json:"quantity"`
	Confidence float64   `json:"confidence"`
}

// DailyForecastArray 定义了 DailyForecast 结构体切片。
type DailyForecastArray []*DailyForecast

func (a DailyForecastArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *DailyForecastArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// SalesForecast 实体代表一个商品的销售预测聚合根。
type SalesForecast struct {
	ID                uint               `json:"id"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	SKUID             uint64             `json:"sku_id"`
	AverageDailySales int32              `json:"average_daily_sales"`
	TrendRate         float64            `json:"trend_rate"`
	Predictions       DailyForecastArray `json:"predictions"`
}

// InventoryWarning 实体代表一个SKU的库存预警信息。
type InventoryWarning struct {
	ID                          uint      `json:"id"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
	SKUID                       uint64    `json:"sku_id"`
	CurrentStock                int32     `json:"current_stock"`
	WarningThreshold            int32     `json:"warning_threshold"`
	DaysUntilEmpty              int32     `json:"days_until_empty"`
	EstimatedEmptyDate          time.Time `json:"estimated_empty_date"`
	NeedReplenishment           bool      `json:"need_replenishment"`
	RecommendedReplenishmentQty int32     `json:"recommended_replenishment_qty"`
}

// SlowMovingItem 实体代表一个滞销品的信息。
type SlowMovingItem struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	SKUID           uint64    `json:"sku_id"`
	ProductName     string    `json:"product_name"`
	CurrentStock    int32     `json:"current_stock"`
	DailySalesRate  float64   `json:"daily_sales_rate"`
	DaysInStock     int32     `json:"days_in_stock"`
	TurnoverRate    float64   `json:"turnover_rate"`
	RecommendAction string    `json:"recommend_action"`
}

// ReplenishmentSuggestion 实体代表一个商品的补货建议。
type ReplenishmentSuggestion struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	SKUID         uint64    `json:"sku_id"`
	ProductName   string    `json:"product_name"`
	CurrentStock  int32     `json:"current_stock"`
	SuggestedQty  int32     `json:"suggested_qty"`
	Priority      string    `json:"priority"`
	Reason        string    `json:"reason"`
	EstimatedCost int64     `json:"estimated_cost"`
}

// StockoutRiskLevel 定义了缺货风险的等级。
type StockoutRiskLevel string

const (
	StockoutRiskLevelLow      StockoutRiskLevel = "low"
	StockoutRiskLevelMedium   StockoutRiskLevel = "medium"
	StockoutRiskLevelHigh     StockoutRiskLevel = "high"
	StockoutRiskLevelCritical StockoutRiskLevel = "critical"
)

// StockoutRisk 实体代表一个商品的缺货风险信息。
type StockoutRisk struct {
	ID                    uint              `json:"id"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	SKUID                 uint64            `json:"sku_id"`
	CurrentStock          int32             `json:"current_stock"`
	DaysUntilStockout     int32             `json:"days_until_stockout"`
	EstimatedStockoutDate time.Time         `json:"estimated_stockout_date"`
	RiskLevel             StockoutRiskLevel `json:"risk_level"`
}

// AggregatedDailySales 实体代表每日聚合的销量数据（用于预测）。
type AggregatedDailySales struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SKUID     uint64    `json:"sku_id"`
	Date      time.Time `json:"date"`
	Quantity  int32     `json:"quantity"`
}
