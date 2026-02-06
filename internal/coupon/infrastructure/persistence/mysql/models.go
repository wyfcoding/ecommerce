package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"gorm.io/gorm"
)

// Uint64Array JSON 序列化的 uint64 数组。
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

// StringArray JSON 序列化的 string 数组。
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

// CouponModel 优惠券写模型（持久化专用）。
type CouponModel struct {
	gorm.Model
	CouponNo              string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:优惠券编号"`
	Name                  string              `gorm:"type:varchar(255);not null;comment:名称"`
	Description           string              `gorm:"type:text;comment:描述"`
	Type                  domain.CouponType   `gorm:"default:1;comment:类型"`
	Status                domain.CouponStatus `gorm:"default:1;comment:状态"`
	DiscountAmount        int64               `gorm:"comment:折扣金额/比例"`
	MinOrderAmount        int64               `gorm:"comment:最低订单金额"`
	MaxDiscount           int64               `gorm:"comment:最大折扣金额"`
	ValidFrom             time.Time           `gorm:"comment:有效期开始"`
	ValidTo               time.Time           `gorm:"comment:有效期结束"`
	UsageLimit            int32               `gorm:"default:0;comment:总发行量"`
	UsagePerUser          int32               `gorm:"default:1;comment:每人限领"`
	TotalIssued           int32               `gorm:"default:0;comment:已发行量"`
	TotalUsed             int32               `gorm:"default:0;comment:已使用量"`
	ConditionExpr         string              `gorm:"type:text;comment:判定表达式"`
	ApplicableScope       string              `gorm:"type:varchar(255);comment:适用范围"`
	ApplicableProductIDs  Uint64Array         `gorm:"type:json;comment:适用商品ID列表"`
	ExcludedProductIDs    Uint64Array         `gorm:"type:json;comment:排除商品ID列表"`
	ApplicableCategoryIDs StringArray         `gorm:"type:json;comment:适用分类ID列表"`
	ExcludedCategoryIDs   StringArray         `gorm:"type:json;comment:排除分类ID列表"`
	UserTierRequirement   string              `gorm:"type:varchar(32);comment:会员等级要求"`
	CanStack              bool                `gorm:"default:false;comment:是否可叠加"`
	StackingRules         string              `gorm:"type:text;comment:叠加规则(JSON)"`
}

// UserCouponModel 用户优惠券写模型。
type UserCouponModel struct {
	gorm.Model
	UserID     uint64     `gorm:"not null;index;comment:用户ID"`
	CouponID   uint64     `gorm:"not null;index;comment:优惠券ID"`
	CouponNo   string     `gorm:"type:varchar(64);not null;comment:优惠券编号"`
	Status     string     `gorm:"type:varchar(32);default:'unused';comment:状态"`
	UsedAt     *time.Time `gorm:"comment:使用时间"`
	OrderID    string     `gorm:"type:varchar(64);comment:订单ID"`
	ReceivedAt time.Time  `gorm:"comment:领取时间"`
}

// CouponActivityModel 优惠券活动写模型。
type CouponActivityModel struct {
	gorm.Model
	Name        string      `gorm:"type:varchar(255);not null;comment:活动名称"`
	Description string      `gorm:"type:text;comment:活动描述"`
	StartTime   time.Time   `gorm:"comment:开始时间"`
	EndTime     time.Time   `gorm:"comment:结束时间"`
	CouponIDs   Uint64Array `gorm:"type:json;comment:关联优惠券ID"`
	Status      string      `gorm:"type:varchar(32);default:'active';comment:状态"`
}

func (CouponModel) TableName() string {
	return "coupons"
}

func (UserCouponModel) TableName() string {
	return "user_coupons"
}

func (CouponActivityModel) TableName() string {
	return "coupon_activities"
}

func toCouponModel(coupon *domain.Coupon) *CouponModel {
	if coupon == nil {
		return nil
	}
	model := &CouponModel{
		Model: gorm.Model{
			ID:        uint(coupon.ID),
			CreatedAt: coupon.CreatedAt,
			UpdatedAt: coupon.UpdatedAt,
		},
		CouponNo:              coupon.CouponNo,
		Name:                  coupon.Name,
		Description:           coupon.Description,
		Type:                  coupon.Type,
		Status:                coupon.Status,
		DiscountAmount:        coupon.DiscountAmount,
		MinOrderAmount:        coupon.MinOrderAmount,
		MaxDiscount:           coupon.MaxDiscount,
		ValidFrom:             coupon.ValidFrom,
		ValidTo:               coupon.ValidTo,
		UsageLimit:            coupon.UsageLimit,
		UsagePerUser:          coupon.UsagePerUser,
		TotalIssued:           coupon.TotalIssued,
		TotalUsed:             coupon.TotalUsed,
		ConditionExpr:         coupon.ConditionExpr,
		ApplicableScope:       coupon.ApplicableScope,
		ApplicableProductIDs:  Uint64Array(coupon.ApplicableProductIDs),
		ExcludedProductIDs:    Uint64Array(coupon.ExcludedProductIDs),
		ApplicableCategoryIDs: StringArray(coupon.ApplicableCategoryIDs),
		ExcludedCategoryIDs:   StringArray(coupon.ExcludedCategoryIDs),
		UserTierRequirement:   coupon.UserTierRequirement,
		CanStack:              coupon.CanStack,
		StackingRules:         coupon.StackingRules,
	}
	return model
}

func toCoupon(model *CouponModel) *domain.Coupon {
	if model == nil {
		return nil
	}
	coupon := &domain.Coupon{
		ID:                    uint64(model.ID),
		CreatedAt:             model.CreatedAt,
		UpdatedAt:             model.UpdatedAt,
		CouponNo:              model.CouponNo,
		Name:                  model.Name,
		Description:           model.Description,
		Type:                  model.Type,
		Status:                model.Status,
		DiscountAmount:        model.DiscountAmount,
		MinOrderAmount:        model.MinOrderAmount,
		MaxDiscount:           model.MaxDiscount,
		ValidFrom:             model.ValidFrom,
		ValidTo:               model.ValidTo,
		UsageLimit:            model.UsageLimit,
		UsagePerUser:          model.UsagePerUser,
		TotalIssued:           model.TotalIssued,
		TotalUsed:             model.TotalUsed,
		ConditionExpr:         model.ConditionExpr,
		ApplicableScope:       model.ApplicableScope,
		ApplicableProductIDs:  append([]uint64{}, model.ApplicableProductIDs...),
		ExcludedProductIDs:    append([]uint64{}, model.ExcludedProductIDs...),
		ApplicableCategoryIDs: append([]string{}, model.ApplicableCategoryIDs...),
		ExcludedCategoryIDs:   append([]string{}, model.ExcludedCategoryIDs...),
		UserTierRequirement:   model.UserTierRequirement,
		CanStack:              model.CanStack,
		StackingRules:         model.StackingRules,
	}
	return coupon
}

func toUserCouponModel(coupon *domain.UserCoupon) *UserCouponModel {
	if coupon == nil {
		return nil
	}
	model := &UserCouponModel{
		Model: gorm.Model{
			ID:        uint(coupon.ID),
			CreatedAt: coupon.CreatedAt,
			UpdatedAt: coupon.UpdatedAt,
		},
		UserID:     coupon.UserID,
		CouponID:   coupon.CouponID,
		CouponNo:   coupon.CouponNo,
		Status:     coupon.Status,
		UsedAt:     coupon.UsedAt,
		OrderID:    coupon.OrderID,
		ReceivedAt: coupon.ReceivedAt,
	}
	return model
}

func toUserCoupon(model *UserCouponModel) *domain.UserCoupon {
	if model == nil {
		return nil
	}
	userCoupon := &domain.UserCoupon{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		UserID:     model.UserID,
		CouponID:   model.CouponID,
		CouponNo:   model.CouponNo,
		Status:     model.Status,
		UsedAt:     model.UsedAt,
		OrderID:    model.OrderID,
		ReceivedAt: model.ReceivedAt,
	}
	return userCoupon
}

func toActivityModel(activity *domain.CouponActivity) *CouponActivityModel {
	if activity == nil {
		return nil
	}
	model := &CouponActivityModel{
		Model: gorm.Model{
			ID:        uint(activity.ID),
			CreatedAt: activity.CreatedAt,
			UpdatedAt: activity.UpdatedAt,
		},
		Name:        activity.Name,
		Description: activity.Description,
		StartTime:   activity.StartTime,
		EndTime:     activity.EndTime,
		CouponIDs:   Uint64Array(activity.CouponIDs),
		Status:      activity.Status,
	}
	return model
}

func toActivity(model *CouponActivityModel) *domain.CouponActivity {
	if model == nil {
		return nil
	}
	return &domain.CouponActivity{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Description: model.Description,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		CouponIDs:   append([]uint64{}, model.CouponIDs...),
		Status:      model.Status,
	}
}
