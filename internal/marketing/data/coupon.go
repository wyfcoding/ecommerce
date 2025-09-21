package data

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/marketing/biz"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type couponRepo struct {
	*Data
}

// NewCouponRepo 是 couponRepo 的构造函数。
func NewCouponRepo(data *Data) biz.CouponRepo {
	return &couponRepo{Data: data}
}

// toBizCouponTemplate 将 data.CouponTemplate 转换为 biz.CouponTemplate。
func (r *couponRepo) toBizCouponTemplate(po *CouponTemplate) *biz.CouponTemplate {
	if po == nil {
		return nil
	}
	return &biz.CouponTemplate{
		ID:                  po.ID,
		Title:               po.Title,
		Type:                po.Type,
		ScopeType:           po.ScopeType,
		ScopeIDs:            po.ScopeIDs,
		Rules:               biz.RuleSet(po.Rules),
		TotalQuantity:       po.TotalQuantity,
		IssuedQuantity:      po.IssuedQuantity,
		PerUserLimit:        po.PerUserLimit,
		ValidityType:        po.ValidityType,
		ValidFrom:           po.ValidFrom,
		ValidTo:             po.ValidTo,
		ValidDaysAfterClaim: po.ValidDaysAfterClaim,
		Status:              po.Status,
	}
}

// toBizUserCoupon 将 data.UserCoupon 转换为 biz.UserCoupon。
func (r *couponRepo) toBizUserCoupon(po *UserCoupon) *biz.UserCoupon {
	if po == nil {
		return nil
	}
	return &biz.UserCoupon{
		ID:         po.ID,
		TemplateID: po.TemplateID,
		UserID:     po.UserID,
		CouponCode: po.CouponCode,
		Status:     po.Status,
		ClaimedAt:  po.ClaimedAt,
		ValidFrom:  po.ValidFrom,
		ValidTo:    po.ValidTo,
	}
}

// CreateTemplate 创建一个新的优惠券模板。
func (r *couponRepo) CreateTemplate(ctx context.Context, template *biz.CouponTemplate) (*biz.CouponTemplate, error) {
	po := &CouponTemplate{
		Title:               template.Title,
		Type:                template.Type,
		ScopeType:           template.ScopeType,
		ScopeIDs:            JSONUint64Array(template.ScopeIDs),
		Rules:               JSONRuleSet(template.Rules),
		TotalQuantity:       template.TotalQuantity,
		IssuedQuantity:      0,
		PerUserLimit:        template.PerUserLimit,
		ValidityType:        template.ValidityType,
		ValidFrom:           template.ValidFrom,
		ValidTo:             template.ValidTo,
		ValidDaysAfterClaim: template.ValidDaysAfterClaim,
		Status:              1,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return r.toBizCouponTemplate(po), nil
}

// ClaimCoupon 领取优惠券。
func (r *couponRepo) ClaimCoupon(ctx context.Context, userID, templateID uint64) (*biz.UserCoupon, error) {
	var userCoupon *biz.UserCoupon

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var template CouponTemplate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&template, templateID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("coupon template not found")
			}
			return err
		}

		if template.Status != 1 {
			return errors.New("coupon is not available")
		}
		if template.TotalQuantity > 0 && template.IssuedQuantity >= template.TotalQuantity {
			return errors.New("coupon has been fully claimed")
		}
		var count int64
		tx.Model(&UserCoupon{}).Where("user_id = ? AND template_id = ?", userID, templateID).Count(&count)
		if count >= int64(template.PerUserLimit) {
			return errors.New("you have reached the claim limit for this coupon")
		}

		now := time.Now()
		var validFrom, validTo time.Time
		if template.ValidityType == 1 {
			validFrom = *template.ValidFrom
			validTo = *template.ValidTo
		} else {
			validFrom = now
			validTo = now.AddDate(0, 0, int(template.ValidDaysAfterClaim))
		}

		newUserCoupon := UserCoupon{
			TemplateID: templateID,
			UserID:     userID,
			CouponCode: uuid.New().String(),
			Status:     1,
			ClaimedAt:  now,
			ValidFrom:  validFrom,
			ValidTo:    validTo,
		}
		if err := tx.Create(&newUserCoupon).Error; err != nil {
			return err
		}

		result := tx.Model(&CouponTemplate{}).Where("id = ?", templateID).Update("issued_quantity", gorm.Expr("issued_quantity + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("failed to update coupon template quantity")
		}

		userCoupon = r.toBizUserCoupon(&newUserCoupon)
		return nil
	})

	return userCoupon, err
}

// GetUserCouponByCode 根据券码获取用户优惠券。
func (r *couponRepo) GetUserCouponByCode(ctx context.Context, userID uint64, code string) (*biz.UserCoupon, error) {
	var userCoupon UserCoupon
	err := r.db.WithContext(ctx).Where("user_id = ? AND coupon_code = ?", userID, code).First(&userCoupon).Error
	if err != nil {
		return nil, err
	}
	return r.toBizUserCoupon(&userCoupon), nil
}

// GetTemplateByID 根据ID获取优惠券模板。
func (r *couponRepo) GetTemplateByID(ctx context.Context, templateID uint64) (*biz.CouponTemplate, error) {
	var template CouponTemplate
	err := r.db.WithContext(ctx).First(&template, templateID).Error;
	if err != nil {
		return nil, err
	}
	return r.toBizCouponTemplate(&template), nil
}

// ListUserCoupons 获取用户优惠券列表。
func (r *couponRepo) ListUserCoupons(ctx context.Context, userID uint64, status int8) ([]*biz.UserCoupon, error) {
	var userCoupons []UserCoupon
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != 0 { // 0表示所有状态
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	
bizUserCoupons := make([]*biz.UserCoupon, len(userCoupons))
	for i, uc := range userCoupons {
		bizUserCoupons[i] = r.toBizUserCoupon(&uc)
	}
	return bizUserCoupons, nil
}

// UpdateUserCouponStatus 更新用户优惠券状态。
func (r *couponRepo) UpdateUserCouponStatus(ctx context.Context, userCouponID uint64, newStatus int8, orderID *uint64) error {
	updates := map[string]interface{}{
		"status": newStatus,
		"used_at": time.Now(),
	}
	if orderID != nil {
		updates["order_id"] = *orderID
	}
	return r.db.WithContext(ctx).Model(&UserCoupon{}).Where("id = ?", userCouponID).Updates(updates).Error
}
