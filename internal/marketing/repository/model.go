package repository

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONUint64Array 是一个自定义类型，用于将 []uint64 json 化存入数据库。
type JSONUint64Array []uint64

func (a *JSONUint64Array) Scan(value interface{}) error { return json.Unmarshal(value.([]byte), a) }
func (a JSONUint64Array) Value() (driver.Value, error)  { return json.Marshal(a) }

// JSONRuleSet 是一个自定义类型，用于将 biz.RuleSet json 化存入数据库。
type JSONRuleSet biz.RuleSet

func (r *JSONRuleSet) Scan(value interface{}) error { return json.Unmarshal(value.([]byte), r) }
func (r JSONRuleSet) Value() (driver.Value, error)  { return json.Marshal(r) }

// CouponTemplate 对应数据库中的 `coupon_templates` 表。
type CouponTemplate struct {
	gorm.Model
	Title               string          `gorm:"not null;size:255;comment:优惠券标题" json:"title"`
	Type                int8            `gorm:"not null;comment:1-满减, 2-折扣, 3-无门槛" json:"type"`
	ScopeType           int8            `gorm:"not null;comment:1-全场, 2-指定分类, 3-指定SPU" json:"scopeType"`
	ScopeIDs            JSONUint64Array `gorm:"type:json;comment:适用范围ID列表" json:"scopeIds"`
	Rules               JSONRuleSet     `gorm:"not null;type:json;comment:优惠规则" json:"rules"`
	TotalQuantity       uint            `gorm:"not null;default:0;comment:总发行量, 0为不限量" json:"totalQuantity"`
	IssuedQuantity      uint            `gorm:"not null;default:0;comment:已领取数量" json:"issuedQuantity"`
	PerUserLimit        uint8           `gorm:"not null;default:1;comment:每人限领张数" json:"perUserLimit"`
	ValidityType        int8            `gorm:"not null;comment:1-固定时间, 2-领取后生效" json:"validityType"`
	ValidFrom           *time.Time      `gorm:"comment:固定有效期-开始时间" json:"validFrom"`
	ValidTo             *time.Time      `gorm:"comment:固定有效期-结束时间" json:"validTo"`
	ValidDaysAfterClaim uint            `gorm:"comment:领取后N天内有效" json:"validDaysAfterClaim"`
	Status              int8            `gorm:"index;not null;default:1;comment:1-可用, 2-禁用" json:"status"`
}

// UserCoupon 对应数据库中的 `user_coupons` 表。
type UserCoupon struct {
	gorm.Model
	TemplateID uint64    `gorm:"index;not null;comment:外键,关联coupon_template" json:"templateId"`
	UserID     uint64    `gorm:"index:idx_user_status;not null;comment:用户ID" json:"userId"`
	CouponCode string    `gorm:"uniqueIndex;not null;size:64;comment:唯一的券码" json:"couponCode"`
	Status     int8      `gorm:"index:idx_user_status;not null;default:1;comment:1-未使用, 2-已使用, 3-已过期" json:"status"`
	ClaimedAt  time.Time `gorm:"not null;comment:领取时间" json:"claimedAt"`
	ValidFrom  time.Time `gorm:"not null;comment:有效期-开始时间" json:"validFrom"`
	ValidTo    time.Time `gorm:"not null;comment:有效期-结束时间" json:"validTo"`
}

// Promotion 对应数据库中的 `promotions` 表。
type Promotion struct {
	gorm.Model
	Name           string          `gorm:"not null;size:255;comment:促销名称" json:"name"`
	Type           int8            `gorm:"not null;comment:1: 限时折扣, 2: 满减活动, 3: 买赠" json:"type"`
	Description    string          `gorm:"type:text;comment:促销描述" json:"description"`
	StartTime      *time.Time      `gorm:"comment:开始时间" json:"startTime"`
	EndTime        *time.Time      `gorm:"comment:结束时间" json:"endTime"`
	Status         *int8           `gorm:"index;default:3;comment:1: 进行中, 2: 已结束, 3: 未开始, 4: 已禁用" json:"status"`
	ProductIDs     JSONUint64Array `gorm:"type:json;comment:关联的商品ID列表" json:"productIds"`
	DiscountValue  uint64          `gorm:"comment:折扣值或满减金额" json:"discountValue"`
	MinOrderAmount uint64          `gorm:"comment:最小订单金额" json:"minOrderAmount"`
}

func (CouponTemplate) TableName() string {
	return "coupon_templates"
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}

func (Promotion) TableName() string {
	return "promotions"
}
