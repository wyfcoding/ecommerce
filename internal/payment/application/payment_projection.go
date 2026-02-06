// 生成摘要：新增支付读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以支付ID为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentProjectionService 负责将事件转换为读模型更新。
type PaymentProjectionService struct {
	repo       domain.PaymentRepository
	readRepo   domain.PaymentReadRepository
	searchRepo domain.PaymentSearchRepository
	logger     *slog.Logger
}

// NewPaymentProjectionService 创建支付投影服务。
func NewPaymentProjectionService(repo domain.PaymentRepository, readRepo domain.PaymentReadRepository, searchRepo domain.PaymentSearchRepository, logger *slog.Logger) *PaymentProjectionService {
	return &PaymentProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnPaymentInitiated 处理支付创建事件。
func (s *PaymentProjectionService) OnPaymentInitiated(ctx context.Context, event *domain.PaymentInitiatedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// OnPaymentAuthorized 处理支付授权事件。
func (s *PaymentProjectionService) OnPaymentAuthorized(ctx context.Context, event *domain.PaymentAuthorizedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// OnPaymentCaptured 处理支付捕获事件。
func (s *PaymentProjectionService) OnPaymentCaptured(ctx context.Context, event *domain.PaymentCapturedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// OnPaymentPaid 处理支付完成事件。
func (s *PaymentProjectionService) OnPaymentPaid(ctx context.Context, event *domain.PaymentPaidEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// OnRefundFinished 处理退款完成事件。
func (s *PaymentProjectionService) OnRefundFinished(ctx context.Context, event *domain.RefundFinishedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// OnPaymentClosed 处理支付关闭事件。
func (s *PaymentProjectionService) OnPaymentClosed(ctx context.Context, event *domain.PaymentClosedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.AggregateID(), event.OrderID)
}

// refreshReadModel 从写模型加载支付并刷新读侧。
func (s *PaymentProjectionService) refreshReadModel(ctx context.Context, userID uint64, paymentNo string, orderID uint64) error {
	if paymentNo == "" {
		return nil
	}

	if userID == 0 && s.repo != nil {
		lookupID, err := s.repo.GetUserIDByPaymentNo(ctx, paymentNo)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to lookup user_id for payment projection", "payment_no", paymentNo, "error", err)
		} else {
			userID = lookupID
		}
	}

	var payment *domain.Payment
	var err error
	if userID != 0 {
		payment, err = s.repo.FindByPaymentNo(ctx, userID, paymentNo)
	} else if orderID != 0 {
		payment, err = s.repo.FindByOrderID(ctx, userID, orderID)
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load payment for projection", "payment_no", paymentNo, "error", err)
		return err
	}

	if payment == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, userID, 0, paymentNo, orderID)
		}
		if s.searchRepo != nil {
			// 不知道 paymentID 时无法精准删除，忽略即可
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, payment); err != nil {
			s.logger.ErrorContext(ctx, "failed to save payment read model", "payment_no", paymentNo, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, payment); err != nil {
			s.logger.ErrorContext(ctx, "failed to index payment search model", "payment_no", paymentNo, "error", err)
			return err
		}
	}

	return nil
}
