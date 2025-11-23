package application

import (
	"context"
	"ecommerce/internal/order/domain/entity"
	"ecommerce/internal/order/domain/repository"
	"errors"
	"fmt"
	"time"

	"ecommerce/pkg/idgen"

	"log/slog"
)

type OrderService struct {
	repo   repository.OrderRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

func NewOrderService(repo repository.OrderRepository, idGen idgen.Generator, logger *slog.Logger) *OrderService {
	return &OrderService{
		repo:   repo,
		idGen:  idGen,
		logger: logger,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, items []*entity.OrderItem, shippingAddr *entity.ShippingAddress) (*entity.Order, error) {
	// Generate Order No
	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := entity.NewOrder(orderNo, userID, items, shippingAddr)
	if err := s.repo.Save(ctx, order); err != nil {
		s.logger.Error("failed to create order", "error", err)
		return nil, err
	}
	return order, nil
}

// GetOrder 获取订单
func (s *OrderService) GetOrder(ctx context.Context, id uint64) (*entity.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(ctx context.Context, id uint64, paymentMethod string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Pay(paymentMethod, "User"); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// ShipOrder 发货
func (s *OrderService) ShipOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Ship(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// DeliverOrder 送达
func (s *OrderService) DeliverOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Deliver(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// CompleteOrder 完成订单
func (s *OrderService) CompleteOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Complete(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, id uint64, operator, reason string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Cancel(operator, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Order, int64, error) {
	offset := (page - 1) * pageSize
	var orderStatus *entity.OrderStatus
	if status != nil {
		s := entity.OrderStatus(*status)
		orderStatus = &s
	}
	return s.repo.List(ctx, userID, orderStatus, offset, pageSize)
}
