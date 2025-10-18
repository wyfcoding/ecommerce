package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/payment/model"
)

// PaymentRepository 定义了支付数据仓库的接口
type PaymentRepository interface {
	CreatePaymentTransaction(ctx context.Context, tx *model.PaymentTransaction) error
	GetPaymentByTransactionSN(ctx context.Context, sn string) (*model.PaymentTransaction, error)
	GetPaymentByOrderSN(ctx context.Context, orderSN string) (*model.PaymentTransaction, error)
	UpdatePaymentTransaction(ctx context.Context, tx *model.PaymentTransaction) error

	CreateRefundTransaction(ctx context.Context, tx *model.RefundTransaction) error
	GetRefundByRefundSN(ctx context.Context, sn string) (*model.RefundTransaction, error)
	UpdateRefundTransaction(ctx context.Context, tx *model.RefundTransaction) error
}

// paymentRepository 是接口的具体实现
type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建一个新的 paymentRepository 实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// CreatePaymentTransaction 在数据库中创建一条新的支付流水记录
func (r *paymentRepository) CreatePaymentTransaction(ctx context.Context, tx *model.PaymentTransaction) error {
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		return fmt.Errorf("数据库创建支付流水失败: %w", err)
	}
	return nil
}

// GetPaymentByTransactionSN 根据系统内部流水号获取支付记录
func (r *paymentRepository) GetPaymentByTransactionSN(ctx context.Context, sn string) (*model.PaymentTransaction, error) {
	var tx model.PaymentTransaction
	if err := r.db.WithContext(ctx).Where("transaction_sn = ?", sn).First(&tx).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 记录不存在
		}
		return nil, fmt.Errorf("数据库查询支付流水失败: %w", err)
	}
	return &tx, nil
}

// GetPaymentByOrderSN 根据订单号获取支付记录
// 注意：一个订单可能对应多个支付流水（例如，首次支付失败后重试），这里只取最新的
func (r *paymentRepository) GetPaymentByOrderSN(ctx context.Context, orderSN string) (*model.PaymentTransaction, error) {
	var tx model.PaymentTransaction
	if err := r.db.WithContext(ctx).Where("order_sn = ?", orderSN).Order("created_at desc").First(&tx).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询支付流水失败: %w", err)
	}
	return &tx, nil
}

// UpdatePaymentTransaction 更新支付流水记录
func (r *paymentRepository) UpdatePaymentTransaction(ctx context.Context, tx *model.PaymentTransaction) error {
	if err := r.db.WithContext(ctx).Save(tx).Error; err != nil {
		return fmt.Errorf("数据库更新支付流水失败: %w", err)
	}
	return nil
}

// CreateRefundTransaction 在数据库中创建一条新的退款流水记录
func (r *paymentRepository) CreateRefundTransaction(ctx context.Context, tx *model.RefundTransaction) error {
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		return fmt.Errorf("数据库创建退款流水失败: %w", err)
	}
	return nil
}

// GetRefundByRefundSN 根据系统内部退款流水号获取退款记录
func (r *paymentRepository) GetRefundByRefundSN(ctx context.Context, sn string) (*model.RefundTransaction, error) {
	var tx model.RefundTransaction
	if err := r.db.WithContext(ctx).Where("refund_sn = ?", sn).First(&tx).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询退款流水失败: %w", err)
	}
	return &tx, nil
}

// UpdateRefundTransaction 更新退款流水记录
func (r *paymentRepository) UpdateRefundTransaction(ctx context.Context, tx *model.RefundTransaction) error {
	if err := r.db.WithContext(ctx).Save(tx).Error; err != nil {
		return fmt.Errorf("数据库更新退款流水失败: %w", err)
	}
	return nil
}
