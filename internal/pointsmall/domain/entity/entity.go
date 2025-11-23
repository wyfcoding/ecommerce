package entity

import (
	"time"

	"gorm.io/gorm"
)

// PointsProductStatus 积分商品状态
type PointsProductStatus int8

const (
	PointsProductStatusOffline PointsProductStatus = 0 // 下架
	PointsProductStatusOnline  PointsProductStatus = 1 // 上架
)

// PointsProduct 积分商品实体
type PointsProduct struct {
	gorm.Model
	Name         string              `gorm:"type:varchar(255);not null;comment:商品名称" json:"name"`
	Description  string              `gorm:"type:text;comment:商品描述" json:"description"`
	ImageURL     string              `gorm:"type:varchar(255);comment:图片URL" json:"image_url"`
	Points       int64               `gorm:"not null;comment:所需积分" json:"points"`
	Stock        int32               `gorm:"not null;default:0;comment:库存" json:"stock"`
	SoldCount    int32               `gorm:"not null;default:0;comment:已售数量" json:"sold_count"`
	LimitPerUser int32               `gorm:"not null;default:0;comment:每人限购" json:"limit_per_user"`
	Status       PointsProductStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
}

// PointsOrderStatus 积分订单状态
type PointsOrderStatus int8

const (
	PointsOrderStatusPending   PointsOrderStatus = 0 // 待发货
	PointsOrderStatusShipped   PointsOrderStatus = 1 // 已发货
	PointsOrderStatusCompleted PointsOrderStatus = 2 // 已完成
	PointsOrderStatusCanceled  PointsOrderStatus = 3 // 已取消
)

// PointsOrder 积分订单实体
type PointsOrder struct {
	gorm.Model
	OrderNo     string            `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"order_no"`
	UserID      uint64            `gorm:"not null;index;comment:用户ID" json:"user_id"`
	ProductID   uint64            `gorm:"not null;index;comment:商品ID" json:"product_id"`
	ProductName string            `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`
	Quantity    int32             `gorm:"not null;comment:数量" json:"quantity"`
	Points      int64             `gorm:"not null;comment:单价积分" json:"points"`
	TotalPoints int64             `gorm:"not null;comment:总积分" json:"total_points"`
	Status      PointsOrderStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
	Address     string            `gorm:"type:varchar(255);comment:收货地址" json:"address"`
	Phone       string            `gorm:"type:varchar(20);comment:联系电话" json:"phone"`
	Receiver    string            `gorm:"type:varchar(64);comment:收货人" json:"receiver"`
	ShippedAt   *time.Time        `gorm:"comment:发货时间" json:"shipped_at"`
	CompletedAt *time.Time        `gorm:"comment:完成时间" json:"completed_at"`
}

// PointsAccount 积分账户
type PointsAccount struct {
	gorm.Model
	UserID      uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	TotalPoints int64  `gorm:"not null;default:0;comment:总积分" json:"total_points"`
	UsedPoints  int64  `gorm:"not null;default:0;comment:已用积分" json:"used_points"`
}

// PointsTransaction 积分流水
type PointsTransaction struct {
	gorm.Model
	UserID      uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Type        string `gorm:"type:varchar(32);not null;comment:类型" json:"type"` // earn, spend
	Points      int64  `gorm:"not null;comment:变动积分" json:"points"`
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`
	RefID       string `gorm:"type:varchar(64);index;comment:关联ID" json:"ref_id"` // order_no, etc.
}
