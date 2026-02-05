// 生成摘要：新增订单读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以订单ID为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// OrderProjectionService 负责将事件转换为读模型更新。
type OrderProjectionService struct {
	repo       domain.OrderRepository
	readRepo   domain.OrderReadRepository
	searchRepo domain.OrderSearchRepository
	logger     *slog.Logger
}

// NewOrderProjectionService 创建订单投影服务。
func NewOrderProjectionService(repo domain.OrderRepository, readRepo domain.OrderReadRepository, searchRepo domain.OrderSearchRepository, logger *slog.Logger) *OrderProjectionService {
	return &OrderProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnOrderCreated 处理订单创建事件。
func (s *OrderProjectionService) OnOrderCreated(ctx context.Context, event *domain.OrderCreatedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderPaid 处理订单支付事件。
func (s *OrderProjectionService) OnOrderPaid(ctx context.Context, event *domain.OrderPaidEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderShipped 处理订单发货事件。
func (s *OrderProjectionService) OnOrderShipped(ctx context.Context, event *domain.OrderShippedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderDelivered 处理订单送达事件。
func (s *OrderProjectionService) OnOrderDelivered(ctx context.Context, event *domain.OrderDeliveredEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderCompleted 处理订单完成事件。
func (s *OrderProjectionService) OnOrderCompleted(ctx context.Context, event *domain.OrderCompletedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderCancelled 处理订单取消事件。
func (s *OrderProjectionService) OnOrderCancelled(ctx context.Context, event *domain.OrderCancelledEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderConfirmed 处理订单确认事件。
func (s *OrderProjectionService) OnOrderConfirmed(ctx context.Context, event *domain.OrderConfirmedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderRefundRequested 处理订单退款申请事件。
func (s *OrderProjectionService) OnOrderRefundRequested(ctx context.Context, event *domain.OrderRefundRequestedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// OnOrderRefundApproved 处理订单退款完成事件。
func (s *OrderProjectionService) OnOrderRefundApproved(ctx context.Context, event *domain.OrderRefundApprovedEvent) error {
	return s.refreshReadModel(ctx, event.UserID, event.OrderID, event.OrderNo)
}

// refreshReadModel 从写模型加载订单并刷新读侧。
func (s *OrderProjectionService) refreshReadModel(ctx context.Context, userID, orderID uint64, orderNo string) error {
	order, err := s.repo.FindByID(ctx, userID, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load order for projection", "order_id", orderID, "error", err)
		return err
	}

	if order == nil {
		// 清理读侧残留
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, userID, orderID, orderNo)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, orderID)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to save order read model", "order_id", orderID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to index order search model", "order_id", orderID, "error", err)
			return err
		}
	}

	return nil
}
