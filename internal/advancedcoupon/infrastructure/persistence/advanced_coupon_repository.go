package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"

	"gorm.io/gorm"
)

type contextKey struct{}

var txKey = contextKey{}

type advancedCouponRepository struct {
	db *gorm.DB
}

// NewAdvancedCouponRepository 定义了数据持久层接口。
func NewAdvancedCouponRepository(db *gorm.DB) domain.AdvancedCouponRepository {
	return &advancedCouponRepository{db: db}
}

func (r *advancedCouponRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

// 优惠券
func (r *advancedCouponRepository) Save(ctx context.Context, coupon *domain.Coupon) error {
	return r.getDB(ctx).Save(coupon).Error
}

func (r *advancedCouponRepository) GetByID(ctx context.Context, id uint64) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.getDB(ctx).First(&coupon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *advancedCouponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.getDB(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
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

	db := r.getDB(ctx).Model(&domain.Coupon{})
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

func (r *advancedCouponRepository) IncrementUsage(ctx context.Context, couponID uint64) error {
	result := r.getDB(ctx).Model(&domain.Coupon{}).
		Where("id = ? AND (total_quantity = 0 OR used_quantity < total_quantity)", couponID).
		UpdateColumn("used_quantity", gorm.Expr("used_quantity + ?", 1))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("coupon exhausted or not found")
	}
	return nil
}

// 使用记录
func (r *advancedCouponRepository) SaveUsage(ctx context.Context, usage *domain.CouponUsage) error {
	return r.getDB(ctx).Save(usage).Error
}

func (r *advancedCouponRepository) CountUsageByUser(ctx context.Context, userID, couponID uint64) (int64, error) {
	var count int64
	err := r.getDB(ctx).Model(&domain.CouponUsage{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}

// 统计
func (r *advancedCouponRepository) SaveStatistics(ctx context.Context, stats *domain.CouponStatistics) error {
	return r.getDB(ctx).Save(stats).Error
}

func (r *advancedCouponRepository) GetStatistics(ctx context.Context, couponID uint64) (*domain.CouponStatistics, error) {
	var stats domain.CouponStatistics
	if err := r.getDB(ctx).Where("coupon_id = ?", couponID).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

func (r *advancedCouponRepository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
