package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// Channel 实体代表一个销售渠道。
// 它包含了渠道的名称、类型、API凭证和启用状态等信息。
type Channel struct {
	gorm.Model        // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name       string `gorm:"type:varchar(255);uniqueIndex;not null;comment:渠道名称" json:"name"` // 渠道名称，唯一索引，不允许为空。
	Type       string `gorm:"type:varchar(32);not null;comment:类型" json:"type"`                // 渠道类型，例如“marketplace”（市场），“social”（社交），“direct”（官网）。
	APIKey     string `gorm:"type:varchar(255);comment:API Key" json:"api_key"`                // 访问渠道API的Key。
	APISecret  string `gorm:"type:varchar(255);comment:API Secret" json:"api_secret"`          // 访问渠道API的Secret。
	IsEnabled  bool   `gorm:"default:true;comment:是否启用" json:"is_enabled"`                     // 渠道是否启用。
}

// OrderItem 值对象定义了订单中的一个商品项。
// 它是多渠道订单同步中的一个子组件。
type OrderItem struct {
	ProductID   uint64 `json:"product_id"`   // 商品ID。
	ProductName string `json:"product_name"` // 商品名称。
	Quantity    int32  `json:"quantity"`     // 购买数量。
	Price       int64  `json:"price"`        // 商品价格（单位：分）。
	SKU         string `json:"sku"`          // 商品SKU。
}

// OrderItemArray 定义了一个 OrderItem 结构体切片，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将 OrderItem 切片作为JSON字符串存储到数据库，并从数据库读取。
type OrderItemArray []*OrderItem

// Value 实现 driver.Valuer 接口，将 OrderItemArray 转换为数据库可以存储的值（JSON字节数组）。
func (a OrderItemArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 OrderItemArray。
func (a *OrderItemArray) Scan(value interface{}) error {
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

// BuyerInfo 值对象定义了订单买家的信息。
type BuyerInfo struct {
	Name    string `json:"name"`    // 买家姓名。
	Email   string `json:"email"`   // 买家邮箱。
	Phone   string `json:"phone"`   // 买家电话。
	Country string `json:"country"` // 买家所在国家。
}

// Value 实现 driver.Valuer 接口，将 BuyerInfo 转换为数据库可以存储的值（JSON字节数组）。
func (b BuyerInfo) Value() (driver.Value, error) {
	return json.Marshal(b) // 将结构体编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 BuyerInfo。
func (b *BuyerInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b) // 将JSON字节数组解码为结构体。
}

// ShippingInfo 值对象定义了订单的配送信息。
type ShippingInfo struct {
	Address string `json:"address"`  // 收货地址。
	City    string `json:"city"`     // 城市。
	State   string `json:"state"`    // 省份/州。
	ZipCode string `json:"zip_code"` // 邮政编码。
	Country string `json:"country"`  // 国家。
}

// Value 实现 driver.Valuer 接口，将 ShippingInfo 转换为数据库可以存储的值（JSON字节数组）。
func (s ShippingInfo) Value() (driver.Value, error) {
	return json.Marshal(s) // 将结构体编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 ShippingInfo。
func (s *ShippingInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s) // 将JSON字节数组解码为结构体。
}

// LocalOrder 实体代表一个从外部渠道同步到本地系统的订单。
// 它包含了渠道订单的基本信息、商品项、买家信息和配送信息。
type LocalOrder struct {
	gorm.Model                    // 嵌入gorm.Model。
	ChannelID      uint64         `gorm:"not null;index;comment:渠道ID" json:"channel_id"`                           // 订单来源渠道的ID，索引字段。
	ChannelName    string         `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"`             // 订单来源渠道的名称。
	ChannelOrderID string         `gorm:"type:varchar(255);index;not null;comment:渠道订单ID" json:"channel_order_id"` // 外部渠道的订单ID，索引字段。
	Items          OrderItemArray `gorm:"type:json;comment:订单项" json:"items"`                                      // 订单包含的商品项列表，存储为JSON。
	TotalAmount    int64          `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`                   // 订单总金额（单位：分）。
	BuyerInfo      BuyerInfo      `gorm:"type:json;comment:买家信息" json:"buyer_info"`                                // 买家信息，存储为JSON。
	ShippingInfo   ShippingInfo   `gorm:"type:json;comment:配送信息" json:"shipping_info"`                             // 配送信息，存储为JSON。
	Status         string         `gorm:"type:varchar(32);not null;comment:状态" json:"status"`                      // 订单状态，例如“pending”，“processing”，“shipped”，“delivered”，“cancelled”。
}

// ChannelSyncLog 实体代表一次渠道数据同步的日志记录。
type ChannelSyncLog struct {
	gorm.Model             // 嵌入gorm.Model。
	ChannelID    uint64    `gorm:"not null;index;comment:渠道ID" json:"channel_id"`               // 关联的渠道ID，索引字段。
	ChannelName  string    `gorm:"type:varchar(255);not null;comment:渠道名称" json:"channel_name"` // 渠道名称。
	Type         string    `gorm:"type:varchar(32);not null;comment:同步类型" json:"type"`          // 同步的数据类型，例如“order”（订单），“inventory”（库存）。
	Status       string    `gorm:"type:varchar(32);not null;comment:状态" json:"status"`          // 同步操作的状态，例如“success”，“failure”。
	Message      string    `gorm:"type:text;comment:消息" json:"message"`                         // 同步操作的日志消息。
	ItemsCount   int32     `gorm:"default:0;comment:同步条目数" json:"items_count"`                  // 本次同步操作涉及的条目总数。
	SuccessCount int32     `gorm:"default:0;comment:成功数" json:"success_count"`                  // 成功同步的条目数量。
	FailureCount int32     `gorm:"default:0;comment:失败数" json:"failure_count"`                  // 失败的条目数量。
	StartTime    time.Time `gorm:"comment:开始时间" json:"start_time"`                              // 同步操作开始时间。
	EndTime      time.Time `gorm:"comment:结束时间" json:"end_time"`                                // 同步操作结束时间。
}
