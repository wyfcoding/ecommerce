package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口。
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
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

// OrderItem 值对象定义了订单中的一个商品项。
type OrderItem struct {
	ProductID uint64 `json:"product_id"`
	SkuID     uint64 `json:"sku_id"`
	Quantity  int32  `json:"quantity"`
	Price     int64  `json:"price"` // 单价（单位：分）。
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

// ShippingAddress 值对象定义了订单的配送地址信息。
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

func (s *ShippingAddress) Scan(value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// Uint64Array 定义了一个 uint64 切片类型。
type Uint64Array []uint64

func (a Uint64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *Uint64Array) Scan(value any) error {
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

// MergedOrder 实体代表一个合并后的订单。
type MergedOrder struct {
	ID               uint            `json:"id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	UserID           uint64          `json:"user_id"`
	OriginalOrderIDs Uint64Array     `json:"original_order_ids"`
	Items            OrderItemArray  `json:"items"`
	TotalAmount      int64           `json:"total_amount"`
	DiscountAmount   int64           `json:"discount_amount"`
	FinalAmount      int64           `json:"final_amount"`
	ShippingAddress  ShippingAddress `json:"shipping_address"`
	Status           string          `json:"status"`
}

// SplitOrder 实体代表一个拆分后的子订单。
type SplitOrder struct {
	ID              uint            `json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	OriginalOrderID uint64          `json:"original_order_id"`
	SplitIndex      int32           `json:"split_index"`
	Items           OrderItemArray  `json:"items"`
	Amount          int64           `json:"amount"`
	WarehouseID     uint64          `json:"warehouse_id"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	Status          string          `json:"status"`
}

// WarehouseAllocation 值对象定义了仓库分配的详细信息。
type WarehouseAllocation struct {
	ProductID   uint64  `json:"product_id"`
	Quantity    int32   `json:"quantity"`
	WarehouseID uint64  `json:"warehouse_id"`
	Distance    float64 `json:"distance"`
}

// WarehouseAllocationArray 定义了一个 WarehouseAllocation 结构体切片。
type WarehouseAllocationArray []*WarehouseAllocation

func (a WarehouseAllocationArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *WarehouseAllocationArray) Scan(value any) error {
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

// WarehouseAllocationPlan 实体代表一个订单的仓库分配计划。
type WarehouseAllocationPlan struct {
	ID          uint                     `json:"id"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	OrderID     uint64                   `json:"order_id"`
	Allocations WarehouseAllocationArray `json:"allocations"`
}
