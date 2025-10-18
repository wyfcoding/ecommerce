package repository

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/payment/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentTransactionRepo 定义了支付交易数据的存储接口。
type PaymentTransactionRepo interface {
	// CreatePaymentTransaction 创建一个新的支付交易记录。
	CreatePaymentTransaction(ctx context.Context, transaction *model.PaymentTransaction) (*model.PaymentTransaction, error)
	// GetPaymentTransactionByID 根据ID获取支付交易详情。
	GetPaymentTransactionByID(ctx context.Context, id uint64) (*model.PaymentTransaction, error)
	// GetPaymentTransactionByTransactionNo 根据内部交易编号获取支付交易详情。
	GetPaymentTransactionByTransactionNo(ctx context.Context, transactionNo string) (*model.PaymentTransaction, error)
	// GetPaymentTransactionByGatewayTransactionID 根据支付网关交易ID获取支付交易详情。
	GetPaymentTransactionByGatewayTransactionID(ctx context.Context, gatewayTransactionID string) (*model.PaymentTransaction, error)
	// UpdatePaymentTransaction 更新支付交易信息。
	UpdatePaymentTransaction(ctx context.Context, transaction *model.PaymentTransaction) (*model.PaymentTransaction, error)
	// ListPaymentTransactions 根据条件分页查询支付交易列表。
	ListPaymentTransactions(ctx context.Context, query *PaymentTransactionListQuery) ([]*model.PaymentTransaction, int64, error)
}

// RefundTransactionRepo 定义了退款交易数据的存储接口。
type RefundTransactionRepo interface {
	// CreateRefundTransaction 创建一个新的退款交易记录。
	CreateRefundTransaction(ctx context.Context, transaction *model.RefundTransaction) (*model.RefundTransaction, error)
	// GetRefundTransactionByID 根据ID获取退款交易详情。
	GetRefundTransactionByID(ctx context.Context, id uint64) (*model.RefundTransaction, error)
	// GetRefundTransactionByRefundNo 根据内部退款编号获取退款交易详情。
	GetRefundTransactionByRefundNo(ctx context.Context, refundNo string) (*model.RefundTransaction, error)
	// GetRefundTransactionByGatewayRefundID 根据支付网关退款ID获取退款交易详情。
	GetRefundTransactionByGatewayRefundID(ctx context.Context, gatewayRefundID string) (*model.RefundTransaction, error)
	// UpdateRefundTransaction 更新退款交易信息。
	UpdateRefundTransaction(ctx context.Context, transaction *model.RefundTransaction) (*model.RefundTransaction, error)
	// ListRefundTransactions 根据条件分页查询退款交易列表。
	ListRefundTransactions(ctx context.Context, query *RefundTransactionListQuery) ([]*model.RefundTransaction, int64, error)
}

// PaymentTransactionListQuery 定义支付交易列表查询的参数。
type PaymentTransactionListQuery struct {
	Page      int32
	PageSize  int32
	UserID    uint64
	OrderID   uint64
	Status    model.PaymentStatus
	StartTime *time.Time
	EndTime   *time.Time
}

// RefundTransactionListQuery 定义退款交易列表查询的参数。
type RefundTransactionListQuery struct {
	Page      int32
	PageSize  int32
	UserID    uint64
	OrderID   uint64
	Status    model.RefundStatus
	StartTime *time.Time
	EndTime   *time.Time
}

// paymentTransactionRepoImpl 是 PaymentTransactionRepo 接口的 GORM 实现。
type paymentTransactionRepoImpl struct {
	db *gorm.DB
}

// NewPaymentTransactionRepo 创建一个新的 PaymentTransactionRepo 实例。
func NewPaymentTransactionRepo(db *gorm.DB) PaymentTransactionRepo {
	return &paymentTransactionRepoImpl{db: db}
}

// CreatePaymentTransaction 实现 CreatePaymentTransaction 方法。
func (r *paymentTransactionRepoImpl) CreatePaymentTransaction(ctx context.Context, transaction *model.PaymentTransaction) (*model.PaymentTransaction, error) {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		zap.S().Errorf("failed to create payment transaction: %v", err)
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}
	return transaction, nil
}

// GetPaymentTransactionByID 实现 GetPaymentTransactionByID 方法。
func (r *paymentTransactionRepoImpl) GetPaymentTransactionByID(ctx context.Context, id uint64) (*model.PaymentTransaction, error) {
	var transaction model.PaymentTransaction
	if err := r.db.WithContext(ctx).First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get payment transaction by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get payment transaction by id: %w", err)
	}
	return &transaction, nil
}

// GetPaymentTransactionByTransactionNo 实现 GetPaymentTransactionByTransactionNo 方法。
func (r *paymentTransactionRepoImpl) GetPaymentTransactionByTransactionNo(ctx context.Context, transactionNo string) (*model.PaymentTransaction, error) {
	var transaction model.PaymentTransaction
	if err := r.db.WithContext(ctx).Where("transaction_no = ?", transactionNo).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get payment transaction by transaction_no %s: %v", transactionNo, err)
		return nil, fmt.Errorf("failed to get payment transaction by transaction_no: %w", err)
	}
	return &transaction, nil
}

// GetPaymentTransactionByGatewayTransactionID 实现 GetPaymentTransactionByGatewayTransactionID 方法。
func (r *paymentTransactionRepoImpl) GetPaymentTransactionByGatewayTransactionID(ctx context.Context, gatewayTransactionID string) (*model.PaymentTransaction, error) {
	var transaction model.PaymentTransaction
	if err := r.db.WithContext(ctx).Where("gateway_transaction_id = ?", gatewayTransactionID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get payment transaction by gateway_transaction_id %s: %v", gatewayTransactionID, err)
		return nil, fmt.Errorf("failed to get payment transaction by gateway_transaction_id: %w", err)
	}
	return &transaction, nil
}

// UpdatePaymentTransaction 实现 UpdatePaymentTransaction 方法。
func (r *paymentTransactionRepoImpl) UpdatePaymentTransaction(ctx context.Context, transaction *model.PaymentTransaction) (*model.PaymentTransaction, error) {
	if err := r.db.WithContext(ctx).Save(transaction).Error; err != nil {
		zap.S().Errorf("failed to update payment transaction %d: %v", transaction.ID, err)
		return nil, fmt.Errorf("failed to update payment transaction: %w", err)
	}
	return transaction, nil
}

// ListPaymentTransactions 实现 ListPaymentTransactions 方法。
func (r *paymentTransactionRepoImpl) ListPaymentTransactions(ctx context.Context, query *PaymentTransactionListQuery) ([]*model.PaymentTransaction, int64, error) {
	var transactions []*model.PaymentTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&model.PaymentTransaction{})

	// 应用筛选条件
	if query.UserID != 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrderID != 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	if query.Status != model.PaymentStatusUnspecified {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		zap.S().Errorf("failed to count payment transactions: %v", err)
		return nil, 0, fmt.Errorf("failed to count payment transactions: %w", err)
	}

	// 应用分页
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(int(query.PageSize)).Offset(int(offset))
	}

	// 查询数据
	if err := db.Find(&transactions).Error; err != nil {
		zap.S().Errorf("failed to list payment transactions: %v", err)
		return nil, 0, fmt.Errorf("failed to list payment transactions: %w", err)
	}

	return transactions, total, nil
}

// refundTransactionRepoImpl 是 RefundTransactionRepo 接口的 GORM 实现。
type refundTransactionRepoImpl struct {
	db *gorm.DB
}

// NewRefundTransactionRepo 创建一个新的 RefundTransactionRepo 实例。
func NewRefundTransactionRepo(db *gorm.DB) RefundTransactionRepo {
	return &refundTransactionRepoImpl{db: db}
}

// CreateRefundTransaction 实现 CreateRefundTransaction 方法。
func (r *refundTransactionRepoImpl) CreateRefundTransaction(ctx context.Context, transaction *model.RefundTransaction) (*model.RefundTransaction, error) {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		zap.S().Errorf("failed to create refund transaction: %v", err)
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}
	return transaction, nil
}

// GetRefundTransactionByID 实现 GetRefundTransactionByID 方法。
func (r *refundTransactionRepoImpl) GetRefundTransactionByID(ctx context.Context, id uint64) (*model.RefundTransaction, error) {
	var transaction model.RefundTransaction
	if err := r.db.WithContext(ctx).First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get refund transaction by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get refund transaction by id: %w", err)
	}
	return &transaction, nil
}

// GetRefundTransactionByRefundNo 实现 GetRefundTransactionByRefundNo 方法。
func (r *refundTransactionRepoImpl) GetRefundTransactionByRefundNo(ctx context.Context, refundNo string) (*model.RefundTransaction, error) {
	var transaction model.RefundTransaction
	if err := r.db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get refund transaction by refund_no %s: %v", refundNo, err)
		return nil, fmt.Errorf("failed to get refund transaction by refund_no: %w", err)
	}
	return &transaction, nil
}

// GetRefundTransactionByGatewayRefundID 实现 GetRefundTransactionByGatewayRefundID 方法。
func (r *refundTransactionRepoImpl) GetRefundTransactionByGatewayRefundID(ctx context.Context, gatewayRefundID string) (*model.RefundTransaction, error) {
	var transaction model.RefundTransaction
	if err := r.db.WithContext(ctx).Where("gateway_refund_id = ?", gatewayRefundID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get refund transaction by gateway_refund_id %s: %v", gatewayRefundID, err)
		return nil, fmt.Errorf("failed to get refund transaction by gateway_refund_id: %w", err)
	}
	return &transaction, nil
}

// UpdateRefundTransaction 实现 UpdateRefundTransaction 方法。
func (r *refundTransactionRepoImpl) UpdateRefundTransaction(ctx context.Context, transaction *model.RefundTransaction) (*model.RefundTransaction, error) {
	if err := r.db.WithContext(ctx).Save(transaction).Error; err != nil {
		zap.S().Errorf("failed to update refund transaction %d: %v", transaction.ID, err)
		return nil, fmt.Errorf("failed to update refund transaction: %w", err)
	}
	return transaction, nil
}

// ListRefundTransactions 实现 ListRefundTransactions 方法。
func (r *refundTransactionRepoImpl) ListRefundTransactions(ctx context.Context, query *RefundTransactionListQuery) ([]*model.RefundTransaction, int64, error) {
	var transactions []*model.RefundTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&model.RefundTransaction{})

	// 应用筛选条件
	if query.UserID != 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrderID != 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	if query.Status != model.RefundStatusUnspecified {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		zap.S().Errorf("failed to count refund transactions: %v", err)
		return nil, 0, fmt.Errorf("failed to count refund transactions: %w", err)
	}

	// 应用分页
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(int(query.PageSize)).Offset(int(offset))
	}

	// 查询数据
	if err := db.Find(&transactions).Error; err != nil {
		zap.S().Errorf("failed to list refund transactions: %v", err)
		return nil, 0, fmt.Errorf("failed to list refund transactions: %w", err)
	}

	return transactions, total, nil
}
