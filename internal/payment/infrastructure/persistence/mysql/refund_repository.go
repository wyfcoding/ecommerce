package mysql

import (
	"context"
	"errors"
	"time"

	paypb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
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

func (r *refundRepository) List(ctx context.Context, userID uint64, orderID uint64, status paypb.RefundStatus, startTime, endTime *time.Time, page, pageSize int) ([]*domain.Refund, int64, error) {
	db := r.getDB(userID)
	query := db.WithContext(ctx).Model(&RefundModel{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if orderID > 0 {
		query = query.Where("order_id = ?", orderID)
	}
	if mappedStatus, ok := mapRefundStatus(status); ok {
		query = query.Where("status = ?", mappedStatus)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var models []RefundModel
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Refund, 0, len(models))
	for i := range models {
		result = append(result, toDomainRefund(&models[i]))
	}

	return result, total, nil
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

func mapRefundStatus(status paypb.RefundStatus) (domain.PaymentStatus, bool) {
	switch status {
	case paypb.RefundStatus_PENDING_REFUND:
		return domain.PaymentStatus(paypb.PaymentStatus_REFUNDING), true
	case paypb.RefundStatus_REFUND_SUCCESS:
		return domain.PaymentStatus(paypb.PaymentStatus_REFUNDED), true
	case paypb.RefundStatus_REFUND_FAILED:
		return domain.PaymentStatus(paypb.PaymentStatus_FAILED), true
	case paypb.RefundStatus_REFUND_CLOSED:
		return domain.PaymentStatus(paypb.PaymentStatus_CLOSED), true
	default:
		return domain.PaymentStatus(paypb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED), false
	}
}
