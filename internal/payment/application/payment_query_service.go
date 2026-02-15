package application

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type PaymentQueryService struct {
	paymentRepo domain.PaymentRepository
	refundRepo  domain.RefundRepository
	readRepo    domain.PaymentReadRepository
	searchRepo  domain.PaymentSearchRepository
	eventStore  domain.EventStore
	logger      *slog.Logger
}

func NewPaymentQueryService(paymentRepo domain.PaymentRepository, refundRepo domain.RefundRepository, readRepo domain.PaymentReadRepository, searchRepo domain.PaymentSearchRepository, eventStore domain.EventStore, logger *slog.Logger) *PaymentQueryService {
	return &PaymentQueryService{
		paymentRepo: paymentRepo,
		refundRepo:  refundRepo,
		readRepo:    readRepo,
		searchRepo:  searchRepo,
		eventStore:  eventStore,
		logger:      logger,
	}
}

// GetPayment 获取支付详情 (按 ID)。
func (q *PaymentQueryService) GetPayment(ctx context.Context, userID uint64, paymentID uint64) (*domain.Payment, error) {
	if q.readRepo != nil {
		if payment, err := q.readRepo.GetByID(ctx, userID, paymentID); err == nil && payment != nil {
			return payment, nil
		}
	}

	return q.paymentRepo.FindByID(ctx, userID, paymentID)
}

// GetPaymentByNo 获取支付详情 (按支付单号)。
func (q *PaymentQueryService) GetPaymentByNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
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
func (q *PaymentQueryService) GetPaymentByOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
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
func (q *PaymentQueryService) GetPaymentLogs(ctx context.Context, userID uint64, paymentID uint64) ([]*domain.PaymentLog, error) {
	return q.paymentRepo.FindLogsByPaymentID(ctx, userID, paymentID)
}

// GetUserIDByPaymentNo 根据支付单号查找用户 ID。
func (q *PaymentQueryService) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	return q.paymentRepo.GetUserIDByPaymentNo(ctx, paymentNo)
}

// GetPaymentStatus 获取支付状态 (按单号或 ID 的通用查询)。
func (q *PaymentQueryService) GetPaymentStatus(ctx context.Context, userID uint64, identifier any) (*domain.Payment, error) {
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
			events, loadErr := q.eventStore.Load(ctx, v)
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

type ListPaymentsFilter struct {
	UserID    uint64
	OrderID   uint64
	Status    pb.PaymentStatus
	StartTime *time.Time
	EndTime   *time.Time
	Page      int32
	PageSize  int32
}

type ListPaymentsResult struct {
	Transactions []*domain.Payment
	Total        int32
	Page         int32
	PageSize     int32
}

func (q *PaymentQueryService) ListPaymentTransactions(ctx context.Context, filter *ListPaymentsFilter) (*ListPaymentsResult, error) {
	q.logger.DebugContext(ctx, "ListPaymentTransactions called", "user_id", filter.UserID, "page", filter.Page, "page_size", filter.PageSize)

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	offset := int((filter.Page - 1) * filter.PageSize)
	var transactions []*domain.Payment
	var total int64
	var err error

	if q.searchRepo != nil {
		transactions, total, err = q.searchRepo.Search(ctx, &filter.UserID, &filter.Status, offset, int(filter.PageSize), filter.StartTime, filter.EndTime, "created_at DESC")
		if err != nil {
			q.logger.WarnContext(ctx, "search repo failed, falling back to db", "error", err)
		}
	}

	if transactions == nil {
		transactions, total, err = q.listFromDB(ctx, filter)
		if err != nil {
			return nil, err
		}
	}

	return &ListPaymentsResult{
		Transactions: transactions,
		Total:        int32(total),
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	}, nil
}

func (q *PaymentQueryService) listFromDB(ctx context.Context, filter *ListPaymentsFilter) ([]*domain.Payment, int64, error) {
	return []*domain.Payment{}, 0, nil
}

type ListRefundsFilter struct {
	UserID    uint64
	OrderID   uint64
	Status    pb.RefundStatus
	StartTime *time.Time
	EndTime   *time.Time
	Page      int32
	PageSize  int32
}

type ListRefundsResult struct {
	Transactions []*domain.Refund
	Total        int32
	Page         int32
	PageSize     int32
}

func (q *PaymentQueryService) GetRefundStatus(ctx context.Context, userID uint64, refundNo string) (*domain.Refund, error) {
	q.logger.DebugContext(ctx, "GetRefundStatus called", "user_id", userID, "refund_no", refundNo)

	if q.refundRepo == nil {
		return nil, domain.ErrInvalidParameter
	}

	refund, err := q.refundRepo.FindByRefundNo(ctx, userID, refundNo)
	if err != nil {
		return nil, err
	}

	return refund, nil
}

func (q *PaymentQueryService) ListRefundTransactions(ctx context.Context, filter *ListRefundsFilter) (*ListRefundsResult, error) {
	q.logger.DebugContext(ctx, "ListRefundTransactions called", "user_id", filter.UserID, "page", filter.Page, "page_size", filter.PageSize)

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if q.refundRepo == nil {
		return &ListRefundsResult{
			Transactions: []*domain.Refund{},
			Total:        0,
			Page:         filter.Page,
			PageSize:     filter.PageSize,
		}, nil
	}

	transactions, total, err := q.refundRepo.List(ctx, filter.UserID, filter.OrderID, filter.Status, filter.StartTime, filter.EndTime, int(filter.Page), int(filter.PageSize))
	if err != nil {
		return nil, err
	}

	return &ListRefundsResult{
		Transactions: transactions,
		Total:        int32(total),
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	}, nil
}
