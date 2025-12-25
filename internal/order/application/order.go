package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// Order 订单服务门面（Facade）
type Order struct {
	Manager *OrderManager
	Query   *OrderQuery
	logger  *slog.Logger
}

// NewOrder 创建订单服务实例。
func NewOrder(
	manager *OrderManager,
	query *OrderQuery,
	logger *slog.Logger,
) *Order {
	return &Order{
		Manager: manager,
		Query:   query,
		logger:  logger,
	}
}

// --- Delegate Methods ---

// CreateOrder 创建订单。
func (s *Order) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress) (*domain.Order, error) {
	return s.Manager.CreateOrder(ctx, userID, items, shippingAddr)
}

// PayOrder 支付订单。
func (s *Order) PayOrder(ctx context.Context, id uint64, paymentMethod string) error {
	return s.Manager.PayOrder(ctx, id, paymentMethod)
}

// ShipOrder 发货订单。
func (s *Order) ShipOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.ShipOrder(ctx, id, operator)
}

// DeliverOrder 确认收货。
func (s *Order) DeliverOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.DeliverOrder(ctx, id, operator)
}

// CompleteOrder 完成订单。
func (s *Order) CompleteOrder(ctx context.Context, id uint64, operator string) error {
	return s.Manager.CompleteOrder(ctx, id, operator)
}

// CancelOrder 取消订单。
func (s *Order) CancelOrder(ctx context.Context, id uint64, operator, reason string) error {
	return s.Manager.CancelOrder(ctx, id, operator, reason)
}

// HandleInventoryReserved 处理库存预留成功事件。
func (s *Order) HandleInventoryReserved(ctx context.Context, orderID uint64) error {
	return s.Manager.HandleInventoryReserved(ctx, orderID)
}

// HandleInventoryReservationFailed 处理库存预留失败事件。
func (s *Order) HandleInventoryReservationFailed(ctx context.Context, orderID uint64, reason string) error {
	return s.Manager.HandleInventoryReservationFailed(ctx, orderID, reason)
}

// HandlePaymentProcessed 处理支付成功事件。
func (s *Order) HandlePaymentProcessed(ctx context.Context, orderID uint64) error {
	return s.Manager.HandlePaymentProcessed(ctx, orderID)
}

// HandlePaymentFailed 处理支付失败事件。
func (s *Order) HandlePaymentFailed(ctx context.Context, orderID uint64, reason string) error {
	return s.Manager.HandlePaymentFailed(ctx, orderID, reason)
}

// GetOrder 获取订单详情。
func (s *Order) GetOrder(ctx context.Context, id uint64) (*domain.Order, error) {
	return s.Query.GetOrder(ctx, id)
}

// ListOrders 获取订单列表。
func (s *Order) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	return s.Query.ListOrders(ctx, userID, status, page, pageSize)
}
