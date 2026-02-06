package domain

import (
	"fmt"  // 导入格式化包。
	"time" // 导入时间包。
)

// CouponType 定义了优惠券的类型。
type CouponType int

const (
	CouponTypeDiscount CouponType = 1 // 折扣券：提供一定比例的折扣，如8折。
	CouponTypeCash     CouponType = 2 // 现金券：直接抵扣固定金额。
	CouponTypeGift     CouponType = 3 // 礼品券：兑换实物礼品或虚拟商品。
	CouponTypeExchange CouponType = 4 // 兑换券：兑换指定商品。
)

// CouponStatus 定义了优惠券的生命周期状态。
type CouponStatus int

const (
	CouponStatusDraft    CouponStatus = 1 // 草稿：优惠券已创建但未启用。
	CouponStatusActive   CouponStatus = 2 // 激活：优惠券已启用，可以发放和使用。
	CouponStatusInactive CouponStatus = 3 // 停用：优惠券已停用，不能发放和使用。
	CouponStatusExpired  CouponStatus = 4 // 过期：优惠券已超过有效期。
	CouponStatusDeleted  CouponStatus = 5 // 删除：优惠券已删除。
)

// Coupon 实体是优惠券模块的聚合根。
// 它代表一个优惠券模板，包含了优惠券的规则、状态、发行和使用统计等信息。
type Coupon struct {
	ID                    uint64       `json:"id"`                      // 优惠券ID。
	CreatedAt             time.Time    `json:"created_at"`              // 创建时间。
	UpdatedAt             time.Time    `json:"updated_at"`              // 更新时间。
	CouponNo              string       `json:"coupon_no"`               // 优惠券编号。
	Name                  string       `json:"name"`                    // 优惠券名称。
	Description           string       `json:"description"`             // 优惠券描述。
	Type                  CouponType   `json:"type"`                    // 优惠券类型。
	Status                CouponStatus `json:"status"`                  // 优惠券状态。
	DiscountAmount        int64        `json:"discount_amount"`         // 优惠金额（单位：分）或折扣比例（0-100）。
	MinOrderAmount        int64        `json:"min_order_amount"`        // 使用优惠券的最低订单金额（单位：分）。
	MaxDiscount           int64        `json:"max_discount"`            // 折扣券的最大优惠金额。
	ValidFrom             time.Time    `json:"valid_from"`              // 优惠券有效期开始时间。
	ValidTo               time.Time    `json:"valid_to"`                // 优惠券有效期结束时间。
	UsageLimit            int32        `json:"usage_limit"`             // 优惠券的总发行数量，0表示不限制。
	UsagePerUser          int32        `json:"usage_per_user"`          // 每个用户可以领取的优惠券数量。
	TotalIssued           int32        `json:"total_issued"`            // 已经发放的数量。
	TotalUsed             int32        `json:"total_used"`              // 已经使用的数量。
	ConditionExpr         string       `json:"condition_expr"`          // 判定表达式，用于规则引擎判定。
	ApplicableScope       string       `json:"applicable_scope"`        // 优惠券适用范围。
	ApplicableProductIDs  []uint64     `json:"applicable_product_ids"`  // 适用商品ID列表。
	ExcludedProductIDs    []uint64     `json:"excluded_product_ids"`    // 排除商品ID列表。
	ApplicableCategoryIDs []string     `json:"applicable_category_ids"` // 适用分类ID列表。
	ExcludedCategoryIDs   []string     `json:"excluded_category_ids"`   // 排除分类ID列表。
	UserTierRequirement   string       `json:"usertier_requirement"`    // 会员等级要求。
	CanStack              bool         `json:"can_stack"`               // 是否允许与其他优惠券叠加使用。
	StackingRules         string       `json:"stacking_rules"`          // 叠加规则(JSON)。
}

// UserCoupon 实体代表用户拥有的优惠券。
// 它是Coupon实体的一个实例，与特定用户关联。
type UserCoupon struct {
	ID         uint64     `json:"id"`          // 用户优惠券ID。
	CreatedAt  time.Time  `json:"created_at"`  // 创建时间。
	UpdatedAt  time.Time  `json:"updated_at"`  // 更新时间。
	UserID     uint64     `json:"user_id"`     // 优惠券所属的用户ID。
	CouponID   uint64     `json:"coupon_id"`   // 关联的优惠券模板ID。
	CouponNo   string     `json:"coupon_no"`   // 优惠券的唯一编号。
	Status     string     `json:"status"`      // 用户优惠券状态。
	UsedAt     *time.Time `json:"used_at"`     // 优惠券使用时间。
	OrderID    string     `json:"order_id"`    // 使用该优惠券的订单ID。
	ReceivedAt time.Time  `json:"received_at"` // 用户领取优惠券的时间。
}

// CouponActivity 实体是优惠券模块的聚合根，代表一个优惠券营销活动。
// 一个活动可以关联多个优惠券模板。
type CouponActivity struct {
	ID          uint64    `json:"id"`          // 活动ID。
	CreatedAt   time.Time `json:"created_at"`  // 创建时间。
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间。
	Name        string    `json:"name"`        // 活动名称。
	Description string    `json:"description"` // 活动描述。
	StartTime   time.Time `json:"start_time"`  // 活动开始时间。
	EndTime     time.Time `json:"end_time"`    // 活动结束时间。
	CouponIDs   []uint64  `json:"coupon_ids"`  // 活动关联的优惠券模板ID列表。
	Status      string    `json:"status"`      // 活动状态。
}

// NewCoupon 创建并返回一个新的 Coupon 实体实例。
// name, description: 优惠券名称和描述。
// couponType: 优惠券类型。
// discountAmount: 优惠金额。
// minOrderAmount: 最低订单金额。
func NewCoupon(name, description string, couponType CouponType, discountAmount, minOrderAmount int64) *Coupon {
	now := time.Now()
	validTo := now.AddDate(0, 3, 0) // 默认有效期为3个月。

	return &Coupon{
		CouponNo:             generateCouponNo(), // 生成唯一的优惠券编号。
		Name:                 name,
		Description:          description,
		Type:                 couponType,
		Status:               CouponStatusDraft, // 初始状态为草稿。
		DiscountAmount:       discountAmount,
		MinOrderAmount:       minOrderAmount,
		ValidFrom:            now,        // 有效期从当前时间开始。
		ValidTo:              validTo,    // 默认有效期结束时间。
		UsageLimit:           10000,      // 默认总发行量为10000张。
		UsagePerUser:         1,          // 默认每人限领1张。
		ApplicableProductIDs: []uint64{}, // 初始化适用商品ID列表。
	}
}

// Activate 激活优惠券，将其状态从草稿变更为激活。
func (c *Coupon) Activate() error {
	if c.Status != CouponStatusDraft {
		return fmt.Errorf("only draft coupon can be activated") // 只有草稿状态的优惠券才能被激活。
	}
	c.Status = CouponStatusActive // 状态更新为激活。
	return nil
}

// Deactivate 停用优惠券，将其状态从激活变更为停用。
func (c *Coupon) Deactivate() error {
	if c.Status != CouponStatusActive {
		return fmt.Errorf("only active coupon can be deactivated") // 只有激活状态的优惠券才能被停用。
	}
	c.Status = CouponStatusInactive // 状态更新为停用。
	return nil
}

// CheckAvailability 检查优惠券当前是否可用（是否激活、是否在有效期内、是否达到发行上限）。
func (c *Coupon) CheckAvailability() error {
	if c.Status != CouponStatusActive {
		return fmt.Errorf("coupon is not active") // 优惠券未激活。
	}

	now := time.Now()
	if now.Before(c.ValidFrom) {
		return fmt.Errorf("coupon is not yet valid") // 优惠券尚未生效。
	}
	if now.After(c.ValidTo) {
		return fmt.Errorf("coupon has expired") // 优惠券已过期。
	}

	if c.UsageLimit > 0 && c.TotalIssued >= c.UsageLimit { // 检查总发行量是否已达上限。
		return fmt.Errorf("coupon usage limit reached")
	}

	return nil
}

// Issue 发放优惠券，增加已发行量。
func (c *Coupon) Issue(quantity int32) {
	c.TotalIssued += quantity
}

// Use 使用优惠券，增加已使用量。
func (c *Coupon) Use() {
	c.TotalUsed++
}

// NewUserCoupon 创建并返回一个新的 UserCoupon 实体实例。
// userID: 领取优惠券的用户ID。
// couponID: 关联的优惠券模板ID。
// couponNo: 关联优惠券的编号。
func NewUserCoupon(userID, couponID uint64, couponNo string) *UserCoupon {
	return &UserCoupon{
		UserID:     userID,
		CouponID:   couponID,
		CouponNo:   couponNo,
		Status:     "unused",   // 初始状态为未使用。
		ReceivedAt: time.Now(), // 记录领取时间。
	}
}

// Use 使用用户优惠券，将其状态从未用变更为已用，并记录使用时间和订单ID。
func (u *UserCoupon) Use(orderID string) error {
	if u.Status != "unused" {
		return fmt.Errorf("coupon already used or expired") // 优惠券已使用或已过期。
	}
	u.Status = "used" // 状态更新为已使用。
	now := time.Now()
	u.UsedAt = &now     // 记录使用时间。
	u.OrderID = orderID // 记录关联的订单ID。
	return nil
}

// NewCouponActivity 创建并返回一个新的 CouponActivity 实体实例。
// name, description: 活动名称和描述。
// startTime, endTime: 活动的开始和结束时间。
// couponIDs: 活动关联的优惠券模板ID列表。
func NewCouponActivity(name, description string, startTime, endTime time.Time, couponIDs []uint64) *CouponActivity {
	return &CouponActivity{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		CouponIDs:   couponIDs,
		Status:      "active", // 默认状态为进行中。
	}
}

// generateCouponNo 是一个辅助函数，用于生成唯一的优惠券编号。
func generateCouponNo() string {
	// 使用当前时间的纳秒值作为基础，生成一个前缀为“CPN”的编号。
	return fmt.Sprintf("CPN%d", time.Now().UnixNano())
}
