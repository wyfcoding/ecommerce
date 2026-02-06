package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentQuery 支付查询服务。
type PaymentQuery struct {
	paymentRepo domain.PaymentRepository
	readRepo    domain.PaymentReadRepository
	searchRepo  domain.PaymentSearchRepository
	eventStore  domain.EventStore
	logger      *slog.Logger
}

// NewPaymentQuery 构造函数。
func NewPaymentQuery(paymentRepo domain.PaymentRepository, readRepo domain.PaymentReadRepository, searchRepo domain.PaymentSearchRepository, eventStore domain.EventStore, logger *slog.Logger) *PaymentQuery {
	return &PaymentQuery{
		paymentRepo: paymentRepo,
		readRepo:    readRepo,
		searchRepo:  searchRepo,
		eventStore:  eventStore,
		logger:      logger,
	}
}

// GetPayment 获取支付详情 (按 ID)。
func (q *PaymentQuery) GetPayment(ctx context.Context, userID uint64, paymentID uint64) (*domain.Payment, error) {
	if q.readRepo != nil {
		if payment, err := q.readRepo.GetByID(ctx, userID, paymentID); err == nil && payment != nil {
			return payment, nil
		}
	}

	return q.paymentRepo.FindByID(ctx, userID, paymentID)
}

// GetPaymentByNo 获取支付详情 (按支付单号)。
func (q *PaymentQuery) GetPaymentByNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
	if q.readRepo != nil {
		if payment, err := q.readRepo.GetByPaymentNo(ctx, userID, paymentNo); err == nil && payment != nil {
			return payment, nil
		}
	}
	if q.searchRepo != nil {
		if payment, err := q.searchRepo.FindByPaymentNo(ctx, paymentNo); err == nil && payment != nil {
			return payment, nil
		}
	}

	return q.paymentRepo.FindByPaymentNo(ctx, userID, paymentNo)
}

// GetPaymentByOrder 获取订单支付信息。
func (q *PaymentQuery) GetPaymentByOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	if q.readRepo != nil {
		if payment, err := q.readRepo.GetByOrderID(ctx, userID, orderID); err == nil && payment != nil {
			return payment, nil
		}
	}
	if q.searchRepo != nil {
		if payment, err := q.searchRepo.FindByOrderID(ctx, orderID); err == nil && payment != nil {
			return payment, nil
		}
	}

	return q.paymentRepo.FindByOrderID(ctx, userID, orderID)
}

// GetPaymentLogs 获取支付日志。
func (q *PaymentQuery) GetPaymentLogs(ctx context.Context, userID uint64, paymentID uint64) ([]*domain.PaymentLog, error) {
	return q.paymentRepo.FindLogsByPaymentID(ctx, userID, paymentID)
}

// GetUserIDByPaymentNo 根据支付单号查找用户 ID。
func (q *PaymentQuery) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	return q.paymentRepo.GetUserIDByPaymentNo(ctx, paymentNo)
}

// GetPaymentStatus 获取支付状态 (按单号或 ID 的通用查询)。
func (q *PaymentQuery) GetPaymentStatus(ctx context.Context, userID uint64, identifier any) (*domain.Payment, error) {
	switch v := identifier.(type) {
	case string:
		payment, err := q.GetPaymentByNo(ctx, userID, v)
		if err != nil {
			return nil, err
		}
		if payment != nil {
			return payment, nil
		}
		if q.eventStore != nil {
			events, loadErr := q.eventStore.GetHistory(ctx, v)
			if loadErr != nil {
				q.logger.WarnContext(ctx, "event store load failed", "payment_no", v, "error", loadErr)
				return nil, err
			}
			if len(events) > 0 {
				p, rebuildErr := domain.RebuildPaymentFromEvents(events)
				if rebuildErr != nil {
					q.logger.WarnContext(ctx, "payment rebuild failed", "payment_no", v, "error", rebuildErr)
					return nil, err
				}
				return p, nil
			}
		}
		return nil, err
	case uint64:
		return q.paymentRepo.FindByID(ctx, userID, v)
	default:
		return nil, domain.ErrInvalidParameter
	}
}
