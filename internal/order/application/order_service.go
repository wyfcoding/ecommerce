package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// OrderService 订单服务门面（Facade）
type OrderService struct {
	Manager *OrderManager
	Query   *OrderQuery
	logger  *slog.Logger
}

// NewOrderService 定义了 NewOrder 相关的服务逻辑。
func NewOrderService(
	manager *OrderManager,
	query *OrderQuery,
	logger *slog.Logger,
) *OrderService {
	return &OrderService{
		Manager: manager,
		Query:   query,
		logger:  logger,
	}
}

// --- Delegate Methods ---

func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress) (*domain.Order, error) {
	return s.Manager.CreateOrder(ctx, userID, items, shippingAddr)
}

func (s *OrderService) PayOrder(ctx context.Context, id uint64, paymentMethod string) error {
	return s.Manager.PayOrder(ctx, id, paymentMethod)
}

func (s *OrderService) ShipOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.ShipOrder(ctx, id, operator)
}

func (s *OrderService) DeliverOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.DeliverOrder(ctx, id, operator)
}

func (s *OrderService) CompleteOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.CompleteOrder(ctx, id, operator)
}

func (s *OrderService) CancelOrder(ctx context.Context, id uint64, operator, reason string) error {
	return s.Manager.CancelOrder(ctx, id, operator, reason)
}

func (s *OrderService) HandleInventoryReserved(ctx context.Context, orderID uint64) error {
	return s.Manager.HandleInventoryReserved(ctx, orderID)
}

func (s *OrderService) HandleInventoryReservationFailed(ctx context.Context, orderID uint64, reason string) error {
	return s.Manager.HandleInventoryReservationFailed(ctx, orderID, reason)
}

func (s *OrderService) HandlePaymentProcessed(ctx context.Context, orderID uint64) error {
	return s.Manager.HandlePaymentProcessed(ctx, orderID)
}

func (s *OrderService) HandlePaymentFailed(ctx context.Context, orderID uint64, reason string) error {
	return s.Manager.HandlePaymentFailed(ctx, orderID, reason)
}

func (s *OrderService) GetOrder(ctx context.Context, id uint64) (*domain.Order, error) {
	return s.Query.GetOrder(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	return s.Query.ListOrders(ctx, userID, status, page, pageSize)
}
