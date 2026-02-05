// 生成摘要：新增订单事件溯源载荷、事件类型常量与回放函数。
// 假设：事件数据以 JSON 存储，回放时通过事件类型进行解析与状态恢复。
package domain

import (
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/pkg/eventsourcing"
)

const (
	// OrderEventTypeCreated 表示订单创建事件（事件溯源专用）。
	OrderEventTypeCreated = "order.es.created"
	// OrderEventTypePaid 表示订单支付事件（事件溯源专用）。
	OrderEventTypePaid = "order.es.paid"
	// OrderEventTypeShipped 表示订单发货事件（事件溯源专用）。
	OrderEventTypeShipped = "order.es.shipped"
	// OrderEventTypeDelivered 表示订单送达事件（事件溯源专用）。
	OrderEventTypeDelivered = "order.es.delivered"
	// OrderEventTypeCompleted 表示订单完成事件（事件溯源专用）。
	OrderEventTypeCompleted = "order.es.completed"
	// OrderEventTypeCancelled 表示订单取消事件（事件溯源专用）。
	OrderEventTypeCancelled = "order.es.cancelled"
	// OrderEventTypeConfirmed 表示订单确认事件（事件溯源专用）。
	OrderEventTypeConfirmed = "order.es.confirmed"
	// OrderEventTypeRefundRequested 表示订单退款申请事件（事件溯源专用）。
	OrderEventTypeRefundRequested = "order.es.refund_requested"
	// OrderEventTypeRefundApproved 表示订单退款完成事件（事件溯源专用）。
	OrderEventTypeRefundApproved = "order.es.refund_approved"
)

// OrderEventLog 定义事件溯源中的操作日志载荷。
type OrderEventLog struct {
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Remark    string    `json:"remark"`
	LoggedAt  time.Time `json:"logged_at"`
}

// OrderCreatedPayload 定义订单创建事件载荷。
type OrderCreatedPayload struct {
	OrderID              uint64           `json:"order_id"`
	OrderNo              string           `json:"order_no"`
	UserID               uint64           `json:"user_id"`
	Status               pb.OrderStatus   `json:"status"`
	TotalAmount          int64            `json:"total_amount"`
	ActualAmount         int64            `json:"actual_amount"`
	ShippingFee          int64            `json:"shipping_fee"`
	DiscountAmount       int64            `json:"discount_amount"`
	PaymentMethod        string           `json:"payment_method"`
	PaymentTransactionID string           `json:"payment_transaction_id"`
	Remark               string           `json:"remark"`
	TrackingNumber       string           `json:"tracking_number"`
	LogisticsCompany     string           `json:"logistics_company"`
	RefundAmount         int64            `json:"refund_amount"`
	RefundReason         string           `json:"refund_reason"`
	ShippingAddress      *ShippingAddress `json:"shipping_address"`
	Items                []*OrderItem     `json:"items"`
	CreatedAt            time.Time        `json:"created_at"`
	InitLog              *OrderEventLog   `json:"init_log"`
}

// OrderPaidPayload 定义订单支付事件载荷。
type OrderPaidPayload struct {
	OrderID              uint64         `json:"order_id"`
	OrderNo              string         `json:"order_no"`
	UserID               uint64         `json:"user_id"`
	PaymentMethod        string         `json:"payment_method"`
	PaymentTransactionID string         `json:"payment_transaction_id"`
	OldStatus            pb.OrderStatus `json:"old_status"`
	NewStatus            pb.OrderStatus `json:"new_status"`
	PaidAt               time.Time      `json:"paid_at"`
	Log                  *OrderEventLog `json:"log"`
}

// OrderShippedPayload 定义订单发货事件载荷。
type OrderShippedPayload struct {
	OrderID          uint64         `json:"order_id"`
	OrderNo          string         `json:"order_no"`
	UserID           uint64         `json:"user_id"`
	OldStatus        pb.OrderStatus `json:"old_status"`
	NewStatus        pb.OrderStatus `json:"new_status"`
	TrackingNumber   string         `json:"tracking_number"`
	LogisticsCompany string         `json:"logistics_company"`
	ShippedAt        time.Time      `json:"shipped_at"`
	Log              *OrderEventLog `json:"log"`
}

// OrderDeliveredPayload 定义订单送达事件载荷。
type OrderDeliveredPayload struct {
	OrderID     uint64         `json:"order_id"`
	OrderNo     string         `json:"order_no"`
	UserID      uint64         `json:"user_id"`
	OldStatus   pb.OrderStatus `json:"old_status"`
	NewStatus   pb.OrderStatus `json:"new_status"`
	DeliveredAt time.Time      `json:"delivered_at"`
	Log         *OrderEventLog `json:"log"`
}

// OrderCompletedPayload 定义订单完成事件载荷。
type OrderCompletedPayload struct {
	OrderID     uint64         `json:"order_id"`
	OrderNo     string         `json:"order_no"`
	UserID      uint64         `json:"user_id"`
	OldStatus   pb.OrderStatus `json:"old_status"`
	NewStatus   pb.OrderStatus `json:"new_status"`
	CompletedAt time.Time      `json:"completed_at"`
	Log         *OrderEventLog `json:"log"`
}

// OrderCancelledPayload 定义订单取消事件载荷。
type OrderCancelledPayload struct {
	OrderID     uint64         `json:"order_id"`
	OrderNo     string         `json:"order_no"`
	UserID      uint64         `json:"user_id"`
	OldStatus   pb.OrderStatus `json:"old_status"`
	NewStatus   pb.OrderStatus `json:"new_status"`
	Reason      string         `json:"reason"`
	CancelledAt time.Time      `json:"cancelled_at"`
	Log         *OrderEventLog `json:"log"`
}

// OrderConfirmedPayload 定义订单确认事件载荷。
type OrderConfirmedPayload struct {
	OrderID   uint64         `json:"order_id"`
	OrderNo   string         `json:"order_no"`
	UserID    uint64         `json:"user_id"`
	OldStatus pb.OrderStatus `json:"old_status"`
	NewStatus pb.OrderStatus `json:"new_status"`
	Confirmed time.Time      `json:"confirmed_at"`
	Log       *OrderEventLog `json:"log"`
}

// OrderRefundRequestedPayload 定义订单退款申请事件载荷。
type OrderRefundRequestedPayload struct {
	OrderID      uint64         `json:"order_id"`
	OrderNo      string         `json:"order_no"`
	UserID       uint64         `json:"user_id"`
	OldStatus    pb.OrderStatus `json:"old_status"`
	NewStatus    pb.OrderStatus `json:"new_status"`
	RefundAmount int64          `json:"refund_amount"`
	RefundReason string         `json:"refund_reason"`
	RequestedAt  time.Time      `json:"requested_at"`
	Log          *OrderEventLog `json:"log"`
}

// OrderRefundApprovedPayload 定义订单退款完成事件载荷。
type OrderRefundApprovedPayload struct {
	OrderID    uint64         `json:"order_id"`
	OrderNo    string         `json:"order_no"`
	UserID     uint64         `json:"user_id"`
	OldStatus  pb.OrderStatus `json:"old_status"`
	NewStatus  pb.OrderStatus `json:"new_status"`
	RefundedAt time.Time      `json:"refunded_at"`
	Log        *OrderEventLog `json:"log"`
}

// RebuildOrderFromEvents 基于事件流重建订单聚合状态。
func RebuildOrderFromEvents(events []eventsourcing.DomainEvent) (*Order, error) {
	if len(events) == 0 {
		return nil, nil
	}

	order := &Order{}
	for _, evt := range events {
		if err := ApplyOrderEvent(order, evt); err != nil {
			return nil, err
		}
	}

	order.initFSM()
	return order, nil
}

// ApplyOrderEvent 将事件应用到订单实体。
func ApplyOrderEvent(order *Order, event eventsourcing.DomainEvent) error {
	switch event.EventType() {
	case OrderEventTypeCreated:
		var payload OrderCreatedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.ID = uint(payload.OrderID)
		order.OrderNo = payload.OrderNo
		order.UserID = payload.UserID
		order.Status = payload.Status
		order.TotalAmount = payload.TotalAmount
		order.ActualAmount = payload.ActualAmount
		order.ShippingFee = payload.ShippingFee
		order.DiscountAmount = payload.DiscountAmount
		order.PaymentMethod = payload.PaymentMethod
		order.PaymentTransactionID = payload.PaymentTransactionID
		order.Remark = payload.Remark
		order.TrackingNumber = payload.TrackingNumber
		order.LogisticsCompany = payload.LogisticsCompany
		order.RefundAmount = payload.RefundAmount
		order.RefundReason = payload.RefundReason
		order.ShippingAddress = payload.ShippingAddress
		order.Items = payload.Items
		order.CreatedAt = payload.CreatedAt
		order.Logs = []*OrderLog{}
		appendEventLog(order, payload.OrderID, payload.InitLog)
	case OrderEventTypePaid:
		var payload OrderPaidPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.PaymentMethod = payload.PaymentMethod
		order.PaymentTransactionID = payload.PaymentTransactionID
		order.PaidAt = &payload.PaidAt
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeShipped:
		var payload OrderShippedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.TrackingNumber = payload.TrackingNumber
		order.LogisticsCompany = payload.LogisticsCompany
		order.ShippedAt = &payload.ShippedAt
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeDelivered:
		var payload OrderDeliveredPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.DeliveredAt = &payload.DeliveredAt
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeCompleted:
		var payload OrderCompletedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.CompletedAt = &payload.CompletedAt
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeCancelled:
		var payload OrderCancelledPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.CancelledAt = &payload.CancelledAt
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeConfirmed:
		var payload OrderConfirmedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeRefundRequested:
		var payload OrderRefundRequestedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		order.RefundAmount = payload.RefundAmount
		order.RefundReason = payload.RefundReason
		appendEventLog(order, payload.OrderID, payload.Log)
	case OrderEventTypeRefundApproved:
		var payload OrderRefundApprovedPayload
		if err := decodeEventData(event, &payload); err != nil {
			return err
		}
		order.Status = payload.NewStatus
		appendEventLog(order, payload.OrderID, payload.Log)
	default:
		return fmt.Errorf("unknown order event type: %s", event.EventType())
	}

	order.Version = event.Version()
	return nil
}

// decodeEventData 将事件载荷解码为指定结构体。
func decodeEventData(event eventsourcing.DomainEvent, out any) error {
	if base, ok := event.(*eventsourcing.BaseEvent); ok {
		if base.Data == nil {
			return nil
		}
		raw, err := json.Marshal(base.Data)
		if err != nil {
			return fmt.Errorf("marshal event data failed: %w", err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("unmarshal event data failed: %w", err)
		}
		return nil
	}

	raw, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal event failed: %w", err)
	}
	return nil
}

// appendEventLog 将事件日志追加到订单日志列表。
func appendEventLog(order *Order, orderID uint64, payload *OrderEventLog) {
	if payload == nil {
		return
	}

	log := &OrderLog{
		OrderID:   orderID,
		Operator:  payload.Operator,
		Action:    payload.Action,
		OldStatus: payload.OldStatus,
		NewStatus: payload.NewStatus,
		Remark:    payload.Remark,
	}
	log.CreatedAt = payload.LoggedAt
	log.UpdatedAt = payload.LoggedAt

	order.Logs = append(order.Logs, log)
}
