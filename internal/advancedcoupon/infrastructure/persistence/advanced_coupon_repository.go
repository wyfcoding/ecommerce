package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"

	"gorm.io/gorm"
)

type advancedCouponRepository struct {
	db *gorm.DB
}

// NewAdvancedCouponRepository 定义了数据持久层接口。
func NewAdvancedCouponRepository(db *gorm.DB) domain.AdvancedCouponRepository {
	return &advancedCouponRepository{db: db}
}

// 优惠券
func (r *advancedCouponRepository) Save(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *advancedCouponRepository) GetByID(ctx context.Context, id uint64) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *advancedCouponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *advancedCouponRepository) List(ctx context.Context, status domain.CouponStatus, offset, limit int) ([]*domain.Coupon, int64, error) {
	var list []*domain.Coupon
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Coupon{})
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 使用记录
func (r *advancedCouponRepository) SaveUsage(ctx context.Context, usage *domain.CouponUsage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

func (r *advancedCouponRepository) CountUsageByUser(ctx context.Context, userID, couponID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.CouponUsage{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}

// 统计
func (r *advancedCouponRepository) SaveStatistics(ctx context.Context, stats *domain.CouponStatistics) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

func (r *advancedCouponRepository) GetStatistics(ctx context.Context, couponID uint64) (*domain.CouponStatistics, error) {
	var stats domain.CouponStatistics
	if err := r.db.WithContext(ctx).Where("coupon_id = ?", couponID).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}
