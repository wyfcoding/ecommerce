package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// CouponType 优惠券类型
type CouponType string

const (
	CouponTypePercentage   CouponType = "percentage"
	CouponTypeFixed        CouponType = "fixed"
	CouponTypeFreeShipping CouponType = "free_shipping"
)

// CouponStatus 优惠券状态
type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "active"
	CouponStatusInactive CouponStatus = "inactive"
	CouponStatusRevoked  CouponStatus = "revoked"
	CouponStatusExpired  CouponStatus = "expired"
)

// StringArray 定义了一个字符串切片，实现了 sql.Scanner 和 driver.Valuer 接口
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value any) error {
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

// Uint64Array 定义了一个 uint64 切片，实现了 sql.Scanner 和 driver.Valuer 接口
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

// Coupon 优惠券实体
type Coupon struct {
	gorm.Model
	Code                 string       `gorm:"type:varchar(64);uniqueIndex;not null;comment:优惠券代码" json:"code"`
	Type                 CouponType   `gorm:"type:varchar(32);not null;comment:类型" json:"type"`
	DiscountValue        int64        `gorm:"not null;comment:折扣值(分/百分比)" json:"discount_value"`
	MinPurchaseAmount    int64        `gorm:"not null;default:0;comment:最低消费金额(分)" json:"min_purchase_amount"`
	MaxDiscountAmount    int64        `gorm:"not null;default:0;comment:最大折扣金额(分)" json:"max_discount_amount"`
	ValidFrom            time.Time    `gorm:"not null;comment:生效时间" json:"valid_from"`
	ValidUntil           time.Time    `gorm:"not null;comment:过期时间" json:"valid_until"`
	TotalQuantity        int64        `gorm:"not null;default:0;comment:总数量" json:"total_quantity"`
	UsedQuantity         int64        `gorm:"not null;default:0;comment:已使用数量" json:"used_quantity"`
	PerUserLimit         int64        `gorm:"not null;default:1;comment:每人限领" json:"per_user_limit"`
	ApplicableCategories StringArray  `gorm:"type:json;comment:适用分类" json:"applicable_categories"`
	ExcludedCategories   StringArray  `gorm:"type:json;comment:排除分类" json:"excluded_categories"`
	ApplicableProducts   Uint64Array  `gorm:"type:json;comment:适用商品" json:"applicable_products"`
	ExcludedProducts     Uint64Array  `gorm:"type:json;comment:排除商品" json:"excluded_products"`
	UserTierRequirement  string       `gorm:"type:varchar(32);comment:会员等级要求" json:"usertier_requirement"`
	Status               CouponStatus `gorm:"type:varchar(32);default:'active';comment:状态" json:"status"`
}

// NewCoupon 创建优惠券
func NewCoupon(code string, couponType CouponType, discountValue int64, validFrom, validUntil time.Time, totalQuantity int64) *Coupon {
	return &Coupon{
		Code:          code,
		Type:          couponType,
		DiscountValue: discountValue,
		ValidFrom:     validFrom,
		ValidUntil:    validUntil,
		TotalQuantity: totalQuantity,
		Status:        CouponStatusActive,
	}
}

// IsValid 检查优惠券是否有效
func (c *Coupon) IsValid() bool {
	now := time.Now()
	return c.Status == CouponStatusActive &&
		now.After(c.ValidFrom) &&
		now.Before(c.ValidUntil) &&
		(c.TotalQuantity == 0 || c.UsedQuantity < c.TotalQuantity)
}

// CouponUsage 优惠券使用记录
type CouponUsage struct {
	gorm.Model
	UserID   uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	CouponID uint64    `gorm:"not null;index;comment:优惠券ID" json:"coupon_id"`
	OrderID  uint64    `gorm:"not null;index;comment:订单ID" json:"order_id"`
	Code     string    `gorm:"type:varchar(64);not null;comment:优惠券代码" json:"code"`
	UsedAt   time.Time `gorm:"not null;comment:使用时间" json:"used_at"`
}

// CouponStatistics 优惠券统计
type CouponStatistics struct {
	gorm.Model
	CouponID      uint64  `gorm:"not null;uniqueIndex;comment:优惠券ID" json:"coupon_id"`
	Code          string  `gorm:"type:varchar(64);not null;comment:优惠券代码" json:"code"`
	TotalQuantity int64   `gorm:"not null;default:0;comment:总数量" json:"total_quantity"`
	UsedQuantity  int64   `gorm:"not null;default:0;comment:已使用数量" json:"used_quantity"`
	UsageRate     float64 `gorm:"type:decimal(5,2);default:0.00;comment:使用率" json:"usage_rate"`
	UniqueUsers   int64   `gorm:"not null;default:0;comment:独立用户数" json:"unique_users"`
	TotalDiscount int64   `gorm:"not null;default:0;comment:总优惠金额" json:"total_discount"`
}
