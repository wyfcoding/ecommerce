package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CouponType 优惠券类型
type CouponType int

const (
	CouponTypeDiscount CouponType = 1 // 折扣券
	CouponTypeCash     CouponType = 2 // 现金券
	CouponTypeGift     CouponType = 3 // 礼品券
	CouponTypeExchange CouponType = 4 // 兑换券
)

// CouponStatus 优惠券状态
type CouponStatus int

const (
	CouponStatusDraft    CouponStatus = 1 // 草稿
	CouponStatusActive   CouponStatus = 2 // 激活
	CouponStatusInactive CouponStatus = 3 // 停用
	CouponStatusExpired  CouponStatus = 4 // 过期
	CouponStatusDeleted  CouponStatus = 5 // 删除
)

// Coupon 优惠券聚合根
type Coupon struct {
	gorm.Model
	CouponNo        string       `gorm:"type:varchar(64);uniqueIndex;not null;comment:优惠券编号" json:"coupon_no"`
	Name            string       `gorm:"type:varchar(255);not null;comment:名称" json:"name"`
	Description     string       `gorm:"type:text;comment:描述" json:"description"`
	Type            CouponType   `gorm:"default:1;comment:类型" json:"type"`
	Status          CouponStatus `gorm:"default:1;comment:状态" json:"status"`
	DiscountAmount  int64        `gorm:"comment:折扣金额/比例" json:"discount_amount"`
	MinOrderAmount  int64        `gorm:"comment:最低订单金额" json:"min_order_amount"`
	MaxDiscount     int64        `gorm:"comment:最大折扣金额" json:"max_discount"`
	ValidFrom       time.Time    `gorm:"comment:有效期开始" json:"valid_from"`
	ValidTo         time.Time    `gorm:"comment:有效期结束" json:"valid_to"`
	UsageLimit      int32        `gorm:"default:0;comment:总发行量" json:"usage_limit"`
	UsagePerUser    int32        `gorm:"default:1;comment:每人限领" json:"usage_per_user"`
	TotalIssued     int32        `gorm:"default:0;comment:已发行量" json:"total_issued"`
	TotalUsed       int32        `gorm:"default:0;comment:已使用量" json:"total_used"`
	ApplicableScope string       `gorm:"type:varchar(255);comment:适用范围" json:"applicable_scope"`
	ApplicableIDs   []uint64     `gorm:"type:json;serializer:json;comment:适用ID列表" json:"applicable_ids"`
}

// UserCoupon 用户优惠券实体
type UserCoupon struct {
	gorm.Model
	UserID     uint64     `gorm:"not null;index;comment:用户ID" json:"user_id"`
	CouponID   uint64     `gorm:"not null;index;comment:优惠券ID" json:"coupon_id"`
	CouponNo   string     `gorm:"type:varchar(64);not null;comment:优惠券编号" json:"coupon_no"`
	Status     string     `gorm:"type:varchar(32);default:'unused';comment:状态" json:"status"` // unused, used, expired
	UsedAt     *time.Time `gorm:"comment:使用时间" json:"used_at"`
	OrderID    string     `gorm:"type:varchar(64);comment:订单ID" json:"order_id"`
	ReceivedAt time.Time  `gorm:"comment:领取时间" json:"received_at"`
}

// CouponActivity 优惠券活动聚合根
type CouponActivity struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`
	Description string    `gorm:"type:text;comment:活动描述" json:"description"`
	StartTime   time.Time `gorm:"comment:开始时间" json:"start_time"`
	EndTime     time.Time `gorm:"comment:结束时间" json:"end_time"`
	CouponIDs   []uint64  `gorm:"type:json;serializer:json;comment:关联优惠券ID" json:"coupon_ids"`
	Status      string    `gorm:"type:varchar(32);default:'active';comment:状态" json:"status"` // active, inactive, ended
}

// NewCoupon 创建优惠券
func NewCoupon(name, description string, couponType CouponType, discountAmount, minOrderAmount int64) *Coupon {
	now := time.Now()
	validTo := now.AddDate(0, 3, 0) // 3个月后过期

	return &Coupon{
		CouponNo:       generateCouponNo(),
		Name:           name,
		Description:    description,
		Type:           couponType,
		Status:         CouponStatusDraft,
		DiscountAmount: discountAmount,
		MinOrderAmount: minOrderAmount,
		ValidFrom:      now,
		ValidTo:        validTo,
		UsageLimit:     10000, // 默认1万张
		UsagePerUser:   1,
		ApplicableIDs:  []uint64{},
	}
}

// Activate 激活优惠券
func (c *Coupon) Activate() error {
	if c.Status != CouponStatusDraft {
		return fmt.Errorf("only draft coupon can be activated")
	}
	c.Status = CouponStatusActive
	return nil
}

// Deactivate 停用优惠券
func (c *Coupon) Deactivate() error {
	if c.Status != CouponStatusActive {
		return fmt.Errorf("only active coupon can be deactivated")
	}
	c.Status = CouponStatusInactive
	return nil
}

// CheckAvailability 检查优惠券可用性
func (c *Coupon) CheckAvailability() error {
	if c.Status != CouponStatusActive {
		return fmt.Errorf("coupon is not active")
	}

	now := time.Now()
	if now.Before(c.ValidFrom) {
		return fmt.Errorf("coupon is not yet valid")
	}
	if now.After(c.ValidTo) {
		return fmt.Errorf("coupon has expired")
	}

	if c.TotalIssued >= c.UsageLimit {
		return fmt.Errorf("coupon usage limit reached")
	}

	return nil
}

// Issue 发放优惠券
func (c *Coupon) Issue(quantity int32) {
	c.TotalIssued += quantity
}

// Use 使用优惠券
func (c *Coupon) Use() {
	c.TotalUsed++
}

// NewUserCoupon 创建用户优惠券
func NewUserCoupon(userID, couponID uint64, couponNo string) *UserCoupon {
	return &UserCoupon{
		UserID:     userID,
		CouponID:   couponID,
		CouponNo:   couponNo,
		Status:     "unused",
		ReceivedAt: time.Now(),
	}
}

// Use 使用优惠券
func (u *UserCoupon) Use(orderID string) error {
	if u.Status != "unused" {
		return fmt.Errorf("coupon already used or expired")
	}
	u.Status = "used"
	now := time.Now()
	u.UsedAt = &now
	u.OrderID = orderID
	return nil
}

// NewCouponActivity 创建优惠券活动
func NewCouponActivity(name, description string, startTime, endTime time.Time, couponIDs []uint64) *CouponActivity {
	return &CouponActivity{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		CouponIDs:   couponIDs,
		Status:      "active",
	}
}

func generateCouponNo() string {
	return fmt.Sprintf("CPN%d", time.Now().UnixNano())
}
