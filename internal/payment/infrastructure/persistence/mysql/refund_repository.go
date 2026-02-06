package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/database/sharding"
	"gorm.io/gorm"
)

type refundRepository struct {
	sharding *sharding.Manager
	tx       *gorm.DB
}

// NewRefundRepository creates a new refundRepository instance.
func NewRefundRepository(sharding *sharding.Manager) domain.RefundRepository {
	return &refundRepository{sharding: sharding}
}

func (r *refundRepository) getDB(userID uint64) *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	return r.sharding.GetDB(userID)
}

func (r *refundRepository) Save(ctx context.Context, refund *domain.Refund) error {
	if refund == nil {
		return nil
	}
	db := r.getDB(uint64(refund.UserID))
	model := toRefundModel(refund)
	if model == nil {
		return nil
	}
	if err := db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toDomainRefund(model); synced != nil {
		*refund = *synced
	}
	return nil
}

func (r *refundRepository) FindByID(ctx context.Context, userID uint64, id uint64) (*domain.Refund, error) {
	db := r.getDB(userID)
	var refund RefundModel
	if err := db.WithContext(ctx).First(&refund, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainRefund(&refund), nil
}

func (r *refundRepository) FindByRefundNo(ctx context.Context, userID uint64, refundNo string) (*domain.Refund, error) {
	db := r.getDB(userID)
	var refund RefundModel
	if err := db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainRefund(&refund), nil
}

func (r *refundRepository) Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error {
	db := r.getDB(userID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *refundRepository) WithTx(tx any) domain.RefundRepository {
	if db, ok := tx.(*gorm.DB); ok {
		return &refundRepository{
			sharding: r.sharding,
			tx:       db,
		}
	}
	return r
}
