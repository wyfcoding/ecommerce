package domain

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

type ShardStrategy int

const (
	ShardStrategyUserID    ShardStrategy = 1
	ShardStrategyOrderID   ShardStrategy = 2
	ShardStrategyMerchantID ShardStrategy = 3
	ShardStrategyHash      ShardStrategy = 4
)

type ShardConfig struct {
	TotalShards    int
	Strategy       ShardStrategy
	ShardKeyPrefix string
}

func NewShardConfig(totalShards int, strategy ShardStrategy) *ShardConfig {
	return &ShardConfig{
		TotalShards: totalShards,
		Strategy:    strategy,
	}
}

func (c *ShardConfig) GetShardIndex(key uint64) int {
	return int(key % uint64(c.TotalShards))
}

func (c *ShardConfig) GetShardIndexByString(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % c.TotalShards
}

func (c *ShardConfig) GetShardName(index int) string {
	return fmt.Sprintf("shard_%d", index)
}

func (o *Order) GetShardKey() uint64 {
	return o.UserID
}

func (o *Order) GetShardIndex(config *ShardConfig) int {
	switch config.Strategy {
	case ShardStrategyUserID:
		return config.GetShardIndex(o.UserID)
	case ShardStrategyOrderID:
		return config.GetShardIndex(uint64(o.ID))
	default:
		return config.GetShardIndex(o.UserID)
	}
}

type OrderStateMachine struct {
	order       *Order
	transitions map[string][]Transition
}

type Transition struct {
	From   string
	Event  string
	To     string
	Guard  func(*Order) bool
	Action func(*Order, context.Context) error
}

func NewOrderStateMachine(order *Order) *OrderStateMachine {
	sm := &OrderStateMachine{
		order:       order,
		transitions: make(map[string][]Transition),
	}
	sm.defineTransitions()
	return sm
}

func (sm *OrderStateMachine) defineTransitions() {
	sm.addTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "PAY", pb.OrderStatus_PAID.String(),
		nil, sm.onPaymentSuccess)
	
	sm.addTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "CANCEL", pb.OrderStatus_CANCELLED.String(),
		sm.canCancel, sm.onCancel)
	
	sm.addTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "TIMEOUT", pb.OrderStatus_CANCELLED.String(),
		nil, sm.onTimeout)
	
	sm.addTransition(pb.OrderStatus_ALLOCATING.String(), "CONFIRM", pb.OrderStatus_PENDING_PAYMENT.String(),
		nil, nil)
	
	sm.addTransition(pb.OrderStatus_ALLOCATING.String(), "CANCEL", pb.OrderStatus_CANCELLED.String(),
		nil, sm.onCancel)
	
	sm.addTransition(pb.OrderStatus_PAID.String(), "SHIP", pb.OrderStatus_SHIPPED.String(),
		sm.canShip, sm.onShip)
	
	sm.addTransition(pb.OrderStatus_PAID.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String(),
		sm.canRequestRefund, sm.onRefundRequest)
	
	sm.addTransition(pb.OrderStatus_PAID.String(), "PROCESS_VIRTUAL", pb.OrderStatus_COMPLETED.String(),
		sm.isVirtualOrder, sm.onVirtualComplete)
	
	sm.addTransition(pb.OrderStatus_SHIPPED.String(), "DELIVER", pb.OrderStatus_DELIVERED.String(),
		nil, sm.onDeliver)
	
	sm.addTransition(pb.OrderStatus_SHIPPED.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String(),
		nil, sm.onRefundRequest)
	
	sm.addTransition(pb.OrderStatus_DELIVERED.String(), "COMPLETE", pb.OrderStatus_COMPLETED.String(),
		nil, sm.onComplete)
	
	sm.addTransition(pb.OrderStatus_DELIVERED.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String(),
		sm.canRequestRefundAfterDelivery, sm.onRefundRequest)
	
	sm.addTransition(pb.OrderStatus_REFUND_REQUESTED.String(), "REFUND_APPROVE", pb.OrderStatus_REFUNDED.String(),
		nil, sm.onRefundApprove)
	
	sm.addTransition(pb.OrderStatus_REFUND_REQUESTED.String(), "REFUND_REJECT", pb.OrderStatus_PAID.String(),
		nil, sm.onRefundReject)
	
	sm.addTransition(pb.OrderStatus_PENDING_BALANCE.String(), "PAY_BALANCE", pb.OrderStatus_PAID.String(),
		nil, sm.onBalancePayment)
	
	sm.addTransition(pb.OrderStatus_PENDING_BALANCE.String(), "CANCEL", pb.OrderStatus_CANCELLED.String(),
		nil, sm.onCancel)
}

func (sm *OrderStateMachine) addTransition(from, event, to string, guard func(*Order) bool, action func(*Order, context.Context) error) {
	sm.transitions[from] = append(sm.transitions[from], Transition{
		From:  from,
		Event: event,
		To:    to,
		Guard: guard,
		Action: action,
	})
}

func (sm *OrderStateMachine) CanTrigger(event string) bool {
	currentState := sm.order.Status.String()
	transitions, exists := sm.transitions[currentState]
	if !exists {
		return false
	}
	
	for _, t := range transitions {
		if t.Event == event {
			if t.Guard != nil {
				return t.Guard(sm.order)
			}
			return true
		}
	}
	return false
}

func (sm *OrderStateMachine) Trigger(ctx context.Context, event, operator, remark string) error {
	currentState := sm.order.Status.String()
	transitions, exists := sm.transitions[currentState]
	if !exists {
		return fmt.Errorf("no transitions from state %s", currentState)
	}
	
	var targetTransition *Transition
	for _, t := range transitions {
		if t.Event == event {
			targetTransition = &t
			break
		}
	}
	
	if targetTransition == nil {
		return fmt.Errorf("invalid event %s from state %s", event, currentState)
	}
	
	if targetTransition.Guard != nil && !targetTransition.Guard(sm.order) {
		return fmt.Errorf("guard condition failed for event %s", event)
	}
	
	oldStatus := sm.order.Status
	
	for i := 0; i < len(pb.OrderStatus_name); i++ {
		st := pb.OrderStatus(i)
		if st.String() == targetTransition.To {
			sm.order.Status = st
			break
		}
	}
	
	if targetTransition.Action != nil {
		if err := targetTransition.Action(sm.order, ctx); err != nil {
			sm.order.Status = oldStatus
			return fmt.Errorf("action failed: %w", err)
		}
	}
	
	sm.order.AddLog(operator, event, oldStatus.String(), sm.order.Status.String(), remark)
	return nil
}

func (sm *OrderStateMachine) canCancel(order *Order) bool {
	return order.PaymentStatus == pb.PaymentStatus_UNPAID || 
		order.PaymentStatus == pb.PaymentStatus_PROCESSING
}

func (sm *OrderStateMachine) canShip(order *Order) bool {
	return order.PaymentStatus == pb.PaymentStatus_SUCCESS
}

func (sm *OrderStateMachine) canRequestRefund(order *Order) bool {
	return order.PaymentStatus == pb.PaymentStatus_SUCCESS
}

func (sm *OrderStateMachine) canRequestRefundAfterDelivery(order *Order) bool {
	if order.DeliveredAt == nil {
		return false
	}
	return time.Since(*order.DeliveredAt) <= 7*24*time.Hour
}

func (sm *OrderStateMachine) isVirtualOrder(order *Order) bool {
	for _, item := range order.Items {
		if item.ProductType != pb.ProductType_VIRTUAL {
			return false
		}
	}
	return true
}

func (sm *OrderStateMachine) onPaymentSuccess(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_SUCCESS
	now := time.Now()
	order.PaidAt = &now
	return nil
}

func (sm *OrderStateMachine) onCancel(order *Order, ctx context.Context) error {
	if order.PaidAt != nil {
		order.PaymentStatus = pb.PaymentStatus_REFUNDING
	} else {
		order.PaymentStatus = pb.PaymentStatus_UNPAID
	}
	order.ShippingStatus = pb.ShippingStatus_EXCEPTION
	now := time.Now()
	order.CancelledAt = &now
	return nil
}

func (sm *OrderStateMachine) onTimeout(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_UNPAID
	order.ShippingStatus = pb.ShippingStatus_EXCEPTION
	now := time.Now()
	order.CancelledAt = &now
	return nil
}

func (sm *OrderStateMachine) onShip(order *Order, ctx context.Context) error {
	order.ShippingStatus = pb.ShippingStatus_SHIPPING_SHIPPED
	now := time.Now()
	order.ShippedAt = &now
	return nil
}

func (sm *OrderStateMachine) onDeliver(order *Order, ctx context.Context) error {
	order.ShippingStatus = pb.ShippingStatus_SHIPPING_DELIVERED
	now := time.Now()
	order.DeliveredAt = &now
	return nil
}

func (sm *OrderStateMachine) onComplete(order *Order, ctx context.Context) error {
	now := time.Now()
	order.CompletedAt = &now
	return nil
}

func (sm *OrderStateMachine) onRefundRequest(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_REFUNDING
	order.ShippingStatus = pb.ShippingStatus_EXCEPTION
	return nil
}

func (sm *OrderStateMachine) onRefundApprove(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_REFUND_SUCCESS
	order.ShippingStatus = pb.ShippingStatus_EXCEPTION
	return nil
}

func (sm *OrderStateMachine) onRefundReject(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_SUCCESS
	return nil
}

func (sm *OrderStateMachine) onVirtualComplete(order *Order, ctx context.Context) error {
	now := time.Now()
	order.CompletedAt = &now
	order.ShippingStatus = pb.ShippingStatus_SHIPPING_DELIVERED
	return nil
}

func (sm *OrderStateMachine) onBalancePayment(order *Order, ctx context.Context) error {
	order.PaymentStatus = pb.PaymentStatus_SUCCESS
	now := time.Now()
	order.PaidAt = &now
	return nil
}

func (sm *OrderStateMachine) GetValidEvents() []string {
	currentState := sm.order.Status.String()
	transitions, exists := sm.transitions[currentState]
	if !exists {
		return nil
	}
	
	events := make([]string, 0, len(transitions))
	for _, t := range transitions {
		if t.Guard == nil || t.Guard(sm.order) {
			events = append(events, t.Event)
		}
	}
	return events
}

func (sm *OrderStateMachine) GetStatusHistory() []*OrderLog {
	return sm.order.Logs
}

type OrderStateMachineFactory struct{}

func NewOrderStateMachineFactory() *OrderStateMachineFactory {
	return &OrderStateMachineFactory{}
}

func (f *OrderStateMachineFactory) Create(order *Order) *OrderStateMachine {
	return NewOrderStateMachine(order)
}

type TimeoutSchedulerEnhanced struct {
	schedulers map[string]context.CancelFunc
	config     *TimeoutConfig
}

type TimeoutConfig struct {
	PaymentTimeout   time.Duration
	ConfirmTimeout   time.Duration
	ReceiveTimeout   time.Duration
	AutoReceiveDays  int
}

func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		PaymentTimeout:  30 * time.Minute,
		ConfirmTimeout:  24 * time.Hour,
		ReceiveTimeout:  15 * 24 * time.Hour,
		AutoReceiveDays: 7,
	}
}

func NewTimeoutSchedulerEnhanced(config *TimeoutConfig) *TimeoutSchedulerEnhanced {
	if config == nil {
		config = DefaultTimeoutConfig()
	}
	return &TimeoutSchedulerEnhanced{
		schedulers: make(map[string]context.CancelFunc),
		config:     config,
	}
}

func (s *TimeoutSchedulerEnhanced) SchedulePaymentTimeout(orderID string, callback func(orderID string)) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		timer := time.NewTimer(s.config.PaymentTimeout)
		defer timer.Stop()
		
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			callback(orderID)
		}
	}()
	
	s.schedulers[orderID] = cancel
	return cancel
}

func (s *TimeoutSchedulerEnhanced) ScheduleAutoReceive(orderID string, callback func(orderID string)) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	
	timeout := time.Duration(s.config.AutoReceiveDays) * 24 * time.Hour
	
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			callback(orderID)
		}
	}()
	
	s.schedulers[orderID+"_auto_receive"] = cancel
	return cancel
}

func (s *TimeoutSchedulerEnhanced) CancelTimeout(orderID string) {
	if cancel, exists := s.schedulers[orderID]; exists {
		cancel()
		delete(s.schedulers, orderID)
	}
	if cancel, exists := s.schedulers[orderID+"_auto_receive"]; exists {
		cancel()
		delete(s.schedulers, orderID+"_auto_receive")
	}
}

func (s *TimeoutSchedulerEnhanced) Stop() {
	for _, cancel := range s.schedulers {
		cancel()
	}
	s.schedulers = make(map[string]context.CancelFunc)
}

type OrderStateVisitor interface {
	VisitPendingPayment(order *Order) error
	VisitPaid(order *Order) error
	VisitShipped(order *Order) error
	VisitDelivered(order *Order) error
	VisitCompleted(order *Order) error
	VisitCancelled(order *Order) error
	VisitRefundRequested(order *Order) error
	VisitRefunded(order *Order) error
}

func (o *Order) Accept(visitor OrderStateVisitor) error {
	switch o.Status {
	case pb.OrderStatus_PENDING_PAYMENT:
		return visitor.VisitPendingPayment(o)
	case pb.OrderStatus_PAID:
		return visitor.VisitPaid(o)
	case pb.OrderStatus_SHIPPED:
		return visitor.VisitShipped(o)
	case pb.OrderStatus_DELIVERED:
		return visitor.VisitDelivered(o)
	case pb.OrderStatus_COMPLETED:
		return visitor.VisitCompleted(o)
	case pb.OrderStatus_CANCELLED:
		return visitor.VisitCancelled(o)
	case pb.OrderStatus_REFUND_REQUESTED:
		return visitor.VisitRefundRequested(o)
	case pb.OrderStatus_REFUNDED:
		return visitor.VisitRefunded(o)
	default:
		return fmt.Errorf("unknown order status: %s", o.Status)
	}
}

type OrderStateNotifier struct {
	handlers map[pb.OrderStatus][]func(*Order)
}

func NewOrderStateNotifier() *OrderStateNotifier {
	return &OrderStateNotifier{
		handlers: make(map[pb.OrderStatus][]func(*Order)),
	}
}

func (n *OrderStateNotifier) RegisterHandler(status pb.OrderStatus, handler func(*Order)) {
	n.handlers[status] = append(n.handlers[status], handler)
}

func (n *OrderStateNotifier) Notify(order *Order) {
	if handlers, exists := n.handlers[order.Status]; exists {
		for _, handler := range handlers {
			go handler(order)
		}
	}
}
