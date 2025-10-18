package model

import (
	"time"
)

// 注意：这里的 GORM 标签可能不完全适用于 ClickHouse 驱动，仅为示意
// 实际的 ClickHouse 驱动可能有自己的方式来定义表结构

// PageViewEvent 页面浏览事件模型
// 用于追踪用户在网站上的行为
type PageViewEvent struct {
	EventTime time.Time `gorm:"index" json:"event_time"` // 事件发生时间
	UserID    uint      `gorm:"index" json:"user_id"`    // 用户ID (如果已登录)
	SessionID string    `gorm:"index" json:"session_id"` // 会话ID
	URL       string    `json:"url"`                     // 浏览的页面 URL
	Referer   string    `json:"referer"`                 // 来源页面
	UserAgent string    `json:"user_agent"`              // 浏览器 User-Agent
	ClientIP  string    `json:"client_ip"`               // 客户端 IP 地址
}

// SalesFact 销售事实表模型
// 这是一个典型的 OLAP 宽表，用于快速进行销售分析
type SalesFact struct {
	EventTime       time.Time `gorm:"index" json:"event_time"`

	// 订单信息
	OrderID         uint      `gorm:"index" json:"order_id"`
	OrderItemID     uint      `gorm:"primarykey" json:"order_item_id"` // 以订单项为事实表的最小粒度
	OrderTotal      float64   `json:"order_total"`
	DiscountAmount  float64   `json:"discount_amount"`

	// 商品信息
	ProductID       uint      `gorm:"index" json:"product_id"`
	ProductSKU      string    `json:"product_sku"`
	ProductName     string    `json:"product_name"`
	ProductCategory string    `json:"product_category"`
	ProductBrand    string    `json:"product_brand"`
	ItemPrice       float64   `json:"item_price"`
	ItemQuantity    int       `json:"item_quantity"`

	// 用户信息
	UserID          uint      `gorm:"index" json:"user_id"`
	UserCity        string    `json:"user_city"`
	UserCountry     string    `gorm:"index" json:"user_country"`
}

// TableName 自定义表名
func (PageViewEvent) TableName() string {
	return "page_view_events"
}

func (SalesFact) TableName() string {
	return "sales_facts"
}
