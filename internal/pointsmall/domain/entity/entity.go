package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// PointsProductStatus 积分商品状态
type PointsProductStatus int8

const (
	PointsProductStatusOffline PointsProductStatus = 0 // 下架：商品不可购买。
	PointsProductStatusOnline  PointsProductStatus = 1 // 上架：商品可供用户兑换。
)

// PointsProduct 积分商品实体。
// 它包含了积分商品的名称、描述、所需积分、库存和销售情况等信息。
type PointsProduct struct {
	gorm.Model                       // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name         string              `gorm:"type:varchar(255);not null;comment:商品名称" json:"name"`      // 商品名称。
	Description  string              `gorm:"type:text;comment:商品描述" json:"description"`                // 商品描述。
	ImageURL     string              `gorm:"type:varchar(255);comment:图片URL" json:"image_url"`         // 商品图片URL。
	Points       int64               `gorm:"not null;comment:所需积分" json:"points"`                      // 兑换此商品所需的积分数量。
	Stock        int32               `gorm:"not null;default:0;comment:库存" json:"stock"`               // 商品当前库存数量。
	SoldCount    int32               `gorm:"not null;default:0;comment:已售数量" json:"sold_count"`        // 商品已售出数量。
	LimitPerUser int32               `gorm:"not null;default:0;comment:每人限购" json:"limit_per_user"`    // 每位用户兑换此商品的限购数量，0表示不限购。
	Status       PointsProductStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"` // 商品状态，默认为下架。
}

// PointsOrderStatus 积分订单状态
type PointsOrderStatus int8

const (
	PointsOrderStatusPending   PointsOrderStatus = 0 // 待发货：积分订单已创建，等待商品发出。
	PointsOrderStatusShipped   PointsOrderStatus = 1 // 已发货：商品已发出。
	PointsOrderStatusCompleted PointsOrderStatus = 2 // 已完成：商品已送达并完成。
	PointsOrderStatusCanceled  PointsOrderStatus = 3 // 已取消：积分订单被取消。
)

// PointsOrder 积分订单实体。
// 它记录了用户通过积分兑换商品的订单详情。
type PointsOrder struct {
	gorm.Model                    // 嵌入gorm.Model。
	OrderNo     string            `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"order_no"` // 订单编号，唯一索引，不允许为空。
	UserID      uint64            `gorm:"not null;index;comment:用户ID" json:"user_id"`                         // 兑换用户ID，索引字段。
	ProductID   uint64            `gorm:"not null;index;comment:商品ID" json:"product_id"`                      // 兑换商品ID，索引字段。
	ProductName string            `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`        // 商品名称。
	Quantity    int32             `gorm:"not null;comment:数量" json:"quantity"`                                // 兑换数量。
	Points      int64             `gorm:"not null;comment:单价积分" json:"points"`                                // 单个商品所需积分。
	TotalPoints int64             `gorm:"not null;comment:总积分" json:"total_points"`                           // 订单总计消耗积分。
	Status      PointsOrderStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`           // 订单状态，默认为待发货。
	Address     string            `gorm:"type:varchar(255);comment:收货地址" json:"address"`                      // 收货地址。
	Phone       string            `gorm:"type:varchar(20);comment:联系电话" json:"phone"`                         // 联系电话。
	Receiver    string            `gorm:"type:varchar(64);comment:收货人" json:"receiver"`                       // 收货人姓名。
	ShippedAt   *time.Time        `gorm:"comment:发货时间" json:"shipped_at"`                                     // 发货时间。
	CompletedAt *time.Time        `gorm:"comment:完成时间" json:"completed_at"`                                   // 完成时间。
}

// PointsAccount 积分账户实体。
// 它记录了用户的总积分和已使用积分情况。
type PointsAccount struct {
	gorm.Model         // 嵌入gorm.Model。
	UserID      uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`   // 用户ID，唯一索引，不允许为空。
	TotalPoints int64  `gorm:"not null;default:0;comment:总积分" json:"total_points"` // 用户拥有的总积分（包括已使用和可用）。
	UsedPoints  int64  `gorm:"not null;default:0;comment:已用积分" json:"used_points"` // 用户已使用的积分。
}

// PointsTransaction 积分流水实体。
// 它记录了用户积分的每一次变动，包括获取和花费。
type PointsTransaction struct {
	gorm.Model         // 嵌入gorm.Model。
	UserID      uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`        // 关联的用户ID，索引字段。
	Type        string `gorm:"type:varchar(32);not null;comment:类型" json:"type"`  // 交易类型，例如“earn”（获取）或“spend”（花费）。
	Points      int64  `gorm:"not null;comment:变动积分" json:"points"`               // 积分变动数量（正数表示增加，负数表示减少）。
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`   // 交易描述。
	RefID       string `gorm:"type:varchar(64);index;comment:关联ID" json:"ref_id"` // 关联的业务ID，例如积分订单号、活动ID等，索引字段。
}
