package data

import (
	"context"
	"ecommerce/internal/payment/biz"
	"ecommerce/internal/payment/data/model"
	"time"

	"gorm.io/gorm"
)

type paymentRepo struct {
	data *Data
}

// NewPaymentRepo creates a new PaymentRepo.
func NewPaymentRepo(data *Data) biz.PaymentRepo {
	return &paymentRepo{data: data}
}

// CreatePaymentTransaction creates a new payment transaction record.
func (r *paymentRepo) CreatePaymentTransaction(ctx context.Context, tx *biz.PaymentTransaction) (*biz.PaymentTransaction, error) {
	po := &model.PaymentTransaction{
		PaymentID:     tx.PaymentID,
		OrderID:       tx.OrderID,
		UserID:        tx.UserID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		PaymentMethod: tx.PaymentMethod,
		Status:        tx.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	tx.ID = po.ID
	return tx, nil
}

// GetPaymentTransactionByPaymentID retrieves a payment transaction by its payment ID.
func (r *paymentRepo) GetPaymentTransactionByPaymentID(ctx context.Context, paymentID string) (*biz.PaymentTransaction, error) {
	var po model.PaymentTransaction
	if err := r.data.db.WithContext(ctx).Where("payment_id = ?", paymentID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Transaction not found
		}
		return nil, err
	}
	return &biz.PaymentTransaction{
		ID:            po.ID,
		PaymentID:     po.PaymentID,
		OrderID:       po.OrderID,
		UserID:        po.UserID,
		Amount:        po.Amount,
		Currency:      po.Currency,
		PaymentMethod: po.PaymentMethod,
		Status:        po.Status,
		TransactionNo: po.TransactionNo,
		CallbackData:  po.CallbackData,
		PaidAt:        po.PaidAt,
	}, nil
}

// UpdatePaymentTransactionStatus updates the status of a payment transaction.
func (r *paymentRepo) UpdatePaymentTransactionStatus(ctx context.Context, paymentID string, newStatus string, transactionNo string, callbackData string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status":         newStatus,
		"transaction_no": transactionNo,
		"callback_data":  callbackData,
		"paid_at":        paidAt,
	}
	return r.data.db.WithContext(ctx).Model(&model.PaymentTransaction{}).Where("payment_id = ?", paymentID).Updates(updates).Error
}
