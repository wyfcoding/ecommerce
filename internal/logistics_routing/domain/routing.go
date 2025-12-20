package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringArray 定义了一个字符串切片类型，用于JSON存储。
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
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

// Carrier 实体代表一个配送商。
type Carrier struct {
	gorm.Model
	Name              string      `gorm:"type:varchar(255);not null;comment:配送商名称" json:"name"`
	Type              string      `gorm:"type:varchar(32);not null;comment:类型" json:"type"`
	BaseCost          int64       `gorm:"not null;default:0;comment:基础费用(分)" json:"base_cost"`
	WeightRate        float64     `gorm:"type:decimal(10,2);default:0.00;comment:每kg费用" json:"weight_rate"`
	DistanceRate      float64     `gorm:"type:decimal(10,2);default:0.00;comment:每km费用" json:"distance_rate"`
	BaseDeliveryTime  int32       `gorm:"not null;default:0;comment:基础配送时间(小时)" json:"base_delivery_time"`
	SupportedRegions  StringArray `gorm:"type:json;comment:支持地区" json:"supported_regions"`
	AvailableCapacity int32       `gorm:"not null;default:0;comment:可用容量(kg)" json:"available_capacity"`
	Rating            float64     `gorm:"type:decimal(3,2);default:5.00;comment:评分" json:"rating"`
	IsActive          bool        `gorm:"default:true;comment:是否激活" json:"is_active"`
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
type RouteOrder struct {
	OrderID       uint64 `json:"order_id"`
	CarrierID     uint64 `json:"carrier_id"`
	CarrierName   string `json:"carrier_name"`
	EstimatedCost int64  `json:"estimated_cost"`
	EstimatedTime int32  `json:"estimated_time"`
}

// RouteOrderArray 定义了 RouteOrder 结构体切片。
type RouteOrderArray []*RouteOrder

func (a RouteOrderArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *RouteOrderArray) Scan(value interface{}) error {
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

// OptimizedRoute 实体代表一个优化后的配送路线方案。
type OptimizedRoute struct {
	gorm.Model
	Orders      RouteOrderArray `gorm:"type:json;comment:订单列表" json:"orders"`
	OrderCount  int32           `gorm:"not null;default:0;comment:订单数量" json:"order_count"`
	TotalCost   int64           `gorm:"not null;default:0;comment:总费用(分)" json:"total_cost"`
	AverageCost int64           `gorm:"not null;default:0;comment:平均费用(分)" json:"average_cost"`
}

// RoutingStatistics 实体代表路由相关的统计数据。
type RoutingStatistics struct {
	gorm.Model
	TotalRoutes         int64     `gorm:"not null;default:0;comment:总路由数" json:"total_routes"`
	TotalOrders         int64     `gorm:"not null;default:0;comment:总订单数" json:"total_orders"`
	TotalCost           int64     `gorm:"not null;default:0;comment:总费用" json:"total_cost"`
	AverageCost         float64   `gorm:"type:decimal(10,2);default:0.00;comment:平均费用" json:"average_cost"`
	AverageDeliveryTime float64   `gorm:"type:decimal(10,2);default:0.00;comment:平均配送时间" json:"average_delivery_time"`
	StartDate           time.Time `gorm:"comment:开始时间" json:"start_date"`
	EndDate             time.Time `gorm:"comment:结束时间" json:"end_date"`
}
