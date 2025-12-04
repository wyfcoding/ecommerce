package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value interface{}) error {
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

// Carrier 实体代表一个配送商。
// 它包含了配送商的名称、类型、费用模型、服务区域和当前容量等信息。
type Carrier struct {
	gorm.Model                    // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name              string      `gorm:"type:varchar(255);not null;comment:配送商名称" json:"name"`               // 配送商名称。
	Type              string      `gorm:"type:varchar(32);not null;comment:类型" json:"type"`                   // 配送商类型，例如“express”（快递）、“standard”（标准）、“economy”（经济）。
	BaseCost          int64       `gorm:"not null;default:0;comment:基础费用(分)" json:"base_cost"`                // 配送的基础费用（单位：分）。
	WeightRate        float64     `gorm:"type:decimal(10,2);default:0.00;comment:每kg费用" json:"weight_rate"`   // 每公斤的额外费用。
	DistanceRate      float64     `gorm:"type:decimal(10,2);default:0.00;comment:每km费用" json:"distance_rate"` // 每公里的额外费用。
	BaseDeliveryTime  int32       `gorm:"not null;default:0;comment:基础配送时间(小时)" json:"base_delivery_time"`    // 基础配送时间（小时）。
	SupportedRegions  StringArray `gorm:"type:json;comment:支持地区" json:"supported_regions"`                    // 支持的配送地区列表，存储为JSON。
	AvailableCapacity int32       `gorm:"not null;default:0;comment:可用容量(kg)" json:"available_capacity"`      // 配送商的可用配送容量（公斤）。
	Rating            float64     `gorm:"type:decimal(3,2);default:5.00;comment:评分" json:"rating"`            // 配送商的服务评分。
	IsActive          bool        `gorm:"default:true;comment:是否激活" json:"is_active"`                         // 配送商是否处于活跃状态。
}

// SupportsRegion 检查配送商是否支持指定地区。
func (c *Carrier) SupportsRegion(region string) bool {
	for _, r := range c.SupportedRegions {
		if r == region {
			return true
		}
	}
	return false
}

// RouteOrder 结构体定义了优化路由中的单个订单信息。
// 它是一个值对象，通常嵌套在 OptimizedRoute 实体中。
type RouteOrder struct {
	OrderID       uint64 `json:"order_id"`       // 订单ID。
	CarrierID     uint64 `json:"carrier_id"`     // 分配的配送商ID。
	CarrierName   string `json:"carrier_name"`   // 分配的配送商名称。
	EstimatedCost int64  `json:"estimated_cost"` // 预估配送成本。
	EstimatedTime int32  `json:"estimated_time"` // 预估配送时间（小时）。
}

// RouteOrderArray 定义了一个 RouteOrder 结构体切片，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将 RouteOrder 切片作为JSON字符串存储到数据库，并从数据库读取。
type RouteOrderArray []*RouteOrder

// Value 实现 driver.Valuer 接口，将 RouteOrderArray 转换为数据库可以存储的值（JSON字节数组）。
func (a RouteOrderArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 RouteOrderArray。
func (a *RouteOrderArray) Scan(value interface{}) error {
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

// OptimizedRoute 实体代表一个优化后的配送路线方案。
// 它包含了分配给不同配送商的订单列表、总成本和平均成本等汇总信息。
type OptimizedRoute struct {
	gorm.Model                  // 嵌入gorm.Model。
	Orders      RouteOrderArray `gorm:"type:json;comment:订单列表" json:"orders"`                   // 优化路线中包含的订单列表，存储为JSON。
	OrderCount  int32           `gorm:"not null;default:0;comment:订单数量" json:"order_count"`     // 优化路线中的订单总数。
	TotalCost   int64           `gorm:"not null;default:0;comment:总费用(分)" json:"total_cost"`    // 优化路线的总配送成本（单位：分）。
	AverageCost int64           `gorm:"not null;default:0;comment:平均费用(分)" json:"average_cost"` // 优化路线的平均配送成本（单位：分）。
}

// RoutingStatistics 实体代表路由相关的统计数据。
// 此实体通常不直接持久化，而是通过查询计算得出。
type RoutingStatistics struct {
	gorm.Model                    // 嵌入gorm.Model。
	TotalRoutes         int64     `gorm:"not null;default:0;comment:总路由数" json:"total_routes"`                         // 总优化路由方案数量。
	TotalOrders         int64     `gorm:"not null;default:0;comment:总订单数" json:"total_orders"`                         // 涉及优化路由的订单总数。
	TotalCost           int64     `gorm:"not null;default:0;comment:总费用" json:"total_cost"`                            // 所有优化路由的总成本。
	AverageCost         float64   `gorm:"type:decimal(10,2);default:0.00;comment:平均费用" json:"average_cost"`            // 平均每条优化路由的成本。
	AverageDeliveryTime float64   `gorm:"type:decimal(10,2);default:0.00;comment:平均配送时间" json:"average_delivery_time"` // 平均配送时间。
	StartDate           time.Time `gorm:"comment:开始时间" json:"start_date"`                                              // 统计的开始时间。
	EndDate             time.Time `gorm:"comment:结束时间" json:"end_date"`                                                // 统计的结束时间。
}
