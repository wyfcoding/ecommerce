package data

import (
	"context"
	"errors"
	"time"

	"ecommerce/ecommerce/app/marketing/internal/biz"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type couponRepo struct {
	db *gorm.DB
}

func NewCouponRepo(db *gorm.DB) biz.CouponRepo {
	return &couponRepo{db: db}
}

func (r *couponRepo) CreateTemplate(ctx context.Context, template *biz.CouponTemplate) (*biz.CouponTemplate, error) {
	// 将 biz.CouponTemplate 转换为 data.CouponTemplate
	dataTemplate := &CouponTemplate{
		Title:               template.Title,
		Type:                template.Type,
		ScopeType:           template.ScopeType,
		ScopeIDs:            template.ScopeIDs,
		Rules:               JSONRuleSet(template.Rules),
		TotalQuantity:       template.TotalQuantity,
		IssuedQuantity:      0, // 新建时已领取为0
		PerUserLimit:        template.PerUserLimit,
		ValidityType:        template.ValidityType,
		ValidFrom:           template.ValidFrom,
		ValidTo:             template.ValidTo,
		ValidDaysAfterClaim: template.ValidDaysAfterClaim,
		Status:              1, // 默认为可用
	}

	if err := r.db.WithContext(ctx).Create(dataTemplate).Error; err != nil {
		return nil, err
	}

	template.ID = dataTemplate.ID // 将数据库生成的 ID 回写
	return template, nil
}

func (r *couponRepo) ClaimCoupon(ctx context.Context, userID, templateID uint64) (*biz.UserCoupon, error) {
	var userCoupon *biz.UserCoupon

	// 启动事务
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 锁定模板行，防止并发修改库存
		var template CouponTemplate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&template, templateID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("coupon template not found")
			}
			return err
		}

		// 2. 在事务内进行所有校验
		// 2a. 校验模板状态
		if template.Status != 1 {
			return errors.New("coupon is not available")
		}
		// 2b. 校验总库存
		if template.TotalQuantity > 0 && template.IssuedQuantity >= template.TotalQuantity {
			return errors.New("coupon has been fully claimed")
		}
		// 2c. 校验个人限领
		var count int64
		tx.Model(&UserCoupon{}).Where("user_id = ? AND template_id = ?", userID, templateID).Count(&count)
		if count >= int64(template.PerUserLimit) {
			return errors.New("you have reached the claim limit for this coupon")
		}

		// 3. 所有校验通过，创建用户优惠券
		now := time.Now()
		var validFrom, validTo time.Time
		if template.ValidityType == 1 { // 固定时间
			validFrom = *template.ValidFrom
			validTo = *template.ValidTo
		} else { // 领取后有效
			validFrom = now
			validTo = now.AddDate(0, 0, int(template.ValidDaysAfterClaim))
		}

		newUserCoupon := UserCoupon{
			TemplateID: templateID,
			UserID:     userID,
			CouponCode: uuid.New().String(), // 生成唯一券码
			Status:     1,                   // 未使用
			ClaimedAt:  now,
			ValidFrom:  validFrom,
			ValidTo:    validTo,
		}
		if err := tx.Create(&newUserCoupon).Error; err != nil {
			return err
		}

		// 4. 更新模板的已领取数量
		result := tx.Model(&CouponTemplate{}).Where("id = ?", templateID).Update("issued_quantity", gorm.Expr("issued_quantity + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// 如果没有行被更新，说明模板可能被删了，回滚
			return errors.New("failed to update coupon template quantity")
		}

		// 5. 转换模型用于返回
		userCoupon = &biz.UserCoupon{
			ID:         newUserCoupon.ID,
			TemplateID: newUserCoupon.TemplateID,
			UserID:     newUserCoupon.UserID,
			CouponCode: newUserCoupon.CouponCode,
			Status:     newUserCoupon.Status,
			ClaimedAt:  newUserCoupon.ClaimedAt,
			ValidFrom:  newUserCoupon.ValidFrom,
			ValidTo:    newUserCoupon.ValidTo,
		}

		return nil // 提交事务
	})

	return userCoupon, err
}

func (r *couponRepo) GetUserCouponByCode(ctx context.Context, userID uint64, code string) (*biz.UserCoupon, error) {
	var userCoupon UserCoupon
	err := r.db.WithContext(ctx).Where("user_id = ? AND coupon_code = ?", userID, code).First(&userCoupon).Error
	if err != nil {
		return nil, err
	}
	// ... 模型转换: data.UserCoupon -> biz.UserCoupon ...
	return &bizUserCoupon, nil
}

func (r *couponRepo) GetTemplateByID(ctx context.Context, templateID uint64) (*biz.CouponTemplate, error) {
	var template CouponTemplate
	err := r.db.WithContext(ctx).First(&template, templateID).Error
	if err != nil {
		return nil, err
	}
	// ... 模型转换: data.CouponTemplate -> biz.CouponTemplate ...
	return &bizTemplate, nil
}
