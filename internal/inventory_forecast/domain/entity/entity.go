package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONMap defines a map that implements the sql.Scanner and driver.Valuer interfaces
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
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

// DailyForecastArray defines a slice of DailyForecast that implements sql.Scanner and driver.Valuer
type DailyForecastArray []*DailyForecast

func (a DailyForecastArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *DailyForecastArray) Scan(value interface{}) error {
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

// DailyForecast 日预测
type DailyForecast struct {
	Date       time.Time `json:"date"`
	Quantity   int32     `json:"quantity"`
	Confidence float64   `json:"confidence"`
}

// SalesForecast 销售预测
type SalesForecast struct {
	gorm.Model
	SKUID             uint64             `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	AverageDailySales int32              `gorm:"not null;default:0;comment:日均销量" json:"average_daily_sales"`
	TrendRate         float64            `gorm:"type:decimal(5,2);default:0.00;comment:趋势率" json:"trend_rate"`
	Predictions       DailyForecastArray `gorm:"type:json;comment:预测详情" json:"predictions"`
}

// InventoryWarning 库存预警
type InventoryWarning struct {
	gorm.Model
	SKUID                       uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	CurrentStock                int32     `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`
	WarningThreshold            int32     `gorm:"not null;default:0;comment:预警阈值" json:"warning_threshold"`
	DaysUntilEmpty              int32     `gorm:"not null;default:0;comment:预计售罄天数" json:"days_until_empty"`
	EstimatedEmptyDate          time.Time `gorm:"comment:预计售罄日期" json:"estimated_empty_date"`
	NeedReplenishment           bool      `gorm:"default:false;comment:是否需要补货" json:"need_replenishment"`
	RecommendedReplenishmentQty int32     `gorm:"not null;default:0;comment:建议补货数量" json:"recommended_replenishment_qty"`
}

// SlowMovingItem 滞销品
type SlowMovingItem struct {
	gorm.Model
	SKUID           uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	ProductName     string  `gorm:"type:varchar(255);comment:商品名称" json:"product_name"`
	CurrentStock    int32   `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`
	DailySalesRate  float64 `gorm:"type:decimal(10,4);default:0.0000;comment:日均动销率" json:"daily_sales_rate"`
	DaysInStock     int32   `gorm:"not null;default:0;comment:库龄(天)" json:"days_in_stock"`
	TurnoverRate    float64 `gorm:"type:decimal(5,2);default:0.00;comment:周转率" json:"turnover_rate"`
	RecommendAction string  `gorm:"type:varchar(255);comment:建议措施" json:"recommend_action"`
}

// ReplenishmentSuggestion 补货建议
type ReplenishmentSuggestion struct {
	gorm.Model
	SKUID         uint64 `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	ProductName   string `gorm:"type:varchar(255);comment:商品名称" json:"product_name"`
	CurrentStock  int32  `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`
	SuggestedQty  int32  `gorm:"not null;default:0;comment:建议补货数量" json:"suggested_qty"`
	Priority      string `gorm:"type:varchar(32);comment:优先级" json:"priority"` // high, medium, low
	Reason        string `gorm:"type:varchar(255);comment:原因" json:"reason"`
	EstimatedCost int64  `gorm:"not null;default:0;comment:预计成本(分)" json:"estimated_cost"`
}

// StockoutRiskLevel 缺货风险等级
type StockoutRiskLevel string

const (
	StockoutRiskLevelLow      StockoutRiskLevel = "low"
	StockoutRiskLevelMedium   StockoutRiskLevel = "medium"
	StockoutRiskLevelHigh     StockoutRiskLevel = "high"
	StockoutRiskLevelCritical StockoutRiskLevel = "critical"
)

// StockoutRisk 缺货风险
type StockoutRisk struct {
	gorm.Model
	SKUID                 uint64            `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	CurrentStock          int32             `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`
	DaysUntilStockout     int32             `gorm:"not null;default:0;comment:预计缺货天数" json:"days_until_stockout"`
	EstimatedStockoutDate time.Time         `gorm:"comment:预计缺货日期" json:"estimated_stockout_date"`
	RiskLevel             StockoutRiskLevel `gorm:"type:varchar(32);comment:风险等级" json:"risk_level"`
}
