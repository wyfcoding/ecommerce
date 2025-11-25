package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) domain.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Save(ctx context.Context, entity *domain.Payment) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint64) (*domain.Payment, error) {
	var entity domain.Payment
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) Update(ctx context.Context, entity *domain.Payment) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *paymentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Payment{}, id).Error
}

func (r *paymentRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var entity domain.Payment
	if err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID uint64) (*domain.Payment, error) {
	var entity domain.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Payment, int64, error) {
	var entities []*domain.Payment
	var total int64

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *paymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *paymentRepository) FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*domain.PaymentLog, error) {
	var logs []*domain.PaymentLog
	if err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
