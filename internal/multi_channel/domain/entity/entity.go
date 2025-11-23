package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Channel 渠道
type Channel struct {
	gorm.Model
	Name      string `gorm:"type:varchar(255);uniqueIndex;not null;comment:渠道名称" json:"name"`
	Type      string `gorm:"type:varchar(32);not null;comment:类型" json:"type"` // marketplace, social, direct
	APIKey    string `gorm:"type:varchar(255);comment:API Key" json:"api_key"`
	APISecret string `gorm:"type:varchar(255);comment:API Secret" json:"api_secret"`
	IsEnabled bool   `gorm:"default:true;comment:是否启用" json:"is_enabled"`
}

// OrderItem 订单项 (Value Object)
type OrderItem struct {
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
	Price       int64  `json:"price"`
	SKU         string `json:"sku"`
}

// OrderItemArray defines a slice of OrderItem that implements sql.Scanner and driver.Valuer
type OrderItemArray []*OrderItem

func (a OrderItemArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *OrderItemArray) Scan(value interface{}) error {
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

// BuyerInfo 买家信息 (Value Object)
type BuyerInfo struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
}

func (b BuyerInfo) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *BuyerInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

// ShippingInfo 配送信息 (Value Object)
type ShippingInfo struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

func (s ShippingInfo) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ShippingInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// LocalOrder 本地订单 (Synced from Channel)
type LocalOrder struct {
	gorm.Model
	ChannelID      uint64         `gorm:"not null;index;comment:渠道ID" json:"channel_id"`
	ChannelName    string         `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"`
	ChannelOrderID string         `gorm:"type:varchar(255);index;not null;comment:渠道订单ID" json:"channel_order_id"`
	Items          OrderItemArray `gorm:"type:json;comment:订单项" json:"items"`
	TotalAmount    int64          `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`
	BuyerInfo      BuyerInfo      `gorm:"type:json;comment:买家信息" json:"buyer_info"`
	ShippingInfo   ShippingInfo   `gorm:"type:json;comment:配送信息" json:"shipping_info"`
	Status         string         `gorm:"type:varchar(32);not null;comment:状态" json:"status"` // pending, processing, shipped, delivered, cancelled
}

// ChannelSyncLog 渠道同步日志
type ChannelSyncLog struct {
	gorm.Model
	ChannelID    uint64 `gorm:"not null;index;comment:渠道ID" json:"channel_id"`
	ChannelName  string `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"`
	Type         string `gorm:"type:varchar(32);not null;comment:同步类型" json:"type"` // order, inventory
	Status       string `gorm:"type:varchar(32);not null;comment:状态" json:"status"` // success, failure
	Message      string `gorm:"type:text;comment:消息" json:"message"`
	ItemsCount   int32  `gorm:"default:0;comment:同步条目数" json:"items_count"`
	SuccessCount int32  `gorm:"default:0;comment:成功数" json:"success_count"`
	FailureCount int32  `gorm:"default:0;comment:失败数" json:"failure_count"`
	StartTime    time.Time
	EndTime      time.Time
}
