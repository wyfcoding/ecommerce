package data

import (
	"time"

	"gorm.io/gorm"
)

<<<<<<< HEAD
// PriceRule 定价规则 (例如，基础价格，特殊价格)。
type PriceRule struct {
	gorm.Model
	ProductID uint64 `gorm:"index;not null;comment:商品ID" json:"productId"`
	SKUID     uint64 `gorm:"index;comment:SKU ID" json:"skuId"` // 可选
	RuleType  string `gorm:"not null;size:50;comment:规则类型 (e.g., BASE, SPECIAL)" json:"ruleType"`
	Price     uint64 `gorm:"not null;comment:价格 (分)" json:"price"`
	ValidFrom *time.Time `gorm:"comment:生效时间" json:"validFrom"`
	ValidTo   *time.Time `gorm:"comment:失效时间" json:"validTo"`
	Priority  int32 `gorm:"comment:优先级" json:"priority"`
	// 添加其他字段，如地区、用户细分等。
}

// Discount 折扣 (例如，来自促销或优惠券)。
type Discount struct {
	gorm.Model
	DiscountType  string `gorm:"not null;size:50;comment:折扣类型 (e.g., COUPON, PROMOTION)" json:"discountType"`
	DiscountID    uint64 `gorm:"comment:关联的优惠券/促销ID" json:"discountId"`
	Amount        uint64 `gorm:"not null;comment:优惠金额 (分)" json:"amount"`
	AppliedTo     string `gorm:"size:50;comment:应用对象 (e.g., ORDER, ITEM)" json:"appliedTo"`
	AppliedItemID uint64 `gorm:"comment:应用到的商品ID (如果AppliedTo是ITEM)" json:"appliedItemId"`
	// 添加其他字段，如描述、代码等。
}

// TableName 指定 PriceRule 的表名。
=======
// PriceRule represents a pricing rule (e.g., base price, special price).
type PriceRule struct {
	gorm.Model
	ProductID   uint64    `gorm:"index;not null;comment:商品ID" json:"productId"`
	SKUID       uint64    `gorm:"index;comment:SKU ID" json:"skuId"` // Optional
	RuleType    string    `gorm:"not null;size:50;comment:规则类型 (e.g., BASE, SPECIAL)" json:"ruleType"`
	Price       uint64    `gorm:"not null;comment:价格 (分)" json:"price"`
	ValidFrom   *time.Time `gorm:"comment:生效时间" json:"validFrom"`
	ValidTo     *time.Time `gorm:"comment:失效时间" json:"validTo"`
	Priority    int32     `gorm:"comment:优先级" json:"priority"`
	// Add other fields like region, user segment, etc.
}

// Discount represents a discount applied (e.g., from a promotion or coupon).
type Discount struct {
	gorm.Model
	DiscountType string    `gorm:"not null;size:50;comment:折扣类型 (e.g., COUPON, PROMOTION)" json:"discountType"`
	DiscountID   uint64    `gorm:"comment:关联的优惠券/促销ID" json:"discountId"`
	Amount       uint64    `gorm:"not null;comment:优惠金额 (分)" json:"amount"`
	AppliedTo    string    `gorm:"size:50;comment:应用对象 (e.g., ORDER, ITEM)" json:"appliedTo"`
	AppliedItemID uint64   `gorm:"comment:应用到的商品ID (如果AppliedTo是ITEM)" json:"appliedItemId"`
	// Add other fields like description, code, etc.
}

// TableName specifies the table name for PriceRule.
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
func (PriceRule) TableName() string {
	return "price_rules"
}

<<<<<<< HEAD
// TableName 指定 Discount 的表名。
=======
// TableName specifies the table name for Discount.
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
func (Discount) TableName() string {
	return "discounts"
}
