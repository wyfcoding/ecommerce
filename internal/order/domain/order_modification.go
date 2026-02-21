package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// ModificationType 修改类型
type ModificationType string

const (
	ModifyShippingAddress ModificationType = "SHIPPING_ADDRESS" // 修改收货地址
	ModifyRemark          ModificationType = "REMARK"           // 修改备注
	ModifyContactInfo     ModificationType = "CONTACT_INFO"     // 修改联系方式
	ModifyProduct         ModificationType = "PRODUCT"          // 修改商品
	ModifyQuantity        ModificationType = "QUANTITY"         // 修改数量
	ModifyPrice           ModificationType = "PRICE"            // 修改价格
	ModifyDiscount        ModificationType = "DISCOUNT"         // 修改折扣
	ModifyShippingMethod  ModificationType = "SHIPPING_METHOD"  // 修改配送方式
)

// ModificationStatus 修改状态
type ModificationStatus string

const (
	ModificationPending   ModificationStatus = "PENDING"   // 待处理
	ModificationApproved  ModificationStatus = "APPROVED"  // 已批准
	ModificationRejected  ModificationStatus = "REJECTED"  // 已拒绝
	ModificationCompleted ModificationStatus = "COMPLETED" // 已完成
	ModificationCancelled ModificationStatus = "CANCELLED" // 已取消
)

// ModificationRequest 修改请求
type ModificationRequest struct {
	ID               uint64             `json:"id"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	RequestNo        string             `json:"request_no"`
	OrderID          uint64             `json:"order_id"`
	OrderNo          string             `json:"order_no"`
	UserID           uint64             `json:"user_id"`
	ModificationType ModificationType   `json:"modification_type"`
	Status           ModificationStatus `json:"status"`

	// 修改前数据
	OldData string `json:"old_data"` // JSON格式
	// 修改后数据
	NewData string `json:"new_data"` // JSON格式

	Reason        string `json:"reason"` // 修改原因
	Remark        string `json:"remark"` // 备注
	RequesterID   uint64 `json:"requester_id"`
	RequesterType string `json:"requester_type"` // USER, ADMIN, SYSTEM
	ReviewerID    uint64 `json:"reviewer_id"`
	ReviewNote    string `json:"review_note"`

	// 审核相关
	ReviewRequired   bool   `json:"review_required"`
	AutoApprove      bool   `json:"auto_approve"`
	AutoApproveRules string `json:"auto_approve_rules"` // 自动批准规则 (JSON)

	// 执行相关
	ExecutedAt      *time.Time `json:"executed_at"`
	ExecutionResult string     `json:"execution_result"`
	Error           string     `json:"error"`

	// 通知相关
	Notified         bool `json:"notified"`
	NotificationSent bool `json:"notification_sent"`
}

// ModificationRule 修改规则
type ModificationRule struct {
	ID               uint64              `json:"id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	RuleName         string              `json:"rule_name"`
	ModificationType ModificationType    `json:"modification_type"`
	OrderStatus      []pb.OrderStatus    `json:"order_status"`    // 允许的订单状态
	ShippingStatus   []pb.ShippingStatus `json:"shipping_status"` // 允许的物流状态
	PaymentStatus    []pb.PaymentStatus  `json:"payment_status"`  // 允许的支付状态
	Enabled          bool                `json:"enabled"`
	AutoApprove      bool                `json:"auto_approve"`
	MaxModifyCount   int64               `json:"max_modify_count"` // 最大修改次数
	TimeLimit        time.Duration       `json:"time_limit"`       // 时间限制
	RequireReview    bool                `json:"require_review"`   // 需要审核
	ReviewerRoles    []string            `json:"reviewer_roles"`   // 审核角色
	Conditions       string              `json:"conditions"`       // 条件表达式 (JSON)
	Description      string              `json:"description"`
	CreatorID        uint64              `json:"creator_id"`
}

// ModificationHistory 修改历史
type ModificationHistory struct {
	ID               uint64           `json:"id"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	OrderID          uint64           `json:"order_id"`
	ModificationType ModificationType `json:"modification_type"`
	RequestID        uint64           `json:"request_id"`
	OldData          string           `json:"old_data"` // JSON格式
	NewData          string           `json:"new_data"` // JSON格式
	OperatorID       uint64           `json:"operator_id"`
	OperatorType     string           `json:"operator_type"`
	OperationType    string           `json:"operation_type"` // CREATE, APPROVE, REJECT, EXECUTE
	Remark           string           `json:"remark"`
	Metadata         string           `json:"metadata"` // JSON格式
}

// ModificationManager 修改管理器
type ModificationManager struct {
	orderRepo   OrderRepository
	requestRepo ModificationRepository
	ruleRepo    ModificationRuleRepository
	historyRepo ModificationHistoryRepository
}

// NewModificationManager 创建修改管理器
func NewModificationManager(orderRepo OrderRepository, requestRepo ModificationRepository,
	ruleRepo ModificationRuleRepository, historyRepo ModificationHistoryRepository) *ModificationManager {
	return &ModificationManager{
		orderRepo:   orderRepo,
		requestRepo: requestRepo,
		ruleRepo:    ruleRepo,
		historyRepo: historyRepo,
	}
}

// RequestModification 请求修改
func (m *ModificationManager) RequestModification(ctx context.Context, orderID uint64, userID uint64,
	modType ModificationType, oldData, newData interface{}, reason, requesterType string) (*ModificationRequest, error) {

	// 获取订单
	order, err := m.orderRepo.FindByID(ctx, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// 检查修改权限
	canModify, err := m.checkModificationPermission(order, userID, modType, requesterType)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}

	if !canModify {
		return nil, fmt.Errorf("modification not allowed for order in current state")
	}

	// 序列化数据
	oldDataJSON, err := json.Marshal(oldData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal old data: %w", err)
	}

	newDataJSON, err := json.Marshal(newData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new data: %w", err)
	}

	// 查找适用的规则
	rule, err := m.findApplicableRule(order, modType)
	if err != nil {
		return nil, fmt.Errorf("failed to find applicable rule: %w", err)
	}

	if rule == nil {
		return nil, fmt.Errorf("no modification rule found for type: %s", modType)
	}

	// 检查修改次数限制
	if rule.MaxModifyCount > 0 {
		modifyCount, err := m.requestRepo.CountModificationsByOrder(ctx, orderID, modType)
		if err != nil {
			return nil, fmt.Errorf("failed to count modifications: %w", err)
		}

		if int64(modifyCount) >= rule.MaxModifyCount {
			return nil, fmt.Errorf("maximum modification count reached: %d", rule.MaxModifyCount)
		}
	}

	// 检查时间限制
	if rule.TimeLimit > 0 {
		timeLimit := order.CreatedAt.Add(rule.TimeLimit)
		if time.Now().After(timeLimit) {
			return nil, fmt.Errorf("modification time limit exceeded")
		}
	}

	// 创建修改请求
	request := &ModificationRequest{
		RequestNo:        generateRequestNo(),
		OrderID:          orderID,
		OrderNo:          order.OrderNo,
		UserID:           userID,
		ModificationType: modType,
		Status:           ModificationPending,
		OldData:          string(oldDataJSON),
		NewData:          string(newDataJSON),
		Reason:           reason,
		RequesterID:      userID,
		RequesterType:    requesterType,
		ReviewRequired:   rule.RequireReview,
		AutoApprove:      rule.AutoApprove,
		AutoApproveRules: rule.Conditions,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 保存请求
	err = m.requestRepo.SaveRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to save modification request: %w", err)
	}

	// 记录修改历史
	history := &ModificationHistory{
		OrderID:          orderID,
		ModificationType: modType,
		RequestID:        request.ID,
		OldData:          string(oldDataJSON),
		NewData:          string(newDataJSON),
		OperatorID:       userID,
		OperatorType:     requesterType,
		OperationType:    "CREATE",
		Remark:           fmt.Sprintf("创建修改请求: %s", reason),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = m.historyRepo.SaveHistory(ctx, history)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to save modification history: %v\n", err)
	}

	// 检查是否自动批准
	if rule.AutoApprove {
		err = m.ApproveModification(ctx, request.ID, userID, "系统自动批准", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to auto approve modification: %w", err)
		}
	}

	return request, nil
}

// ApproveModification 批准修改
func (m *ModificationManager) ApproveModification(ctx context.Context, requestID uint64, reviewerID uint64, reviewNote string, metadata map[string]interface{}) error {
	// 获取请求
	request, err := m.requestRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get modification request: %w", err)
	}

	// 检查状态
	if request.Status != ModificationPending {
		return fmt.Errorf("request is not pending: %s", request.Status)
	}

	// 检查审核权限
	canReview, err := m.checkReviewPermission(request, reviewerID)
	if err != nil {
		return fmt.Errorf("review permission check failed: %w", err)
	}

	if !canReview {
		return fmt.Errorf("no permission to review this request")
	}

	// 更新请求状态
	request.Status = ModificationApproved
	request.ReviewerID = reviewerID
	request.ReviewNote = reviewNote
	request.UpdatedAt = time.Now()

	err = m.requestRepo.UpdateRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	// 执行修改
	result, err := m.executeModification(ctx, request)
	if err != nil {
		// 更新请求状态为失败
		request.Status = ModificationRejected
		request.Error = err.Error()
		request.UpdatedAt = time.Now()
		m.requestRepo.UpdateRequest(ctx, request)

		return fmt.Errorf("failed to execute modification: %w", err)
	}

	// 更新请求状态为已完成
	request.Status = ModificationCompleted
	now := time.Now()
	request.ExecutedAt = &now
	request.ExecutionResult = result
	request.UpdatedAt = time.Now()

	err = m.requestRepo.UpdateRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// 记录批准历史
	history := &ModificationHistory{
		OrderID:          request.OrderID,
		ModificationType: request.ModificationType,
		RequestID:        request.ID,
		OperatorID:       reviewerID,
		OperatorType:     "REVIEWER",
		OperationType:    "APPROVE",
		Remark:           fmt.Sprintf("批准修改: %s", reviewNote),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if metadata != nil {
		metadataJSON, _ := json.Marshal(metadata)
		history.Metadata = string(metadataJSON)
	}

	err = m.historyRepo.SaveHistory(ctx, history)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to save approval history: %v\n", err)
	}

	return nil
}

// RejectModification 拒绝修改
func (m *ModificationManager) RejectModification(ctx context.Context, requestID uint64, reviewerID uint64, reviewNote string) error {
	// 获取请求
	request, err := m.requestRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get modification request: %w", err)
	}

	// 检查状态
	if request.Status != ModificationPending {
		return fmt.Errorf("request is not pending: %s", request.Status)
	}

	// 检查审核权限
	canReview, err := m.checkReviewPermission(request, reviewerID)
	if err != nil {
		return fmt.Errorf("review permission check failed: %w", err)
	}

	if !canReview {
		return fmt.Errorf("no permission to review this request")
	}

	// 更新请求状态
	request.Status = ModificationRejected
	request.ReviewerID = reviewerID
	request.ReviewNote = reviewNote
	request.UpdatedAt = time.Now()

	err = m.requestRepo.UpdateRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	// 记录拒绝历史
	history := &ModificationHistory{
		OrderID:          request.OrderID,
		ModificationType: request.ModificationType,
		RequestID:        request.ID,
		OperatorID:       reviewerID,
		OperatorType:     "REVIEWER",
		OperationType:    "REJECT",
		Remark:           fmt.Sprintf("拒绝修改: %s", reviewNote),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = m.historyRepo.SaveHistory(ctx, history)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to save rejection history: %v\n", err)
	}

	return nil
}

// executeModification 执行修改
func (m *ModificationManager) executeModification(ctx context.Context, request *ModificationRequest) (string, error) {
	// 获取订单
	order, err := m.orderRepo.FindByID(ctx, request.UserID, uint64(request.OrderID))
	if err != nil {
		return "", fmt.Errorf("failed to find order: %w", err)
	}

	// 根据修改类型执行不同的操作
	switch request.ModificationType {
	case ModifyShippingAddress:
		return m.modifyShippingAddress(ctx, order, request)
	case ModifyRemark:
		return m.modifyRemark(ctx, order, request)
	case ModifyContactInfo:
		return m.modifyContactInfo(ctx, order, request)
	case ModifyProduct:
		return m.modifyProduct(ctx, order, request)
	case ModifyQuantity:
		return m.modifyQuantity(ctx, order, request)
	case ModifyPrice:
		return m.modifyPrice(ctx, order, request)
	case ModifyDiscount:
		return m.modifyDiscount(ctx, order, request)
	case ModifyShippingMethod:
		return m.modifyShippingMethod(ctx, order, request)
	default:
		return "", fmt.Errorf("unsupported modification type: %s", request.ModificationType)
	}
}

// modifyShippingAddress 修改收货地址
func (m *ModificationManager) modifyShippingAddress(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// 解析新地址数据
	var newAddr ShippingAddress
	err := json.Unmarshal([]byte(request.NewData), &newAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse new address: %w", err)
	}

	// 检查是否可以修改地址
	if order.ShippingStatus >= pb.ShippingStatus_SHIPPING_SHIPPED {
		return "", fmt.Errorf("cannot modify address after shipment")
	}

	// 执行修改
	err = order.UpdateShippingAddress(&newAddr)
	if err != nil {
		return "", fmt.Errorf("failed to update shipping address: %w", err)
	}

	// 更新订单
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	// 记录操作日志
	order.AddLog("System", "MODIFY_ADDRESS", order.Status.String(), order.Status.String(),
		fmt.Sprintf("修改收货地址: %s", request.Reason))

	result := map[string]interface{}{
		"action":    "modify_shipping_address",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"reason":    request.Reason,
		"timestamp": time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// modifyRemark 修改备注
func (m *ModificationManager) modifyRemark(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// 解析新备注数据
	var newRemark string
	err := json.Unmarshal([]byte(request.NewData), &newRemark)
	if err != nil {
		return "", fmt.Errorf("failed to parse new remark: %w", err)
	}

	// 执行修改
	order.Remark = newRemark
	order.UpdatedAt = time.Now()

	// 更新订单
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to update order: %w", err)
	}

	// 记录操作日志
	order.AddLog("System", "MODIFY_REMARK", order.Status.String(), order.Status.String(),
		fmt.Sprintf("修改订单备注: %s", request.Reason))

	result := map[string]interface{}{
		"action":    "modify_remark",
		"order_id":  order.ID,
		"order_no":  order.OrderNo,
		"reason":    request.Reason,
		"timestamp": time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// Helper methods

func (m *ModificationManager) checkModificationPermission(order *Order, userID uint64, modType ModificationType, requesterType string) (bool, error) {
	// 获取适用的规则
	rule, err := m.findApplicableRule(order, modType)
	if err != nil {
		return false, err
	}

	if rule == nil {
		return false, nil
	}

	// 检查订单状态
	if !containsStatus(order.Status, rule.OrderStatus) {
		return false, nil
	}

	// 检查物流状态
	if !containsShippingStatus(order.ShippingStatus, rule.ShippingStatus) {
		return false, nil
	}

	// 检查支付状态
	if !containsPaymentStatus(order.PaymentStatus, rule.PaymentStatus) {
		return false, nil
	}

	// 检查用户权限
	if requesterType == "USER" && userID != order.UserID {
		return false, nil
	}

	return true, nil
}

func (m *ModificationManager) checkReviewPermission(request *ModificationRequest, reviewerID uint64) (bool, error) {
	// TODO: 实现审核权限检查
	// 可以检查用户角色、部门等

	return true, nil
}

func (m *ModificationManager) findApplicableRule(order *Order, modType ModificationType) (*ModificationRule, error) {
	// TODO: 实现规则查找逻辑
	// 可以根据订单类型、状态等条件查找适用的规则

	return &ModificationRule{
		ModificationType: modType,
		Enabled:          true,
		AutoApprove:      false,
		RequireReview:    true,
	}, nil
}

func containsStatus(status pb.OrderStatus, statuses []pb.OrderStatus) bool {
	if len(statuses) == 0 {
		return true
	}

	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func containsShippingStatus(status pb.ShippingStatus, statuses []pb.ShippingStatus) bool {
	if len(statuses) == 0 {
		return true
	}

	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func containsPaymentStatus(status pb.PaymentStatus, statuses []pb.PaymentStatus) bool {
	if len(statuses) == 0 {
		return true
	}

	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func generateRequestNo() string {
	return fmt.Sprintf("MODIFY_%d", time.Now().UnixNano())
}

// 其他修改方法的实现（简化版）
func (m *ModificationManager) modifyContactInfo(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现联系方式修改
	return "", fmt.Errorf("modify contact info not implemented yet")
}

func (m *ModificationManager) modifyProduct(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现商品修改
	return "", fmt.Errorf("modify product not implemented yet")
}

func (m *ModificationManager) modifyQuantity(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现数量修改
	return "", fmt.Errorf("modify quantity not implemented yet")
}

func (m *ModificationManager) modifyPrice(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现价格修改
	return "", fmt.Errorf("modify price not implemented yet")
}

func (m *ModificationManager) modifyDiscount(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现折扣修改
	return "", fmt.Errorf("modify discount not implemented yet")
}

func (m *ModificationManager) modifyShippingMethod(ctx context.Context, order *Order, request *ModificationRequest) (string, error) {
	// TODO: 实现配送方式修改
	return "", fmt.Errorf("modify shipping method not implemented yet")
}

// ModificationRepository 修改请求仓储接口
type ModificationRepository interface {
	SaveRequest(ctx context.Context, request *ModificationRequest) error
	GetRequestByID(ctx context.Context, id uint64) (*ModificationRequest, error)
	GetRequestByNo(ctx context.Context, requestNo string) (*ModificationRequest, error)
	GetRequestsByOrderID(ctx context.Context, orderID uint64, status []ModificationStatus, page, pageSize int) ([]*ModificationRequest, int64, error)
	UpdateRequest(ctx context.Context, request *ModificationRequest) error
	DeleteRequest(ctx context.Context, id uint64) error
	CountModificationsByOrder(ctx context.Context, orderID uint64, modType ModificationType) (int64, error)
}

// ModificationRuleRepository 修改规则仓储接口
type ModificationRuleRepository interface {
	SaveRule(ctx context.Context, rule *ModificationRule) error
	GetRuleByID(ctx context.Context, id uint64) (*ModificationRule, error)
	GetRulesByType(ctx context.Context, modType ModificationType, enabledOnly bool) ([]*ModificationRule, error)
	UpdateRule(ctx context.Context, rule *ModificationRule) error
	DeleteRule(ctx context.Context, id uint64) error
}

// ModificationHistoryRepository 修改历史仓储接口
type ModificationHistoryRepository interface {
	SaveHistory(ctx context.Context, history *ModificationHistory) error
	GetHistoryByID(ctx context.Context, id uint64) (*ModificationHistory, error)
	GetHistoryByOrderID(ctx context.Context, orderID uint64, modType ModificationType, page, pageSize int) ([]*ModificationHistory, int64, error)
	GetHistoryByRequestID(ctx context.Context, requestID uint64) ([]*ModificationHistory, error)
}
