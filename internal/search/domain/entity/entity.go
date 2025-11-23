package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SearchLog 搜索日志聚合根
type SearchLog struct {
	gorm.Model
	UserID      uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Keyword     string `gorm:"type:varchar(255);index;not null;comment:搜索关键词" json:"keyword"`
	ResultCount int    `gorm:"not null;comment:结果数量" json:"result_count"`
	Duration    int64  `gorm:"comment:搜索耗时(ms)" json:"duration"`
}

// HotKeyword 热门搜索词 (Value Object / Projection)
type HotKeyword struct {
	Keyword     string `json:"keyword"`
	SearchCount int    `json:"search_count"`
}

// SearchHistory 搜索历史
type SearchHistory struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Keyword   string    `gorm:"type:varchar(255);not null;comment:搜索关键词" json:"keyword"`
	Timestamp time.Time `gorm:"not null;comment:搜索时间" json:"timestamp"`
}

// SearchFilter 搜索过滤条件 (Value Object)
type SearchFilter struct {
	Keyword    string   `json:"keyword"`
	CategoryID uint64   `json:"category_id"`
	BrandID    uint64   `json:"brand_id"`
	PriceMin   float64  `json:"price_min"`
	PriceMax   float64  `json:"price_max"`
	Sort       string   `json:"sort"` // e.g., "price_asc", "sales_desc"
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	Tags       []string `json:"tags"`
}

// SearchResult 搜索结果 (Value Object)
type SearchResult struct {
	Total int64         `json:"total"`
	Items []interface{} `json:"items"` // Can be Product or other entities
}

// Suggestion 搜索建议 (Value Object)
type Suggestion struct {
	Keyword string `json:"keyword"`
	Score   int    `json:"score"`
	Type    string `json:"type"` // e.g., "history", "hot", "completion"
}

// StringArray defines a slice of strings that implements sql.Scanner and driver.Valuer
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
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
