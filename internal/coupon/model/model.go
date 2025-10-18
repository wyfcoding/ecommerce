package model

import (
	"time"

	"gorm.io/gorm"
)

// DiscountType 定义了优惠方式的类型
type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "FIXED"      // 固定金额
	DiscountTypePercentage DiscountType = "PERCENTAGE" // 百分比
)

// CouponType 定义了优惠券的类型
type CouponType string

const (
	CouponTypeGeneric  CouponType = "GENERIC"  // 通用券 (任何人可用)
	CouponTypeSpecific CouponType = "SPECIFIC" //- 用户专属券
)

// CouponStatus 定义了用户持有的优惠券的状态
type CouponStatus string

const (
	StatusUnused  CouponStatus = "UNUSED"  // 未使用
	StatusUsed    CouponStatus = "USED"    // 已使用
	StatusExpired CouponStatus = "EXPIRED" // 已过期
)

// CouponDefinition 优惠券定义模型
// 描述了一类优惠券的通用规则
type CouponDefinition struct {
	ID             uint         `gorm:"primarykey" json:"id"`
	Code           string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"` // 优惠券码
	Description    string       `gorm:"type:varchar(255);not null" json:"description"`    // 优惠券描述
	Type           CouponType   `gorm:"type:varchar(20);not null" json:"type"`
	DiscountType   DiscountType `gorm:"type:varchar(20);not null" json:"discount_type"`
	DiscountValue  float64      `gorm:"type:decimal(10,2);not null" json:"discount_value"` // 优惠值 (金额或百分比)
	MinSpend       float64      `gorm:"type:decimal(10,2);default:0" json:"min_spend"`      // 最低消费金额
	MaxDiscount    float64      `gorm:"type:decimal(10,2)" json:"max_discount"`     // 最高优惠金额 (主要用于百分比优惠)
	ValidFrom      time.Time    `gorm:"not null" json:"valid_from"`                     // 有效期开始时间
	ValidTo        time.Time    `gorm:"not null" json:"valid_to"`                       // 有效期结束时间
	TotalQuantity  int          `gorm:"not null;default:0" json:"total_quantity"`       // 总发行量 (0表示无限)
	IssuedQuantity int          `gorm:"not null;default:0" json:"issued_quantity"`      // 已发行数量
	IsActive       bool         `gorm:"not null;default:true" json:"is_active"`        // 是否激活
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// UserCoupon 用户优惠券模型
// 记录了用户与优惠券的关联
type UserCoupon struct {
	ID                 uint         `gorm:"primarykey" json:"id"`
	UserID             uint         `gorm:"not null;index" json:"user_id"`
	CouponDefinitionID uint         `gorm:"not null;index" json:"coupon_definition_id"`
	CouponDefinition   CouponDefinition `gorm:"foreignKey:CouponDefinitionID" json:"coupon_definition"`
	Status             CouponStatus `gorm:"type:varchar(20);not null;default:'UNUSED'" json:"status"`
	RedeemedAt         *time.Time   `json:"redeemed_at"` // 核销时间
	OrderID            *uint        `gorm:"index" json:"order_id"` // 核销时关联的订单ID
	AssignedAt         time.Time    `gorm:"not null" json:"assigned_at"`
}

// TableName 自定义表名
func (CouponDefinition) TableName() string {
	return "coupon_definitions"
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}
