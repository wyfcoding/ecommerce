package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。

	"gorm.io/gorm" // 导入GORM库。
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的map[string]interface{}类型作为JSON字符串存储到数据库，并从数据库读取。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口，将 JSONMap 转换为数据库可以存储的值（JSON字节数组）。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m) // 将map编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 JSONMap。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m) // 将JSON字节数组解码为map。
}

// OrderItem 值对象定义了订单中的一个商品项。
// 它是订单优化过程中的一个子组件。
type OrderItem struct {
	ProductID uint64 `json:"product_id"` // 商品ID。
	Quantity  int32  `json:"quantity"`   // 数量。
	Price     int64  `json:"price"`      // 单价（单位：分）。
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

// ShippingAddress 值对象定义了订单的配送地址信息。
type ShippingAddress struct {
	Name     string `json:"name"`     // 收货人姓名。
	Phone    string `json:"phone"`    // 手机号。
	Province string `json:"province"` // 省份。
	City     string `json:"city"`     // 城市。
	District string `json:"district"` // 区县。
	Address  string `json:"address"`  // 详细地址。
}

// Value 实现 driver.Valuer 接口，将 ShippingAddress 转换为数据库可以存储的值（JSON字节数组）。
func (s ShippingAddress) Value() (driver.Value, error) {
	return json.Marshal(s) // 将结构体编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 ShippingAddress。
func (s *ShippingAddress) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s) // 将JSON字节数组解码为结构体。
}

// Uint64Array 定义了一个 uint64 切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []uint64 类型作为JSON字符串存储到数据库，并从数据库读取。
type Uint64Array []uint64

// Value 实现 driver.Valuer 接口，将 Uint64Array 转换为数据库可以存储的值（JSON字节数组）。
func (a Uint64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 Uint64Array。
func (a *Uint64Array) Scan(value interface{}) error {
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

// MergedOrder 实体代表一个合并后的订单。
// 当多个小订单需要合并处理以优化物流或打包时使用。
type MergedOrder struct {
	gorm.Model                       // 嵌入gorm.Model。
	UserID           uint64          `gorm:"not null;index;comment:用户ID" json:"user_id"`                // 关联的用户ID，索引字段。
	OriginalOrderIDs Uint64Array     `gorm:"type:json;comment:原始订单ID列表" json:"original_order_ids"`      // 包含的原始订单ID列表，存储为JSON。
	Items            OrderItemArray  `gorm:"type:json;comment:订单项" json:"items"`                        // 合并后的订单商品项列表，存储为JSON。
	TotalAmount      int64           `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`     // 合并订单的总金额（单位：分）。
	DiscountAmount   int64           `gorm:"not null;default:0;comment:优惠金额(分)" json:"discount_amount"` // 合并订单的优惠金额（单位：分）。
	FinalAmount      int64           `gorm:"not null;default:0;comment:最终金额(分)" json:"final_amount"`    // 合并订单的最终金额（单位：分）。
	ShippingAddress  ShippingAddress `gorm:"type:json;comment:配送地址" json:"shipping_address"`            // 合并订单的配送地址，存储为JSON。
	Status           string          `gorm:"type:varchar(32);not null;comment:状态" json:"status"`        // 合并订单的状态。
}

// SplitOrder 实体代表一个拆分后的子订单。
// 当一个订单需要拆分成多个子订单以方便不同仓库发货或分批配送时使用。
type SplitOrder struct {
	gorm.Model                      // 嵌入gorm.Model。
	OriginalOrderID uint64          `gorm:"not null;index;comment:原始订单ID" json:"original_order_id"` // 关联的原始订单ID，索引字段。
	SplitIndex      int32           `gorm:"not null;comment:拆分序号" json:"split_index"`               // 拆分后的子订单序号。
	Items           OrderItemArray  `gorm:"type:json;comment:订单项" json:"items"`                     // 子订单包含的商品项列表，存储为JSON。
	Amount          int64           `gorm:"not null;default:0;comment:金额(分)" json:"amount"`         // 子订单金额（单位：分）。
	WarehouseID     uint64          `gorm:"not null;comment:仓库ID" json:"warehouse_id"`              // 分配到的仓库ID。
	ShippingAddress ShippingAddress `gorm:"type:json;comment:配送地址" json:"shipping_address"`         // 子订单的配送地址，存储为JSON。
	Status          string          `gorm:"type:varchar(32);not null;comment:状态" json:"status"`     // 子订单的状态。
}

// WarehouseAllocation 值对象定义了仓库分配的详细信息。
// 它是仓库分配计划中的一个子组件。
type WarehouseAllocation struct {
	ProductID   uint64  `json:"product_id"`   // 商品ID。
	Quantity    int32   `json:"quantity"`     // 分配数量。
	WarehouseID uint64  `json:"warehouse_id"` // 分配到的仓库ID。
	Distance    float64 `json:"distance"`     // 商品到收货地址的距离。
}

// WarehouseAllocationArray 定义了一个 WarehouseAllocation 结构体切片，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将 WarehouseAllocation 切片作为JSON字符串存储到数据库，并从数据库读取。
type WarehouseAllocationArray []*WarehouseAllocation

// Value 实现 driver.Valuer 接口，将 WarehouseAllocationArray 转换为数据库可以存储的值（JSON字节数组）。
func (a WarehouseAllocationArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 WarehouseAllocationArray。
func (a *WarehouseAllocationArray) Scan(value interface{}) error {
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

// WarehouseAllocationPlan 实体代表一个订单的仓库分配计划。
// 它包含了订单中的每个商品应该从哪个仓库出货的详细信息。
type WarehouseAllocationPlan struct {
	gorm.Model                           // 嵌入gorm.Model。
	OrderID     uint64                   `gorm:"not null;index;comment:订单ID" json:"order_id"` // 关联的订单ID，索引字段。
	Allocations WarehouseAllocationArray `gorm:"type:json;comment:分配详情" json:"allocations"`   // 仓库分配的详细列表，存储为JSON。
}
