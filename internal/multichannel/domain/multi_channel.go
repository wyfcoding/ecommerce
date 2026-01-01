package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Channel 实体代表一个销售渠道。
type Channel struct {
	gorm.Model
	Name      string `gorm:"type:varchar(255);uniqueIndex;not null;comment:渠道名称" json:"name"`
	Type      string `gorm:"type:varchar(32);not null;comment:类型" json:"type"`
	APIKey    string `gorm:"type:varchar(255);comment:API Key" json:"api_key"`
	APISecret string `gorm:"type:varchar(255);comment:API Secret" json:"api_secret"`
	IsEnabled bool   `gorm:"default:true;comment:是否启用" json:"is_enabled"`
}

// OrderItem 值对象定义了订单中的一个商品项。
type OrderItem struct {
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
	Price       int64  `json:"price"`
	SKU         string `json:"sku"`
}

// OrderItemArray 定义了一个 OrderItem 结构体切片。
type OrderItemArray []*OrderItem

func (a OrderItemArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *OrderItemArray) Scan(value any) error {
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

// BuyerInfo 值对象定义了订单买家的信息。
type BuyerInfo struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
}

func (b BuyerInfo) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *BuyerInfo) Scan(value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

// ShippingInfo 值对象定义了订单的配送信息。
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

func (s *ShippingInfo) Scan(value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// LocalOrder 实体代表一个从外部渠道同步到本地系统的订单。
type LocalOrder struct {
	gorm.Model
	ChannelID      uint64         `gorm:"not null;index;comment:渠道ID" json:"channel_id"`
	ChannelName    string         `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"`
	ChannelOrderID string         `gorm:"type:varchar(255);index;not null;comment:渠道订单ID" json:"channel_order_id"`
	Items          OrderItemArray `gorm:"type:json;comment:订单项" json:"items"`
	TotalAmount    int64          `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`
	BuyerInfo      BuyerInfo      `gorm:"type:json;comment:买家信息" json:"buyer_info"`
	ShippingInfo   ShippingInfo   `gorm:"type:json;comment:配送信息" json:"shipping_info"`
	Status         string         `gorm:"type:varchar(32);not null;comment:状态" json:"status"`
}

// ChannelSyncLog 实体代表一次渠道数据同步的日志记录。
type ChannelSyncLog struct {
	gorm.Model
	ChannelID    uint64    `gorm:"not null;index;comment:渠道ID" json:"channel_id"`
	ChannelName  string    `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"`
	Type         string    `gorm:"type:varchar(32);not null;comment:同步类型" json:"type"`
	Status       string    `gorm:"type:varchar(32);not null;comment:状态" json:"status"`
	Message      string    `gorm:"type:text;comment:消息" json:"message"`
	ItemsCount   int32     `gorm:"default:0;comment:同步条目数" json:"items_count"`
	SuccessCount int32     `gorm:"default:0;comment:成功数" json:"success_count"`
	FailureCount int32     `gorm:"default:0;comment:失败数" json:"failure_count"`
	StartTime    time.Time `gorm:"comment:开始时间" json:"start_time"`
	EndTime      time.Time `gorm:"comment:结束时间" json:"end_time"`
}

// ChannelAdapter 外部渠道适配器接口
type ChannelAdapter interface {
	// FetchOrders 拉取指定时间范围内的订单数据
	FetchOrders(ctx context.Context, channel *Channel, startTime, endTime time.Time) ([]*LocalOrder, error)
}
