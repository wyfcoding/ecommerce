// 变更说明：
// 促销 DB 仓储层，对接 MySQL，实现了 Command (写) 和 Query (读) 分离接口。
// 为了简化骨架直接合在一处，高并发情况下应该给 GORM 配置 ReadReplica 的 Resolver。
package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/promotion/domain"
	"gorm.io/gorm"
)

type PromotionRepositoryImpl struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) (domain.PromotionRepository, domain.PromotionReadRepository) {
	repo := &PromotionRepositoryImpl{db: db}
	return repo, repo
}

// ============== CQRS: COMMAND SIDE ============== //

func (r *PromotionRepositoryImpl) Save(ctx context.Context, promotion *domain.Promotion) error {
	// GORM 配置了 FullSaveAssociations，会自动级联保存 Rules 等关系表
	if promotion.ID == 0 {
		return r.db.WithContext(ctx).Create(promotion).Error
	}
	// 乐观锁版本叠加
	res := r.db.WithContext(ctx).Model(promotion).
		Where("id = ? AND version = ?", promotion.ID, promotion.Version).
		Updates(map[string]interface{}{
			"status":     promotion.Status,
			"name":       promotion.Name,
			"priority":   promotion.Priority,
			"used_count": promotion.UsedCount,
			"version":    gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrPromotionConflict
	}
	promotion.Version++
	return nil
}

func (r *PromotionRepositoryImpl) IncrementUsage(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).Model(&domain.Promotion{}).
		Where("id = ? AND (usage_limit = 0 OR used_count < usage_limit)", id).
		Update("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrPromotionUsageLimitExceeded
	}
	return nil
}

func (r *PromotionRepositoryImpl) GetByID(ctx context.Context, id uint64) (*domain.Promotion, error) {
	var p domain.Promotion
	if err := r.db.WithContext(ctx).Preload("Rules").First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPromotionNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PromotionRepositoryImpl) ListActive(ctx context.Context, now time.Time) ([]*domain.Promotion, error) {
	var list []*domain.Promotion
	err := r.db.WithContext(ctx).Preload("Rules").
		Where("status = 'ACTIVE' AND start_time <= ? AND end_time >= ?", now, now).
		Find(&list).Error
	return list, err
}

func (r *PromotionRepositoryImpl) ListByScope(ctx context.Context, scope domain.PromotionScope, scopeIDs []uint64, now time.Time) ([]*domain.Promotion, error) {
	var list []*domain.Promotion
	// MySQL 8.0 采用 JSON_CONTAINS 拦截 JSON 字段，或者规范里我们使用多对多关联表。
	// 这里保留查询骨架，示意作用
	err := r.db.WithContext(ctx).Preload("Rules").
		Where("status = 'ACTIVE' AND scope = ? AND start_time <= ? AND end_time >= ?", scope, now, now).
		Find(&list).Error
	return list, err
}

func (r *PromotionRepositoryImpl) GetUserUsageCount(ctx context.Context, promotionID, userID uint64) (int32, error) {
	// 具体应拆表 promotion_usage_records 做流水COUNT查询
	return 0, nil
}

// ============== CQRS: QUERY SIDE ============== //
func (r *PromotionRepositoryImpl) ListActiveByProduct(ctx context.Context, productID uint64, now time.Time) ([]*domain.Promotion, error) {
	// 查询全场通用，以及匹配指定商品范围的所有促销
	var list []*domain.Promotion
	err := r.db.WithContext(ctx).Preload("Rules").
		Where("status = 'ACTIVE' AND start_time <= ? AND end_time >= ?", now, now).
		Where("scope = 'ALL' OR (scope = 'PRODUCT' AND JSON_CONTAINS(scope_ids, ?))", fmt.Sprintf("[%d]", productID)).
		Find(&list).Error
	return list, err
}

func (r *PromotionRepositoryImpl) ListActiveByCategory(ctx context.Context, categoryID uint64, now time.Time) ([]*domain.Promotion, error) {
	return nil, nil // 按照同理查询
}

func (r *PromotionRepositoryImpl) ListActiveByMerchant(ctx context.Context, merchantID uint64, now time.Time) ([]*domain.Promotion, error) {
	return nil, nil // 按照同理查询
}
