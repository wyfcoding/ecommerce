package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/databases/sharding"
	"gorm.io/gorm"
)

type refundRepository struct {
	sharding *sharding.Manager
}

// NewRefundRepository creates a new refundRepository instance.
func NewRefundRepository(sharding *sharding.Manager) domain.RefundRepository {
	return &refundRepository{sharding: sharding}
}

func (r *refundRepository) Save(ctx context.Context, refund *domain.Refund) error {
	// TODO: Refund 的分片逻辑。目前使用分片 0。
	// 理想情况下，Refund 应按 UserID 或 OrderID 分片，类似于 Payment。
	// 假设 Refund 有 PaymentID，如果不存在，我们可能需要查找 Payment 以获取 UserID。
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Create(refund).Error
}

func (r *refundRepository) FindByID(ctx context.Context, id uint64) (*domain.Refund, error) {
	db := r.sharding.GetDB(0)
	var refund domain.Refund
	if err := db.WithContext(ctx).First(&refund, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) FindByRefundNo(ctx context.Context, refundNo string) (*domain.Refund, error) {
	db := r.sharding.GetDB(0)
	var refund domain.Refund
	if err := db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) Update(ctx context.Context, refund *domain.Refund) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Save(refund).Error
}

func (r *refundRepository) Delete(ctx context.Context, id uint64) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Delete(&domain.Refund{}, id).Error
}
