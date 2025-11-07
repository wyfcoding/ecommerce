package model

import (
	"time"
)

// PageViewEvent 页面浏览事件模型。
// 用于追踪用户在网站上的行为，是用户行为分析的基础数据。
type PageViewEvent struct {
	EventTime time.Time `gorm:"index" json:"event_time"` // 事件发生时间
	UserID    uint      `gorm:"index" json:"user_id"`    // 用户ID (如果已登录)，用于关联用户行为
	SessionID string    `gorm:"index" json:"session_id"` // 会话ID，用于追踪用户在一次会话中的行为
	URL       string    `json:"url"`                     // 浏览的页面 URL
	Referer   string    `json:"referer"`                 // 来源页面 URL
	UserAgent string    `json:"user_agent"`              // 浏览器 User-Agent 字符串
	ClientIP  string    `json:"client_ip"`               // 客户端 IP 地址
}

// SalesFact 销售事实表模型。
// 这是一个典型的 OLAP 宽表，用于快速进行销售分析和报表生成。
type SalesFact struct {
	EventTime       time.Time `gorm:"index" json:"event_time"`       // 销售事件发生时间

	// 订单信息
	OrderID         uint      `gorm:"index" json:"order_id"`         // 订单ID
	OrderItemID     uint      `gorm:"primarykey" json:"order_item_id"` // 订单项ID，作为事实表的最小粒度
	OrderSN         string    `gorm:"index" json:"order_sn"`         // 订单号
	OrderTotal      float64   `json:"order_total"`                   // 订单总金额
	DiscountAmount  float64   `json:"discount_amount"`               // 订单优惠金额

	// 商品信息
	ProductID       uint      `gorm:"index" json:"product_id"`       // 商品ID
	ProductSKU      string    `json:"product_sku"`                   // 商品SKU
	ProductName     string    `json:"product_name"`                  // 商品名称
	ProductCategory string    `gorm:"index" json:"product_category"` // 商品分类
	ProductBrand    string    `gorm:"index" json:"product_brand"`    // 商品品牌
	ItemPrice       float64   `json:"item_price"`                    // 单个商品项价格
	ItemQuantity    int       `json:"item_quantity"`                 // 单个商品项数量

	// 用户信息
	UserID          uint      `gorm:"index" json:"user_id"`          // 用户ID
	UserCity        string    `json:"user_city"`                     // 用户所在城市
	UserCountry     string    `gorm:"index" json:"user_country"`     // 用户所在国家
}

// TableName 自定义 PageViewEvent 对应的表名。
func (PageViewEvent) TableName() string {
	return "page_view_events"
}

// TableName 自定义 SalesFact 对应的表名。
func (SalesFact) TableName() string {
	return "sales_facts"
}