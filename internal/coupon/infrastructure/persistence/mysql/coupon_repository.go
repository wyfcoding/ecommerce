package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"gorm.io/gorm"
)

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓储（MySQL）。
func NewCouponRepository(db *gorm.DB) domain.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *couponRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *couponRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *couponRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *couponRepository) SaveCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.saveCouponWithTx(ctx, r.db, coupon)
}

func (r *couponRepository) SaveCouponInTx(ctx context.Context, tx any, coupon *domain.Coupon) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveCouponWithTx(ctx, gormTx, coupon)
}

func (r *couponRepository) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	var model CouponModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCoupon(&model), nil
}

func (r *couponRepository) GetCouponByNo(ctx context.Context, couponNo string) (*domain.Coupon, error) {
	var model CouponModel
	if err := r.db.WithContext(ctx).Where("coupon_no = ?", couponNo).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCoupon(&model), nil
}

func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.saveCouponWithTx(ctx, r.db, coupon)
}

func (r *couponRepository) UpdateCouponInTx(ctx context.Context, tx any, coupon *domain.Coupon) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveCouponWithTx(ctx, gormTx, coupon)
}

func (r *couponRepository) DeleteCoupon(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&CouponModel{}, id).Error
}

func (r *couponRepository) ListCoupons(ctx context.Context, offset, limit int) ([]*domain.Coupon, int64, error) {
	var list []*CouponModel
	var total int64
	db := r.db.WithContext(ctx).Model(&CouponModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.Coupon, len(list))
	for i, model := range list {
		items[i] = toCoupon(model)
	}
	return items, total, nil
}

func (r *couponRepository) SaveUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.saveUserCouponWithTx(ctx, r.db, userCoupon)
}

func (r *couponRepository) SaveUserCouponInTx(ctx context.Context, tx any, userCoupon *domain.UserCoupon) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveUserCouponWithTx(ctx, gormTx, userCoupon)
}

func (r *couponRepository) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	var model UserCouponModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserCoupon(&model), nil
}

func (r *couponRepository) ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*domain.UserCoupon, int64, error) {
	var list []*UserCouponModel
	var total int64
	db := r.db.WithContext(ctx).Model(&UserCouponModel{}).Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.UserCoupon, len(list))
	for i, model := range list {
		items[i] = toUserCoupon(model)
	}
	return items, total, nil
}

func (r *couponRepository) UpdateUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.saveUserCouponWithTx(ctx, r.db, userCoupon)
}

func (r *couponRepository) UpdateUserCouponInTx(ctx context.Context, tx any, userCoupon *domain.UserCoupon) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveUserCouponWithTx(ctx, gormTx, userCoupon)
}

func (r *couponRepository) SaveActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.saveActivityWithTx(ctx, r.db, activity)
}

func (r *couponRepository) GetActivity(ctx context.Context, id uint64) (*domain.CouponActivity, error) {
	var model CouponActivityModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toActivity(&model), nil
}

func (r *couponRepository) UpdateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.saveActivityWithTx(ctx, r.db, activity)
}

func (r *couponRepository) ListActiveActivities(ctx context.Context, now time.Time) ([]*domain.CouponActivity, error) {
	var list []*CouponActivityModel
	if err := r.db.WithContext(ctx).Where("status = ? AND start_time <= ? AND end_time >= ?", "active", now.Unix(), now.Unix()).Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.CouponActivity, len(list))
	for i, model := range list {
		items[i] = toActivity(model)
	}
	return items, nil
}

func (r *couponRepository) saveCouponWithTx(ctx context.Context, tx *gorm.DB, coupon *domain.Coupon) error {
	if coupon == nil {
		return nil
	}
	model := toCouponModel(coupon)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	coupon.ID = uint64(model.ID)
	coupon.CreatedAt = model.CreatedAt
	coupon.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *couponRepository) saveUserCouponWithTx(ctx context.Context, tx *gorm.DB, userCoupon *domain.UserCoupon) error {
	if userCoupon == nil {
		return nil
	}
	model := toUserCouponModel(userCoupon)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	userCoupon.ID = uint64(model.ID)
	userCoupon.CreatedAt = model.CreatedAt
	userCoupon.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *couponRepository) saveActivityWithTx(ctx context.Context, tx *gorm.DB, activity *domain.CouponActivity) error {
	if activity == nil {
		return nil
	}
	model := toActivityModel(activity)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	activity.ID = uint64(model.ID)
	activity.CreatedAt = model.CreatedAt
	activity.UpdatedAt = model.UpdatedAt
	return nil
}
