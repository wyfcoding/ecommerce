package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepository) FindByID(ctx context.Context, id uint64) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.WithContext(ctx).Preload("Logs").Preload("Refunds").First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.WithContext(ctx).Preload("Logs").Preload("Refunds").Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(payment).Error; err != nil {
			return err
		}

		// Save new logs
		for _, log := range payment.Logs {
			if log.ID == 0 {
				log.PaymentID = payment.ID
				if err := tx.Create(log).Error; err != nil {
					return err
				}
			}
		}

		// Save new refunds or update existing
		for _, refund := range payment.Refunds {
			refund.PaymentID = payment.ID
			if refund.ID == 0 {
				if err := tx.Create(refund).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Save(refund).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID uint64) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.WithContext(ctx).Preload("Logs").Preload("Refunds").Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Payment{}, id).Error
}

func (r *PaymentRepository) ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Payment, int64, error) {
	var payments []*domain.Payment
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Payment{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Logs").Preload("Refunds").Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (r *PaymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *PaymentRepository) FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*domain.PaymentLog, error) {
	var logs []*domain.PaymentLog
	if err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).Order("created_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
