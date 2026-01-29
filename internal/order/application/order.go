package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// OrderService 是订单应用服务的门面。
type OrderService struct {
	Command *OrderCommandService
	Query   *OrderQueryService
}

// NewOrderService 构造函数。
func NewOrderService(command *OrderCommandService, query *OrderQueryService) *OrderService {
	return &OrderService{
		Command: command,
		Query:   query,
	}
}

// --- Delegate Methods ---

func (s *OrderService) CreateOrder(ctx context.Context, cmd *CreateOrderCommand) (*domain.Order, error) {
	return s.Command.CreateOrder(ctx, cmd)
}

func (s *OrderService) PayOrder(ctx context.Context, cmd *PayOrderCommand) error {
	return s.Command.PayOrder(ctx, cmd)
}

func (s *OrderService) ShipOrder(ctx context.Context, cmd *ShipOrderCommand) error {
	return s.Command.ShipOrder(ctx, cmd)
}

func (s *OrderService) DeliverOrder(ctx context.Context, cmd *DeliverOrderCommand) error {
	return s.Command.DeliverOrder(ctx, cmd)
}

func (s *OrderService) CompleteOrder(ctx context.Context, cmd *CompleteOrderCommand) error {
	return s.Command.CompleteOrder(ctx, cmd)
}

func (s *OrderService) CancelOrder(ctx context.Context, cmd *CancelOrderCommand) error {
	return s.Command.CancelOrder(ctx, cmd)
}

func (s *OrderService) GetOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Order, error) {
	return s.Query.GetOrder(ctx, userID, orderID)
}

func (s *OrderService) ListUserOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	offset := (page - 1) * pageSize
	return s.Query.ListUserOrders(ctx, userID, status, offset, pageSize)
}
