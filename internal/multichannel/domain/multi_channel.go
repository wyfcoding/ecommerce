package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Channel 实体代表一个销售渠道。
type Channel struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	APIKey    string    `json:"api_key"`
	APISecret string    `json:"api_secret"`
	IsEnabled bool      `json:"is_enabled"`
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
	ID             uint           `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ChannelID      uint64         `json:"channel_id"`
	ChannelName    string         `json:"channel_name"`
	ChannelOrderID string         `json:"channel_order_id"`
	Items          OrderItemArray `json:"items"`
	TotalAmount    int64          `json:"total_amount"`
	BuyerInfo      BuyerInfo      `json:"buyer_info"`
	ShippingInfo   ShippingInfo   `json:"shipping_info"`
	Status         string         `json:"status"`
}

// ChannelSyncLog 实体代表一次渠道数据同步的日志记录。
type ChannelSyncLog struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ChannelID    uint64    `json:"channel_id"`
	ChannelName  string    `json:"channel_name"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	ItemsCount   int32     `json:"items_count"`
	SuccessCount int32     `json:"success_count"`
	FailureCount int32     `json:"failure_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// ChannelAdapter 外部渠道适配器接口
type ChannelAdapter interface {
	// FetchOrders 拉取指定时间范围内的订单数据
	FetchOrders(ctx context.Context, channel *Channel, startTime, endTime time.Time) ([]*LocalOrder, error)
}
