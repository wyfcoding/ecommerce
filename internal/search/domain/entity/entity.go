package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// SearchLog 实体是搜索模块的聚合根。
// 它记录了用户的每次搜索行为，包括关键词、结果数量和耗时。
type SearchLog struct {
	gorm.Model         // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	UserID      uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`                    // 搜索用户ID，索引字段。
	Keyword     string `gorm:"type:varchar(255);index;not null;comment:搜索关键词" json:"keyword"` // 搜索关键词，索引字段。
	ResultCount int    `gorm:"not null;comment:结果数量" json:"result_count"`                     // 搜索结果数量。
	Duration    int64  `gorm:"comment:搜索耗时(ms)" json:"duration"`                              // 搜索操作的耗时（毫秒）。
}

// HotKeyword 值对象代表一个热门搜索词。
// 它通常是从SearchLog聚合数据后得到的一个投影（Projection）。
type HotKeyword struct {
	Keyword     string `json:"keyword"`      // 热门关键词。
	SearchCount int    `json:"search_count"` // 搜索次数。
}

// SearchHistory 实体记录了用户的搜索历史。
type SearchHistory struct {
	gorm.Model           // 嵌入gorm.Model。
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`              // 搜索用户ID，索引字段。
	Keyword    string    `gorm:"type:varchar(255);not null;comment:搜索关键词" json:"keyword"` // 搜索关键词。
	Timestamp  time.Time `gorm:"not null;comment:搜索时间" json:"timestamp"`                  // 搜索发生的时间。
}

// SearchFilter 值对象定义了搜索操作的过滤条件。
type SearchFilter struct {
	Keyword    string   `json:"keyword"`     // 搜索关键词。
	CategoryID uint64   `json:"category_id"` // 分类ID。
	BrandID    uint64   `json:"brand_id"`    // 品牌ID。
	PriceMin   float64  `json:"price_min"`   // 价格下限。
	PriceMax   float64  `json:"price_max"`   // 价格上限。
	Sort       string   `json:"sort"`        // 排序方式，例如 "price_asc", "sales_desc"。
	Page       int      `json:"page"`        // 页码。
	PageSize   int      `json:"page_size"`   // 每页数量。
	Tags       []string `json:"tags"`        // 标签过滤。
}

// SearchResult 值对象代表一次搜索操作的结果。
type SearchResult struct {
	Total int64         `json:"total"` // 搜索到的总记录数。
	Items []interface{} `json:"items"` // 搜索到的商品或其他实体列表。使用interface{}表示结果类型可以多样化。
}

// Suggestion 值对象代表一个搜索建议。
type Suggestion struct {
	Keyword string `json:"keyword"` // 建议关键词。
	Score   int    `json:"score"`   // 建议的得分或优先级。
	Type    string `json:"type"`    // 建议的类型，例如 "history", "hot", "completion"。
}

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value interface{}) error {
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
