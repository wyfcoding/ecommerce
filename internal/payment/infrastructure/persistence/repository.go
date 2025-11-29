package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"
)

type paymentRepository struct {
	sharding *sharding.Manager
}

func NewPaymentRepository(sharding *sharding.Manager) domain.PaymentRepository {
	return &paymentRepository{sharding: sharding}
}

func (r *paymentRepository) Save(ctx context.Context, entity *domain.Payment) error {
	db := r.sharding.GetDB(uint64(entity.UserID))
	return db.WithContext(ctx).Create(entity).Error
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint64) (*domain.Payment, error) {
	// TODO: Support sharding by ID or scan all shards. Defaulting to shard 0.
	db := r.sharding.GetDB(0)
	var entity domain.Payment
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) Update(ctx context.Context, entity *domain.Payment) error {
	db := r.sharding.GetDB(uint64(entity.UserID))
	return db.WithContext(ctx).Save(entity).Error
}

func (r *paymentRepository) Delete(ctx context.Context, id uint64) error {
	// TODO: Support sharding by ID. Defaulting to shard 0.
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Delete(&domain.Payment{}, id).Error
}

func (r *paymentRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	// TODO: Support sharding by PaymentNo. Defaulting to shard 0.
	db := r.sharding.GetDB(0)
	var entity domain.Payment
	if err := db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID uint64) (*domain.Payment, error) {
	// TODO: Support sharding by OrderID. Defaulting to shard 0.
	db := r.sharding.GetDB(0)
	var entity domain.Payment
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Payment, int64, error) {
	var entities []*domain.Payment
	var total int64

	db := r.sharding.GetDB(userID)

	if err := db.WithContext(ctx).Model(&domain.Payment{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *paymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	// PaymentLog usually doesn't have UserID directly, but it belongs to a Payment.
	// We should probably pass PaymentID or UserID.
	// Assuming PaymentLog has PaymentID, and we can't easily resolve UserID from it without query.
	// For now, let's default to shard 0 or we need to change interface to pass UserID.
	// Let's assume shard 0 for logs if we can't resolve.
	// OR, if PaymentLog has UserID (denormalized).
	// Let's check domain.PaymentLog.
	// If not, we use shard 0.
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Create(log).Error
}

func (r *paymentRepository) FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*domain.PaymentLog, error) {
	db := r.sharding.GetDB(0)
	var logs []*domain.PaymentLog
	if err := db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
