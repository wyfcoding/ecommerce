package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的map[string]interface{}类型作为JSON字符串存储到数据库，并从数据库读取。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口，将 JSONMap 转换为数据库可以存储的值（JSON字节数组）。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m) // 将map编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 JSONMap。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m) // 将JSON字节数组解码为map。
}

// DailyForecastArray 定义了一个 DailyForecast 结构体切片，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将 DailyForecast 切片作为JSON字符串存储到数据库，并从数据库读取。
type DailyForecastArray []*DailyForecast

// Value 实现 driver.Valuer 接口，将 DailyForecastArray 转换为数据库可以存储的值（JSON字节数组）。
func (a DailyForecastArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 DailyForecastArray。
func (a *DailyForecastArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a) // 将JSON字节数组解码为切片。
}

// DailyForecast 结构体定义了未来某一天的销售预测数据。
type DailyForecast struct {
	Date       time.Time `json:"date"`       // 预测日期。
	Quantity   int32     `json:"quantity"`   // 预测销量。
	Confidence float64   `json:"confidence"` // 预测的置信度。
}

// SalesForecast 实体代表一个商品的销售预测聚合根。
// 它包含了SKU的整体预测信息、趋势和每日具体的预测数据。
type SalesForecast struct {
	gorm.Model                           // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SKUID             uint64             `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                  // 关联的SKU ID，索引字段。
	AverageDailySales int32              `gorm:"not null;default:0;comment:日均销量" json:"average_daily_sales"`   // 历史日均销量。
	TrendRate         float64            `gorm:"type:decimal(5,2);default:0.00;comment:趋势率" json:"trend_rate"` // 销量增长或下降的趋势率。
	Predictions       DailyForecastArray `gorm:"type:json;comment:预测详情" json:"predictions"`                    // 每日详细预测数据列表，存储为JSON。
}

// InventoryWarning 实体代表一个SKU的库存预警信息。
// 当库存量低于某个阈值时，系统会生成此预警。
type InventoryWarning struct {
	gorm.Model                            // 嵌入gorm.Model。
	SKUID                       uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                            // 关联的SKU ID，索引字段。
	CurrentStock                int32     `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`                   // 当前库存量。
	WarningThreshold            int32     `gorm:"not null;default:0;comment:预警阈值" json:"warning_threshold"`               // 触发预警的库存阈值。
	DaysUntilEmpty              int32     `gorm:"not null;default:0;comment:预计售罄天数" json:"days_until_empty"`              // 预计库存售罄的天数。
	EstimatedEmptyDate          time.Time `gorm:"comment:预计售罄日期" json:"estimated_empty_date"`                             // 预计库存售罄的日期。
	NeedReplenishment           bool      `gorm:"default:false;comment:是否需要补货" json:"need_replenishment"`                 // 是否需要触发补货流程。
	RecommendedReplenishmentQty int32     `gorm:"not null;default:0;comment:建议补货数量" json:"recommended_replenishment_qty"` // 系统建议的补货数量。
}

// SlowMovingItem 实体代表一个滞销品的信息。
// 滞销品是指销量较低、库存周转慢的商品。
type SlowMovingItem struct {
	gorm.Model              // 嵌入gorm.Model。
	SKUID           uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                             // 关联的SKU ID，索引字段。
	ProductName     string  `gorm:"type:varchar(255);comment:商品名称" json:"product_name"`                      // 商品名称。
	CurrentStock    int32   `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`                    // 当前库存量。
	DailySalesRate  float64 `gorm:"type:decimal(10,4);default:0.0000;comment:日均动销率" json:"daily_sales_rate"` // 日均动销率。
	DaysInStock     int32   `gorm:"not null;default:0;comment:库龄(天)" json:"days_in_stock"`                   // 商品在库天数（库龄）。
	TurnoverRate    float64 `gorm:"type:decimal(5,2);default:0.00;comment:周转率" json:"turnover_rate"`         // 库存周转率。
	RecommendAction string  `gorm:"type:varchar(255);comment:建议措施" json:"recommend_action"`                  // 针对滞销品的建议处理措施，例如“促销”、“清仓”。
}

// ReplenishmentSuggestion 实体代表一个商品的补货建议。
type ReplenishmentSuggestion struct {
	gorm.Model           // 嵌入gorm.Model。
	SKUID         uint64 `gorm:"not null;index;comment:SKU ID" json:"sku_id"`              // 关联的SKU ID，索引字段。
	ProductName   string `gorm:"type:varchar(255);comment:商品名称" json:"product_name"`       // 商品名称。
	CurrentStock  int32  `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`     // 当前库存量。
	SuggestedQty  int32  `gorm:"not null;default:0;comment:建议补货数量" json:"suggested_qty"`   // 系统建议的补货数量。
	Priority      string `gorm:"type:varchar(32);comment:优先级" json:"priority"`             // 补货优先级，例如“high”，“medium”，“low”。
	Reason        string `gorm:"type:varchar(255);comment:原因" json:"reason"`               // 补货建议的原因。
	EstimatedCost int64  `gorm:"not null;default:0;comment:预计成本(分)" json:"estimated_cost"` // 预计补货成本。
}

// StockoutRiskLevel 定义了缺货风险的等级。
type StockoutRiskLevel string

const (
	StockoutRiskLevelLow      StockoutRiskLevel = "low"      // 低风险。
	StockoutRiskLevelMedium   StockoutRiskLevel = "medium"   // 中风险。
	StockoutRiskLevelHigh     StockoutRiskLevel = "high"     // 高风险。
	StockoutRiskLevelCritical StockoutRiskLevel = "critical" // 关键风险。
)

// StockoutRisk 实体代表一个商品的缺货风险信息。
type StockoutRisk struct {
	gorm.Model                              // 嵌入gorm.Model。
	SKUID                 uint64            `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                  // 关联的SKU ID，索引字段。
	CurrentStock          int32             `gorm:"not null;default:0;comment:当前库存" json:"current_stock"`         // 当前库存量。
	DaysUntilStockout     int32             `gorm:"not null;default:0;comment:预计缺货天数" json:"days_until_stockout"` // 预计库存售罄的天数。
	EstimatedStockoutDate time.Time         `gorm:"comment:预计缺货日期" json:"estimated_stockout_date"`                // 预计缺货的日期。
	RiskLevel             StockoutRiskLevel `gorm:"type:varchar(32);comment:风险等级" json:"risk_level"`              // 缺货风险等级。
}
