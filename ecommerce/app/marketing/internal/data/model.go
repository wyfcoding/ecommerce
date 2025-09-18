package data

import (
	"database/sql/driver"
	"ecommerce/ecommerce/app/marketing/internal/biz"
	"encoding/json"
	"time"
)

// --- JSON Field Handlers ---
type JSONUint64Array []uint64

func (a *JSONUint64Array) Scan(value interface{}) error { return json.Unmarshal(value.([]byte), a) }
func (a JSONUint64Array) Value() (driver.Value, error)  { return json.Marshal(a) }

type JSONRuleSet biz.RuleSet                        // 直接复用 biz 层的 RuleSet 结构体
func (r *JSONRuleSet) Scan(value interface{}) error { return json.Unmarshal(value.([]byte), r) }
func (r JSONRuleSet) Value() (driver.Value, error)  { return json.Marshal(r) }

// --- GORM Models ---
type CouponTemplate struct {
	ID                  uint64          `gorm:"primarykey"`
	Title               string          `gorm:"size:255;not null"`
	Type                int8            `gorm:"not null"`
	ScopeType           int8            `gorm:"not null"`
	ScopeIDs            JSONUint64Array `gorm:"type:json"`
	Rules               JSONRuleSet     `gorm:"type:json;not null"`
	TotalQuantity       uint            `gorm:"not null;default:0"`
	IssuedQuantity      uint            `gorm:"not null;default:0"`
	PerUserLimit        uint8           `gorm:"not null;default:1"`
	ValidityType        int8            `gorm:"not null"`
	ValidFrom           *time.Time
	ValidTo             *time.Time
	ValidDaysAfterClaim uint
	Status              int8 `gorm:"not null;default:1"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (CouponTemplate) TableName() string {
	return "coupon_template"
}

type UserCoupon struct {
	ID         uint64    `gorm:"primarykey"`
	TemplateID uint64    `gorm:"not null"`
	UserID     uint64    `gorm:"index:idx_user_id_status;not null"`
	CouponCode string    `gorm:"uniqueIndex:uk_coupon_code;size:64;not null"`
	Status     int8      `gorm:"index:idx_user_id_status;not null;default:1"`
	ClaimedAt  time.Time `gorm:"not null"`
	ValidFrom  time.Time `gorm:"not null"`
	ValidTo    time.Time `gorm:"not null"`
	OrderID    uint64
	UsedAt     *time.Time
}

func (UserCoupon) TableName() string {
	return "user_coupon"
}
