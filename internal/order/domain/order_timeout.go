package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// TimeoutPolicyType 超时策略类型
type TimeoutPolicyType string

const (
	TimeoutPayment  TimeoutPolicyType = "PAYMENT"  // 支付超时
	TimeoutShipment TimeoutPolicyType = "SHIPMENT" // 发货超时
	TimeoutDelivery TimeoutPolicyType = "DELIVERY" // 配送超时
	TimeoutConfirm  TimeoutPolicyType = "CONFIRM"  // 确认收货超时
	TimeoutPresale  TimeoutPolicyType = "PRESALE"  // 预售尾款支付超时
	TimeoutRefund   TimeoutPolicyType = "REFUND"   // 退款处理超时
	TimeoutReview   TimeoutPolicyType = "REVIEW"   // 审核超时
)

// TimeoutAction 超时动作
type TimeoutAction string

const (
	ActionAutoCancel   TimeoutAction = "AUTO_CANCEL"   // 自动取消
	ActionAutoComplete TimeoutAction = "AUTO_COMPLETE" // 自动完成
	ActionAutoRefund   TimeoutAction = "AUTO_REFUND"   // 自动退款
	ActionAutoConfirm  TimeoutAction = "AUTO_CONFIRM"  // 自动确认收货
	ActionNotify       TimeoutAction = "NOTIFY"        // 发送通知
	ActionEscalate     TimeoutAction = "ESCALATE"      // 升级处理
)

// TimeoutPolicy 超时策略
type TimeoutPolicy struct {
	ID               uint              `json:"id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	PolicyName       string            `json:"policy_name"`
	PolicyType       TimeoutPolicyType `json:"policy_type"`
	OrderType        pb.OrderType      `json:"order_type"`
	Enabled          bool              `json:"enabled"`
	Priority         int               `json:"priority"`
	TimeoutDuration  time.Duration     `json:"timeout_duration"`
	Action           TimeoutAction     `json:"action"`
	ActionParams     string            `json:"action_params"`     // JSON格式的动作参数
	Notification     bool              `json:"notification"`      // 是否发送通知
	NotificationTime time.Duration     `json:"notification_time"` // 通知提前时间
	RetryCount       int               `json:"retry_count"`       // 重试次数
	RetryInterval    time.Duration     `json:"retry_interval"`    // 重试间隔
	Conditions       string            `json:"conditions"`        // 条件表达式 (JSON)
	Description      string            `json:"description"`
	CreatorID        uint64            `json:"creator_id"`
}

// TimeoutTask 超时任务
type TimeoutTask struct {
	ID         uint              `json:"id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	TaskID     string            `json:"task_id"`
	OrderID    uint64            `json:"order_id"`
	OrderNo    string            `json:"order_no"`
	PolicyID   uint64            `json:"policy_id"`
	PolicyType TimeoutPolicyType `json:"policy_type"`
	Status     string            `json:"status"`      // PENDING, PROCESSING, COMPLETED, FAILED, CANCELLED
	ExecuteAt  time.Time         `json:"execute_at"`  // 计划执行时间
	ExecutedAt *time.Time        `json:"executed_at"` // 实际执行时间
	RetryCount int               `json:"retry_count"` // 已重试次数
	Error      string            `json:"error"`
	Result     string            `json:"result"` // 执行结果 (JSON)
}

// TimeoutManager 超时管理器
type TimeoutManager struct {
	orderRepo   OrderRepository
	policyRepo  TimeoutPolicyRepository
	taskRepo    TimeoutTaskRepository
	scheduler   TimeoutScheduler
	mu          sync.RWMutex
	policies    map[TimeoutPolicyType][]*TimeoutPolicy
	activeTasks map[string]*TimeoutTask
}

// NewTimeoutManager 创建超时管理器
func NewTimeoutManager(orderRepo OrderRepository, policyRepo TimeoutPolicyRepository, taskRepo TimeoutTaskRepository) *TimeoutManager {
	return &TimeoutManager{
		orderRepo:   orderRepo,
		policyRepo:  policyRepo,
		taskRepo:    taskRepo,
		policies:    make(map[TimeoutPolicyType][]*TimeoutPolicy),
		activeTasks: make(map[string]*TimeoutTask),
	}
}

// Initialize 初始化超时管理器
func (m *TimeoutManager) Initialize(ctx context.Context) error {
	// 加载所有启用的策略
	policies, err := m.policyRepo.GetEnabledPolicies(ctx)
	if err != nil {
		return fmt.Errorf("failed to load timeout policies: %w", err)
	}

	// 按策略类型分组
	m.mu.Lock()
	for _, policy := range policies {
		m.policies[policy.PolicyType] = append(m.policies[policy.PolicyType], policy)
	}
	m.mu.Unlock()

	// 加载待处理的超时任务
	pendingTasks, err := m.taskRepo.GetPendingTasks(ctx, 1000)
	if err != nil {
		return fmt.Errorf("failed to load pending tasks: %w", err)
	}

	// 重新调度待处理任务
	for _, task := range pendingTasks {
		if task.ExecuteAt.After(time.Now()) {
			m.scheduleTask(task)
		} else {
			// 已过期的任务立即执行
			go m.executeTask(task)
		}
	}

	return nil
}

// CreateTimeoutTask 创建超时任务
func (m *TimeoutManager) CreateTimeoutTask(ctx context.Context, order *Order, policyType TimeoutPolicyType) (*TimeoutTask, error) {
	// 查找适用的策略
	policy, err := m.findApplicablePolicy(order, policyType)
	if err != nil {
		return nil, fmt.Errorf("failed to find applicable policy: %w", err)
	}

	if policy == nil {
		return nil, nil // 没有适用的策略
	}

	// 计算执行时间
	executeAt := m.calculateExecuteTime(order, policy)

	// 创建超时任务
	task := &TimeoutTask{
		TaskID:     generateTaskID(),
		OrderID:    uint64(order.ID),
		OrderNo:    order.OrderNo,
		PolicyID:   uint64(policy.ID),
		PolicyType: policyType,
		Status:     "PENDING",
		ExecuteAt:  executeAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存任务
	err = m.taskRepo.SaveTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to save timeout task: %w", err)
	}

	// 调度任务
	m.scheduleTask(task)

	return task, nil
}

// CancelTimeoutTask 取消超时任务
func (m *TimeoutManager) CancelTimeoutTask(ctx context.Context, orderID uint64, policyType TimeoutPolicyType) error {
	// 查找任务
	task, err := m.taskRepo.GetTaskByOrderAndType(ctx, orderID, policyType, "PENDING")
	if err != nil {
		return fmt.Errorf("failed to find timeout task: %w", err)
	}

	if task == nil {
		return nil // 没有找到任务
	}

	// 取消任务
	task.Status = "CANCELLED"
	task.UpdatedAt = time.Now()

	err = m.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update timeout task: %w", err)
	}

	// 从调度器中移除
	m.cancelTask(task)

	return nil
}

// ProcessTimeout 处理超时
func (m *TimeoutManager) ProcessTimeout(ctx context.Context, task *TimeoutTask) error {
	// 获取订单
	order, err := m.orderRepo.FindByID(ctx, 0, uint64(task.OrderID))
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// 检查订单状态是否仍然符合条件
	if !m.isOrderEligibleForTimeout(order, task.PolicyType) {
		// 订单状态已改变，取消任务
		task.Status = "CANCELLED"
		task.Error = "Order status changed"
		task.UpdatedAt = time.Now()
		m.taskRepo.UpdateTask(ctx, task)
		return nil
	}

	// 获取策略
	policy, err := m.policyRepo.GetPolicyByID(ctx, uint64(task.PolicyID))
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// 执行超时动作
	result, err := m.executeTimeoutAction(ctx, order, policy)
	if err != nil {
		// 处理失败，检查是否需要重试
		if task.RetryCount < policy.RetryCount {
			// 重试
			task.RetryCount++
			task.ExecuteAt = time.Now().Add(policy.RetryInterval)
			task.UpdatedAt = time.Now()

			err = m.taskRepo.UpdateTask(ctx, task)
			if err != nil {
				return fmt.Errorf("failed to update retry task: %w", err)
			}

			// 重新调度
			m.scheduleTask(task)
			return nil
		} else {
			// 重试次数用尽，标记为失败
			task.Status = "FAILED"
			task.Error = err.Error()
		}
	} else {
		// 执行成功
		task.Status = "COMPLETED"
		task.Result = result
		now := time.Now()
		task.ExecutedAt = &now
	}

	task.UpdatedAt = time.Now()
	err = m.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// executeTimeoutAction 执行超时动作
func (m *TimeoutManager) executeTimeoutAction(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	switch policy.Action {
	case ActionAutoCancel:
		return m.executeAutoCancel(ctx, order, policy)
	case ActionAutoComplete:
		return m.executeAutoComplete(ctx, order, policy)
	case ActionAutoRefund:
		return m.executeAutoRefund(ctx, order, policy)
	case ActionAutoConfirm:
		return m.executeAutoConfirm(ctx, order, policy)
	case ActionNotify:
		return m.executeNotify(ctx, order, policy)
	case ActionEscalate:
		return m.executeEscalate(ctx, order, policy)
	default:
		return "", fmt.Errorf("unknown timeout action: %s", policy.Action)
	}
}

// executeAutoCancel 执行自动取消
func (m *TimeoutManager) executeAutoCancel(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	// 解析动作参数
	var params map[string]interface{}
	if policy.ActionParams != "" {
		err := json.Unmarshal([]byte(policy.ActionParams), &params)
		if err != nil {
			return "", fmt.Errorf("failed to parse action params: %w", err)
		}
	}

	// 获取取消原因
	reason := "系统自动取消（超时未支付）"
	if r, ok := params["reason"].(string); ok {
		reason = r
	}

	// 执行取消
	err := order.Cancel(ctx, "System", reason)
	if err != nil {
		return "", fmt.Errorf("failed to cancel order: %w", err)
	}

	// 更新订单
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	// 记录结果
	result := map[string]interface{}{
		"action":    "auto_cancel",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"reason":    reason,
		"timestamp": time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// executeAutoComplete 执行自动完成
func (m *TimeoutManager) executeAutoComplete(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	// 自动确认收货并完成订单
	err := order.Deliver(ctx, "System")
	if err != nil {
		return "", fmt.Errorf("failed to deliver order: %w", err)
	}

	err = order.Complete(ctx, "System")
	if err != nil {
		return "", fmt.Errorf("failed to complete order: %w", err)
	}

	// 更新订单
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	// 记录结果
	result := map[string]interface{}{
		"action":    "auto_complete",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"timestamp": time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// executeAutoRefund 执行自动退款
func (m *TimeoutManager) executeAutoRefund(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	// 解析动作参数
	var params map[string]interface{}
	if policy.ActionParams != "" {
		if err := json.Unmarshal([]byte(policy.ActionParams), &params); err != nil {
			return "", fmt.Errorf("failed to parse action params: %w", err)
		}
	}

	reason := "系统自动退款（超时未处理）"
	if r, ok := params["reason"].(string); ok && r != "" {
		reason = r
	}
	refundAmount := order.ActualAmount
	if v, ok := toInt64(params["refund_amount"]); ok && v > 0 {
		if v < refundAmount {
			refundAmount = v
		}
	}

	switch order.Status {
	case pb.OrderStatus_REFUND_REQUESTED:
		if err := order.ApproveRefund(ctx, "System"); err != nil {
			return "", fmt.Errorf("failed to approve refund: %w", err)
		}
	case pb.OrderStatus_PAID, pb.OrderStatus_SHIPPED, pb.OrderStatus_DELIVERED:
		if err := order.RequestRefund(ctx, "System", reason); err != nil {
			return "", fmt.Errorf("failed to request refund: %w", err)
		}
		if err := order.ApproveRefund(ctx, "System"); err != nil {
			return "", fmt.Errorf("failed to approve refund: %w", err)
		}
	case pb.OrderStatus_CANCELLED:
		// 订单已取消但未进入退款状态机，补偿标记退款结果。
		if order.PaymentStatus != pb.PaymentStatus_UNPAID && order.PaidAt != nil {
			order.PaymentStatus = pb.PaymentStatus_REFUND_SUCCESS
		}
		order.AddLog("System", "AUTO_REFUND", order.Status.String(), order.Status.String(), reason)
	default:
		return "", fmt.Errorf("order status %s not eligible for auto refund", order.Status.String())
	}

	order.RefundReason = reason
	order.RefundAmount = refundAmount
	order.UpdatedAt = time.Now()

	if err := m.orderRepo.Update(ctx, order); err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	result := map[string]interface{}{
		"action":        "auto_refund",
		"order_id":      order.ID,
		"order_no":      order.OrderNo,
		"refund_amount": refundAmount,
		"reason":        reason,
		"status":        order.Status.String(),
		"timestamp":     time.Now(),
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// executeAutoConfirm 执行自动确认收货
func (m *TimeoutManager) executeAutoConfirm(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	err := order.Deliver(ctx, "System")
	if err != nil {
		return "", fmt.Errorf("failed to confirm delivery: %w", err)
	}

	// 更新订单
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	// 记录结果
	result := map[string]interface{}{
		"action":    "auto_confirm",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"timestamp": time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// executeNotify 执行通知
func (m *TimeoutManager) executeNotify(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	var params map[string]interface{}
	if policy.ActionParams != "" {
		if err := json.Unmarshal([]byte(policy.ActionParams), &params); err != nil {
			return "", fmt.Errorf("failed to parse action params: %w", err)
		}
	}

	message := fmt.Sprintf("订单超时提醒：订单号 %s，策略 %s", order.OrderNo, policy.PolicyName)
	if m, ok := params["message"].(string); ok && m != "" {
		message = m
	}
	channels := toStringSlice(params["channels"])
	if len(channels) == 0 {
		channels = []string{"IN_APP"}
	}

	// 当前阶段先记录通知行为，后续可接入站内信/短信/邮件通道。
	order.AddLog("System", "TIMEOUT_NOTIFY", order.Status.String(), order.Status.String(), message)
	order.UpdatedAt = time.Now()
	if err := m.orderRepo.Update(ctx, order); err != nil {
		return "", fmt.Errorf("failed to update order notify log: %w", err)
	}

	result := map[string]interface{}{
		"action":    "notify",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"message":   message,
		"channels":  channels,
		"timestamp": time.Now(),
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// executeEscalate 执行升级处理
func (m *TimeoutManager) executeEscalate(ctx context.Context, order *Order, policy *TimeoutPolicy) (string, error) {
	var params map[string]interface{}
	if policy.ActionParams != "" {
		if err := json.Unmarshal([]byte(policy.ActionParams), &params); err != nil {
			return "", fmt.Errorf("failed to parse action params: %w", err)
		}
	}

	level := "P2"
	if v, ok := params["level"].(string); ok && v != "" {
		level = strings.ToUpper(v)
	}
	queue := "customer_service"
	if v, ok := params["queue"].(string); ok && v != "" {
		queue = v
	}
	assignee := "AUTO_ROUTER"
	if v, ok := params["assignee"].(string); ok && v != "" {
		assignee = v
	}
	reason := "订单超时需人工介入"
	if v, ok := params["reason"].(string); ok && v != "" {
		reason = v
	}

	escalationNote := fmt.Sprintf("[%s] queue=%s assignee=%s reason=%s", level, queue, assignee, reason)
	if order.AdminNotes == "" {
		order.AdminNotes = escalationNote
	} else {
		order.AdminNotes = order.AdminNotes + "\n" + escalationNote
	}
	order.AddLog("System", "TIMEOUT_ESCALATE", order.Status.String(), order.Status.String(), escalationNote)
	order.UpdatedAt = time.Now()
	if err := m.orderRepo.Update(ctx, order); err != nil {
		return "", fmt.Errorf("failed to update order escalation info: %w", err)
	}

	result := map[string]interface{}{
		"action":    "escalate",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"level":     level,
		"queue":     queue,
		"assignee":  assignee,
		"reason":    reason,
		"timestamp": time.Now(),
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// Helper methods

func (m *TimeoutManager) findApplicablePolicy(order *Order, policyType TimeoutPolicyType) (*TimeoutPolicy, error) {
	m.mu.RLock()
	policies, exists := m.policies[policyType]
	m.mu.RUnlock()

	if !exists || len(policies) == 0 {
		return nil, nil
	}

	// 按优先级排序，找到第一个适用的策略（数值越小优先级越高）。
	sorted := make([]*TimeoutPolicy, 0, len(policies))
	sorted = append(sorted, policies...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	for _, policy := range sorted {
		if !policy.Enabled {
			continue
		}

		// 检查订单类型
		if policy.OrderType != pb.OrderType_ORDER_TYPE_UNSPECIFIED && policy.OrderType != order.OrderType {
			continue
		}

		// 检查其他条件
		if policy.Conditions != "" {
			if !evaluatePolicyConditions(order, policy.Conditions) {
				continue
			}
		}

		return policy, nil
	}

	return nil, nil
}

func (m *TimeoutManager) calculateExecuteTime(order *Order, policy *TimeoutPolicy) time.Time {
	var baseTime time.Time

	switch policy.PolicyType {
	case TimeoutPayment:
		baseTime = order.CreatedAt
	case TimeoutShipment:
		if order.PaidAt != nil {
			baseTime = *order.PaidAt
		} else {
			baseTime = order.CreatedAt
		}
	case TimeoutDelivery:
		if order.ShippedAt != nil {
			baseTime = *order.ShippedAt
		} else {
			return time.Now().Add(policy.TimeoutDuration) // 使用相对时间
		}
	case TimeoutConfirm:
		if order.DeliveredAt != nil {
			baseTime = *order.DeliveredAt
		} else {
			return time.Now().Add(policy.TimeoutDuration)
		}
	case TimeoutPresale:
		// 预售订单尾款支付超时
		baseTime = order.CreatedAt
	default:
		baseTime = order.CreatedAt
	}

	return baseTime.Add(policy.TimeoutDuration)
}

func (m *TimeoutManager) isOrderEligibleForTimeout(order *Order, policyType TimeoutPolicyType) bool {
	switch policyType {
	case TimeoutPayment:
		return order.Status == pb.OrderStatus_PENDING_PAYMENT
	case TimeoutShipment:
		return order.Status == pb.OrderStatus_PAID && order.ShippingStatus == pb.ShippingStatus_PENDING_SHIPMENT
	case TimeoutDelivery:
		return order.Status == pb.OrderStatus_SHIPPED && order.ShippingStatus == pb.ShippingStatus_SHIPPING_SHIPPED
	case TimeoutConfirm:
		return order.Status == pb.OrderStatus_DELIVERED
	case TimeoutPresale:
		return order.OrderType == pb.OrderType_PRE_SALE && order.Status == pb.OrderStatus_PENDING_BALANCE
	default:
		return true
	}
}

func (m *TimeoutManager) scheduleTask(task *TimeoutTask) {
	m.mu.Lock()
	m.activeTasks[task.TaskID] = task
	m.mu.Unlock()

	delay := time.Until(task.ExecuteAt)
	if delay < 0 {
		delay = 0
	}
	if m.scheduler != nil {
		if err := m.scheduler.ScheduleTimeout(task.TaskID, delay, m.executeTaskCallback); err == nil {
			return
		}
	}
	// 无调度器时回退到本地计时器，依然通过 callback 判断任务是否已取消。
	time.AfterFunc(delay, func() {
		m.executeTaskCallback(task.TaskID)
	})
}

func (m *TimeoutManager) cancelTask(task *TimeoutTask) {
	m.mu.Lock()
	delete(m.activeTasks, task.TaskID)
	m.mu.Unlock()

	// 对支持取消能力的调度器执行取消。
	if m.scheduler != nil {
		if cancellable, ok := m.scheduler.(interface{ CancelTimeout(orderID string) error }); ok {
			_ = cancellable.CancelTimeout(task.TaskID)
		}
	}
}

func (m *TimeoutManager) executeTask(task *TimeoutTask) {
	ctx := context.Background()
	err := m.ProcessTimeout(ctx, task)
	if err != nil {
		fmt.Printf("Failed to process timeout task %s: %v\n", task.TaskID, err)
	}
}

func (m *TimeoutManager) executeTaskCallback(taskID string) {
	m.mu.RLock()
	task, exists := m.activeTasks[taskID]
	m.mu.RUnlock()
	if !exists {
		return
	}

	m.executeTask(task)

	if task.Status == "COMPLETED" || task.Status == "FAILED" || task.Status == "CANCELLED" {
		m.mu.Lock()
		delete(m.activeTasks, taskID)
		m.mu.Unlock()
	}
}

func generateTaskID() string {
	return fmt.Sprintf("TIMEOUT_%d", time.Now().UnixNano())
}

func evaluatePolicyConditions(order *Order, conditions string) bool {
	var condMap map[string]interface{}
	if err := json.Unmarshal([]byte(conditions), &condMap); err != nil {
		// 条件表达式非法时不匹配，避免误执行策略。
		return false
	}
	if len(condMap) == 0 {
		return true
	}

	if minAmount, ok := toInt64(condMap["min_amount"]); ok && order.ActualAmount < minAmount {
		return false
	}
	if maxAmount, ok := toInt64(condMap["max_amount"]); ok && order.ActualAmount > maxAmount {
		return false
	}

	if statuses := toStringSlice(condMap["status_in"]); len(statuses) > 0 && !containsString(statuses, order.Status.String()) {
		return false
	}
	if paymentStatuses := toStringSlice(condMap["payment_status_in"]); len(paymentStatuses) > 0 &&
		!containsString(paymentStatuses, order.PaymentStatus.String()) {
		return false
	}
	if shippingStatuses := toStringSlice(condMap["shipping_status_in"]); len(shippingStatuses) > 0 &&
		!containsString(shippingStatuses, order.ShippingStatus.String()) {
		return false
	}
	if methods := toStringSlice(condMap["payment_method_in"]); len(methods) > 0 &&
		!containsString(methods, order.PaymentMethod) {
		return false
	}

	if hasAddr, ok := toBool(condMap["has_shipping_address"]); ok {
		if hasAddr && order.ShippingAddress == nil {
			return false
		}
		if !hasAddr && order.ShippingAddress != nil {
			return false
		}
	}

	return true
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case float32:
		return int64(n), true
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case json.Number:
		parsed, err := n.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func toBool(v interface{}) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		return parsed, err == nil
	default:
		return false, false
	}
}

func toStringSlice(v interface{}) []string {
	switch values := v.(type) {
	case []string:
		return values
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, raw := range values {
			if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if values == "" {
			return nil
		}
		parts := strings.Split(values, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	default:
		return nil
	}
}

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// TimeoutPolicyRepository 超时策略仓储接口
type TimeoutPolicyRepository interface {
	SavePolicy(ctx context.Context, policy *TimeoutPolicy) error
	GetPolicyByID(ctx context.Context, id uint64) (*TimeoutPolicy, error)
	GetPoliciesByType(ctx context.Context, policyType TimeoutPolicyType, enabledOnly bool) ([]*TimeoutPolicy, error)
	GetEnabledPolicies(ctx context.Context) ([]*TimeoutPolicy, error)
	UpdatePolicy(ctx context.Context, policy *TimeoutPolicy) error
	DeletePolicy(ctx context.Context, id uint64) error
}

// TimeoutTaskRepository 超时任务仓储接口
type TimeoutTaskRepository interface {
	SaveTask(ctx context.Context, task *TimeoutTask) error
	GetTaskByID(ctx context.Context, taskID string) (*TimeoutTask, error)
	GetTaskByOrderAndType(ctx context.Context, orderID uint64, policyType TimeoutPolicyType, status string) (*TimeoutTask, error)
	GetPendingTasks(ctx context.Context, limit int) ([]*TimeoutTask, error)
	GetTasksByStatus(ctx context.Context, status string, page, pageSize int) ([]*TimeoutTask, int64, error)
	UpdateTask(ctx context.Context, task *TimeoutTask) error
	DeleteTask(ctx context.Context, taskID string) error
	CleanupExpiredTasks(ctx context.Context, before time.Time) error
}
