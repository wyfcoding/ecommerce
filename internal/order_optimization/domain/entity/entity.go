package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

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

// OrderItem 订单项 (Value Object)
type OrderItem struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Price     int64  `json:"price"`
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

// ShippingAddress 配送地址 (Value Object)
type ShippingAddress struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

func (s ShippingAddress) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ShippingAddress) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// Uint64Array defines a slice of uint64 that implements sql.Scanner and driver.Valuer
type Uint64Array []uint64

func (a Uint64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *Uint64Array) Scan(value interface{}) error {
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

// MergedOrder 合并订单
type MergedOrder struct {
	gorm.Model
	UserID           uint64          `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OriginalOrderIDs Uint64Array     `gorm:"type:json;comment:原始订单ID列表" json:"original_order_ids"`
	Items            OrderItemArray  `gorm:"type:json;comment:订单项" json:"items"`
	TotalAmount      int64           `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`
	DiscountAmount   int64           `gorm:"not null;default:0;comment:优惠金额(分)" json:"discount_amount"`
	FinalAmount      int64           `gorm:"not null;default:0;comment:最终金额(分)" json:"final_amount"`
	ShippingAddress  ShippingAddress `gorm:"type:json;comment:配送地址" json:"shipping_address"`
	Status           string          `gorm:"type:varchar(32);not null;comment:状态" json:"status"`
}

// SplitOrder 拆分订单
type SplitOrder struct {
	gorm.Model
	OriginalOrderID uint64          `gorm:"not null;index;comment:原始订单ID" json:"original_order_id"`
	SplitIndex      int32           `gorm:"not null;comment:拆分序号" json:"split_index"`
	Items           OrderItemArray  `gorm:"type:json;comment:订单项" json:"items"`
	Amount          int64           `gorm:"not null;default:0;comment:金额(分)" json:"amount"`
	WarehouseID     uint64          `gorm:"not null;comment:仓库ID" json:"warehouse_id"`
	ShippingAddress ShippingAddress `gorm:"type:json;comment:配送地址" json:"shipping_address"`
	Status          string          `gorm:"type:varchar(32);not null;comment:状态" json:"status"`
}

// WarehouseAllocation 仓库分配 (Value Object)
type WarehouseAllocation struct {
	ProductID   uint64  `json:"product_id"`
	Quantity    int32   `json:"quantity"`
	WarehouseID uint64  `json:"warehouse_id"`
	Distance    float64 `json:"distance"`
}

// WarehouseAllocationArray defines a slice of WarehouseAllocation that implements sql.Scanner and driver.Valuer
type WarehouseAllocationArray []*WarehouseAllocation

func (a WarehouseAllocationArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *WarehouseAllocationArray) Scan(value interface{}) error {
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

// WarehouseAllocationPlan 仓库分配计划
type WarehouseAllocationPlan struct {
	gorm.Model
	OrderID     uint64                   `gorm:"not null;index;comment:订单ID" json:"order_id"`
	Allocations WarehouseAllocationArray `gorm:"type:json;comment:分配详情" json:"allocations"`
}
