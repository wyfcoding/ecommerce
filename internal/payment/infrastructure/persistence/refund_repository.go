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
	tx       *gorm.DB
}

// NewRefundRepository creates a new refundRepository instance.
func NewRefundRepository(sharding *sharding.Manager) domain.RefundRepository {
	return &refundRepository{sharding: sharding}
}

func (r *refundRepository) Save(ctx context.Context, refund *domain.Refund) error {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0)
	}
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
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0)
	}
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
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0)
	}
	return db.WithContext(ctx).Delete(&domain.Refund{}, id).Error
}

func (r *refundRepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *refundRepository) WithTx(tx any) domain.RefundRepository {
	return &refundRepository{
		sharding: r.sharding,
		tx:       tx.(*gorm.DB),
	}
}
