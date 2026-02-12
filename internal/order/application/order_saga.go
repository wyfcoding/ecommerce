package application

import (
	"context"
	"fmt"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// SagaConfirmOrder Saga 确认。
func (s *OrderCommandService) SagaConfirmOrder(ctx context.Context, userID, orderID uint64) error {
	return s.repo.WithTx(ctx, userID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, userID, orderID)
		if err != nil || order == nil {
			return fmt.Errorf("order not found: %d", orderID)
		}
		if order.Status != orderv1.OrderStatus_ALLOCATING {
			return nil
		}
		oldStatus := order.Status
		if err := order.Trigger(ctx, "CONFIRM", "System", "Saga success"); err != nil {
			return err
		}
		confirmedPayload := &domain.OrderConfirmedPayload{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			OldStatus: oldStatus,
			NewStatus: order.Status,
			Confirmed: time.Now(),
			Log:       buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeConfirmed, confirmedPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.confirmed", order.OrderNo, &domain.OrderConfirmedEvent{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			Timestamp: time.Now(),
		})
	})
}

// SagaCancelOrder Saga 取消。
func (s *OrderCommandService) SagaCancelOrder(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.repo.WithTx(ctx, userID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, userID, orderID)
		if err != nil || order == nil {
			return nil
		}
		if order.Status == orderv1.OrderStatus_CANCELLED {
			return nil
		}
		oldStatus := order.Status
		if err := order.Cancel(ctx, "System", reason); err != nil {
			return err
		}
		cancelledPayload := &domain.OrderCancelledPayload{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			OldStatus:   oldStatus,
			NewStatus:   order.Status,
			Reason:      reason,
			CancelledAt: time.Now(),
			Log:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCancelled, cancelledPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.cancelled", order.OrderNo, &domain.OrderCancelledEvent{
			OrderID:        uint64(order.ID),
			OrderNo:        order.OrderNo,
			UserID:         order.UserID,
			Reason:         reason,
			PaymentStatus:  order.PaymentStatus,
			ShippingStatus: order.ShippingStatus,
			CancelledAt:    *order.CancelledAt,
			Timestamp:      time.Now(),
		})
	})
}
