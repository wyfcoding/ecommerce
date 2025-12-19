package persistence

import (
	"context"
	"time" // 导入时间包，用于查询条件。

	"github.com/wyfcoding/ecommerce/internal/coupon/domain" // 导入优惠券模块的领域层。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// couponRepository 是 CouponRepository 接口的GORM实现。
// 它负责将优惠券模块的领域实体映射到数据库，并执行持久化操作。
type couponRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewCouponRepository 创建并返回一个新的 couponRepository 实例。
// db: GORM数据库连接实例。
func NewCouponRepository(db *gorm.DB) domain.CouponRepository {
	return &couponRepository{db: db}
}

// --- Coupon methods ---

// SaveCoupon 将优惠券实体保存到数据库。
// 如果优惠券已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *couponRepository) SaveCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

// GetCoupon 根据ID从数据库获取优惠券记录。
func (r *couponRepository) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetCouponByNo 根据优惠券编号从数据库获取优惠券记录。
func (r *couponRepository) GetCouponByNo(ctx context.Context, couponNo string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).Where("coupon_no = ?", couponNo).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// UpdateCoupon 更新数据库中的优惠券记录。
func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

// DeleteCoupon 根据ID从数据库删除优惠券记录。
func (r *couponRepository) DeleteCoupon(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Coupon{}, id).Error
}

// ListCoupons 从数据库列出优惠券记录，支持通过状态过滤和分页。
func (r *couponRepository) ListCoupons(ctx context.Context, status domain.CouponStatus, offset, limit int) ([]*domain.Coupon, int64, error) {
	var list []*domain.Coupon
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Coupon{})
	if status != 0 { // 如果状态不为0，则按状态过滤。
		db = db.Where("status = ?", status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页 and 排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- UserCoupon methods ---

// SaveUserCoupon 将用户优惠券实体保存到数据库。
func (r *couponRepository) SaveUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

// GetUserCoupon 根据ID从数据库获取用户优惠券记录。
func (r *couponRepository) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	var userCoupon domain.UserCoupon
	if err := r.db.WithContext(ctx).First(&userCoupon, id).Error; err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

// ListUserCoupons 从数据库列出指定用户ID的优惠券记录，支持通过状态过滤和分页。
func (r *couponRepository) ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*domain.UserCoupon, int64, error) {
	var list []*domain.UserCoupon
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.UserCoupon{}).Where("user_id = ?", userID)
	if status != "" { // 如果状态不为空，则按状态过滤。
		db = db.Where("status = ?", status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页 and 排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateUserCoupon 更新数据库中的用户优惠券记录。
func (r *couponRepository) UpdateUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

// --- CouponActivity methods ---

// SaveActivity 将优惠券活动实体保存到数据库。
func (r *couponRepository) SaveActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

// GetActivity 根据ID从数据库获取优惠券活动记录。
func (r *couponRepository) GetActivity(ctx context.Context, id uint64) (*domain.CouponActivity, error) {
	var activity domain.CouponActivity
	if err := r.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// UpdateActivity 更新数据库中的优惠券活动记录。
func (r *couponRepository) UpdateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

// ListActiveActivities 列出所有当前时间正在进行中的优惠券活动记录。
func (r *couponRepository) ListActiveActivities(ctx context.Context, now time.Time) ([]*domain.CouponActivity, error) {
	var list []*domain.CouponActivity
	// 查询状态为“active”且在有效期内的活动。
	if err := r.db.WithContext(ctx).Where("status = ? AND start_time <= ? AND end_time >= ?", "active", now, now).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
