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

// NewOrderService 创建订单服务实例。
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

// CreateOrder 创建订单。
func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress, couponCode string) (*domain.Order, error) {
	return s.Manager.CreateOrder(ctx, userID, items, shippingAddr, couponCode)
}

func (s *OrderService) SagaConfirmOrder(ctx context.Context, userID, orderID uint64) error {
	return s.Manager.SagaConfirmOrder(ctx, userID, orderID)
}

func (s *OrderService) SagaCancelOrder(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.Manager.SagaCancelOrder(ctx, userID, orderID, reason)
}

func (s *OrderService) SetSvcURL(url string) {
	s.Manager.SetSvcURL(url)
}

// PayOrder 支付订单。
func (s *OrderService) PayOrder(ctx context.Context, userID, id uint64, paymentMethod string) error {
	return s.Manager.PayOrder(ctx, userID, id, paymentMethod)
}

// ShipOrder 发货订单。
func (s *OrderService) ShipOrder(ctx context.Context, userID, id uint64, operator string) error {
	return s.Manager.ShipOrder(ctx, userID, id, operator)
}

// DeliverOrder 确认收货。
func (s *OrderService) DeliverOrder(ctx context.Context, userID, id uint64, operator string) error {
	return s.Manager.DeliverOrder(ctx, userID, id, operator)
}

// CompleteOrder 完成订单。
func (s *OrderService) CompleteOrder(ctx context.Context, userID, id uint64, operator string) error {
	return s.Manager.CompleteOrder(ctx, userID, id, operator)
}

// CancelOrder 取消订单。
func (s *OrderService) CancelOrder(ctx context.Context, userID, id uint64, operator, reason string) error {
	return s.Manager.CancelOrder(ctx, userID, id, operator, reason)
}

// HandleInventoryReserved 处理库存预留成功事件。
func (s *OrderService) HandleInventoryReserved(ctx context.Context, userID, orderID uint64) error {
	return s.Manager.HandleInventoryReserved(ctx, userID, orderID)
}

// HandleInventoryReservationFailed 处理库存预留失败事件。
func (s *OrderService) HandleInventoryReservationFailed(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.Manager.HandleInventoryReservationFailed(ctx, userID, orderID, reason)
}

// HandlePaymentProcessed 处理支付成功事件。
func (s *OrderService) HandlePaymentProcessed(ctx context.Context, userID, orderID uint64) error {
	return s.Manager.HandlePaymentProcessed(ctx, userID, orderID)
}

// HandlePaymentFailed 处理支付失败事件。
func (s *OrderService) HandlePaymentFailed(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.Manager.HandlePaymentFailed(ctx, userID, orderID, reason)
}

// GetOrder 获取订单详情。
func (s *OrderService) GetOrder(ctx context.Context, userID, id uint64) (*domain.Order, error) {
	return s.Query.GetOrder(ctx, userID, id)
}

// ListOrders 获取订单列表。
func (s *OrderService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	return s.Query.ListOrders(ctx, userID, status, page, pageSize)
}
