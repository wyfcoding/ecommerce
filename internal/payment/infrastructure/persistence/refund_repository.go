package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"
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
	// TODO: Sharding logic for Refund. Currently using shard 0.
	// Ideally, Refund should be sharded by UserID or OrderID, similar to Payment.
	// Assuming Refund has a PaymentID, we might need to look up Payment to get UserID if not present.
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
