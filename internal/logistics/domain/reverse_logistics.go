// 变更说明：新增逆向物流功能，支持退货取件、换货、拒收处理、质检流程。
// 假设：退货默认7天内可申请，质检不通过自动退回买家。
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// --- 逆向物流类型 ---

// ReverseLogisticsType 逆向物流类型
type ReverseLogisticsType int

const (
	ReverseTypeReturn   ReverseLogisticsType = 1 // 退货
	ReverseTypeExchange ReverseLogisticsType = 2 // 换货
	ReverseTypeRefusal  ReverseLogisticsType = 3 // 拒收
	ReverseTypeRecall   ReverseLogisticsType = 4 // 召回
)

// --- 逆向物流状态 ---

// ReverseLogisticsStatus 逆向物流状态
type ReverseLogisticsStatus int

const (
	ReverseStatusPending    ReverseLogisticsStatus = 1  // 待处理
	ReverseStatusApproved   ReverseLogisticsStatus = 2  // 已审批
	ReverseStatusPickupWait ReverseLogisticsStatus = 3  // 待取件
	ReverseStatusPickedUp   ReverseLogisticsStatus = 4  // 已取件
	ReverseStatusInTransit  ReverseLogisticsStatus = 5  // 运输中
	ReverseStatusReceived   ReverseLogisticsStatus = 6  // 已收货
	ReverseStatusInspecting ReverseLogisticsStatus = 7  // 质检中
	ReverseStatusInspected  ReverseLogisticsStatus = 8  // 质检完成
	ReverseStatusRefunding  ReverseLogisticsStatus = 9  // 退款中
	ReverseStatusRefunded   ReverseLogisticsStatus = 10 // 已退款
	ReverseStatusResending  ReverseLogisticsStatus = 11 // 重新发货中（换货）
	ReverseStatusCompleted  ReverseLogisticsStatus = 12 // 已完成
	ReverseStatusRejected   ReverseLogisticsStatus = 13 // 已拒绝
	ReverseStatusCancelled  ReverseLogisticsStatus = 14 // 已取消
)

// --- 质检结果 ---

// InspectionResult 质检结果
type InspectionResult int

const (
	InspectionPending InspectionResult = 0 // 待质检
	InspectionPass    InspectionResult = 1 // 质检通过
	InspectionPartial InspectionResult = 2 // 部分通过
	InspectionFail    InspectionResult = 3 // 质检不通过
)

// --- 逆向物流聚合根 ---

// ReverseLogistics 逆向物流聚合根
type ReverseLogistics struct {
	ID                  uint64                 `json:"id"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	ReverseNo           string                 `json:"reverse_no"` // 逆向物流单号
	OrderID             uint64                 `json:"order_id"`
	OrderNo             string                 `json:"order_no"`
	OriginalLogisticsID uint64                 `json:"original_logistics_id"` // 原物流单ID
	Type                ReverseLogisticsType   `json:"type"`
	Status              ReverseLogisticsStatus `json:"status"`
	UserID              uint64                 `json:"user_id"`
	MerchantID          uint64                 `json:"merchant_id"`
	Reason              string                 `json:"reason"`               // 退货/换货原因
	ReasonType          string                 `json:"reason_type"`          // 原因类型
	Description         string                 `json:"description"`          // 问题描述
	Images              []string               `json:"images"`               // 图片凭证
	Items               []*ReverseItem         `json:"items"`                // 退换商品明细
	RefundAmount        int64                  `json:"refund_amount"`        // 申请退款金额
	ActualRefundAmount  int64                  `json:"actual_refund_amount"` // 实际退款金额
	PickupInfo          *PickupInfo            `json:"pickup_info"`          // 取件信息
	ReturnAddress       *Address               `json:"return_address"`       // 退货地址
	InspectionInfo      *InspectionInfo        `json:"inspection_info"`      // 质检信息
	NewLogisticsID      uint64                 `json:"new_logistics_id"`     // 换货新物流单ID
	ApprovedAt          *time.Time             `json:"approved_at"`
	ApprovedBy          string                 `json:"approved_by"`
	PickedUpAt          *time.Time             `json:"picked_up_at"`
	ReceivedAt          *time.Time             `json:"received_at"`
	InspectedAt         *time.Time             `json:"inspected_at"`
	RefundedAt          *time.Time             `json:"refunded_at"`
	CompletedAt         *time.Time             `json:"completed_at"`
	Logs                []*ReverseLogisticsLog `json:"logs"`
}

// ReverseItem 逆向物流商品项
type ReverseItem struct {
	ID               uint64           `json:"id"`
	ReverseID        uint64           `json:"reverse_id"`
	OrderItemID      uint64           `json:"order_item_id"`
	SkuID            uint64           `json:"sku_id"`
	ProductID        uint64           `json:"product_id"`
	ProductName      string           `json:"product_name"`
	Quantity         int32            `json:"quantity"`      // 退换数量
	RefundPrice      int64            `json:"refund_price"`  // 退款单价
	RefundAmount     int64            `json:"refund_amount"` // 退款金额
	InspectionResult InspectionResult `json:"inspection_result"`
}

// PickupInfo 取件信息
type PickupInfo struct {
	CourierCompany   string     `json:"courier_company"`    // 快递公司
	CourierName      string     `json:"courier_name"`       // 快递员姓名
	CourierPhone     string     `json:"courier_phone"`      // 快递员电话
	PickupAddress    string     `json:"pickup_address"`     // 取件地址
	PickupTime       *time.Time `json:"pickup_time"`        // 预约取件时间
	ActualPickupTime *time.Time `json:"actual_pickup_time"` // 实际取件时间
	TrackingNo       string     `json:"tracking_no"`        // 快递单号
}

// Address 地址
type Address struct {
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	ContactName   string `json:"contact_name"`
	ContactPhone  string `json:"contact_phone"`
}

// InspectionInfo 质检信息
type InspectionInfo struct {
	InspectorID      string           `json:"inspector_id"`
	InspectorName    string           `json:"inspector_name"`
	InspectionTime   *time.Time       `json:"inspection_time"`
	OverallResult    InspectionResult `json:"overall_result"`
	IssuesFound      []string         `json:"issues_found"`      // 发现的问题
	AcceptedQuantity int32            `json:"accepted_quantity"` // 接受数量
	RejectedQuantity int32            `json:"rejected_quantity"` // 拒绝数量
	Remark           string           `json:"remark"`
}

// ReverseLogisticsLog 逆向物流操作日志
type ReverseLogisticsLog struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ReverseID uint64    `json:"reverse_id"`
	Action    string    `json:"action"`
	Operator  string    `json:"operator"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Remark    string    `json:"remark"`
}

// NewReverseLogistics 创建逆向物流
func NewReverseLogistics(reverseNo, orderNo string, orderID, userID, merchantID uint64, logisticsType ReverseLogisticsType, reason, reasonType, description string, images []string) *ReverseLogistics {
	return &ReverseLogistics{
		ReverseNo:   reverseNo,
		OrderID:     orderID,
		OrderNo:     orderNo,
		Type:        logisticsType,
		Status:      ReverseStatusPending,
		UserID:      userID,
		MerchantID:  merchantID,
		Reason:      reason,
		ReasonType:  reasonType,
		Description: description,
		Images:      images,
		Items:       make([]*ReverseItem, 0),
		Logs:        make([]*ReverseLogisticsLog, 0),
	}
}

// AddItem 添加退换商品
func (r *ReverseLogistics) AddItem(orderItemID, skuID, productID uint64, productName string, quantity int32, refundPrice int64) {
	item := &ReverseItem{
		OrderItemID:      orderItemID,
		SkuID:            skuID,
		ProductID:        productID,
		ProductName:      productName,
		Quantity:         quantity,
		RefundPrice:      refundPrice,
		RefundAmount:     refundPrice * int64(quantity),
		InspectionResult: InspectionPending,
	}
	r.Items = append(r.Items, item)
	r.RefundAmount += item.RefundAmount
}

// Approve 审批通过
func (r *ReverseLogistics) Approve(ctx context.Context, approver string, returnAddr *Address) error {
	if r.Status != ReverseStatusPending {
		return errors.New("can only approve pending reverse logistics")
	}

	now := time.Now()
	r.Status = ReverseStatusApproved
	r.ApprovedAt = &now
	r.ApprovedBy = approver
	r.ReturnAddress = returnAddr

	r.addLog("APPROVE", approver, ReverseStatusPending.String(), ReverseStatusApproved.String(), "Approved by merchant")
	return nil
}

// Reject 审批拒绝
func (r *ReverseLogistics) Reject(ctx context.Context, rejecter, reason string) error {
	if r.Status != ReverseStatusPending {
		return errors.New("can only reject pending reverse logistics")
	}

	r.Status = ReverseStatusRejected
	r.addLog("REJECT", rejecter, ReverseStatusPending.String(), ReverseStatusRejected.String(), reason)
	return nil
}

// SchedulePickup 预约取件
func (r *ReverseLogistics) SchedulePickup(ctx context.Context, pickupInfo *PickupInfo) error {
	if r.Status != ReverseStatusApproved {
		return errors.New("can only schedule pickup for approved reverse logistics")
	}

	r.PickupInfo = pickupInfo
	r.Status = ReverseStatusPickupWait
	r.addLog("SCHEDULE_PICKUP", "System", ReverseStatusApproved.String(), ReverseStatusPickupWait.String(), fmt.Sprintf("Pickup scheduled: %s", pickupInfo.PickupTime))
	return nil
}

// ConfirmPickup 确认取件
func (r *ReverseLogistics) ConfirmPickup(ctx context.Context, trackingNo string, operator string) error {
	if r.Status != ReverseStatusPickupWait {
		return errors.New("can only confirm pickup for waiting status")
	}

	now := time.Now()
	r.Status = ReverseStatusPickedUp
	r.PickedUpAt = &now
	r.PickupInfo.TrackingNo = trackingNo
	r.PickupInfo.ActualPickupTime = &now

	r.addLog("CONFIRM_PICKUP", operator, ReverseStatusPickupWait.String(), ReverseStatusPickedUp.String(), fmt.Sprintf("Tracking: %s", trackingNo))
	return nil
}

// MarkInTransit 标记运输中
func (r *ReverseLogistics) MarkInTransit(ctx context.Context) error {
	if r.Status != ReverseStatusPickedUp {
		return errors.New("can only mark in transit for picked up status")
	}

	r.Status = ReverseStatusInTransit
	r.addLog("IN_TRANSIT", "System", ReverseStatusPickedUp.String(), ReverseStatusInTransit.String(), "Package in transit")
	return nil
}

// ReceivePackage 收到退货包裹
func (r *ReverseLogistics) ReceivePackage(ctx context.Context, operator string) error {
	if r.Status != ReverseStatusInTransit {
		return errors.New("can only receive for in transit status")
	}

	now := time.Now()
	r.Status = ReverseStatusReceived
	r.ReceivedAt = &now

	r.addLog("RECEIVE", operator, ReverseStatusInTransit.String(), ReverseStatusReceived.String(), "Package received")
	return nil
}

// StartInspection 开始质检
func (r *ReverseLogistics) StartInspection(ctx context.Context, inspectorID, inspectorName string) error {
	if r.Status != ReverseStatusReceived {
		return errors.New("can only start inspection for received status")
	}

	r.Status = ReverseStatusInspecting
	r.InspectionInfo = &InspectionInfo{
		InspectorID:   inspectorID,
		InspectorName: inspectorName,
	}

	r.addLog("START_INSPECTION", inspectorName, ReverseStatusReceived.String(), ReverseStatusInspecting.String(), "Inspection started")
	return nil
}

// CompleteInspection 完成质检
func (r *ReverseLogistics) CompleteInspection(ctx context.Context, result InspectionResult, acceptedQty, rejectedQty int32, issues []string, remark string) error {
	if r.Status != ReverseStatusInspecting {
		return errors.New("can only complete inspection for inspecting status")
	}

	now := time.Now()
	r.Status = ReverseStatusInspected
	r.InspectedAt = &now
	r.InspectionInfo.InspectionTime = &now
	r.InspectionInfo.OverallResult = result
	r.InspectionInfo.AcceptedQuantity = acceptedQty
	r.InspectionInfo.RejectedQuantity = rejectedQty
	r.InspectionInfo.IssuesFound = issues
	r.InspectionInfo.Remark = remark

	// 更新商品质检结果
	for _, item := range r.Items {
		if acceptedQty > 0 {
			item.InspectionResult = InspectionPass
		} else {
			item.InspectionResult = InspectionFail
		}
	}

	// 计算实际退款金额（基于质检通过数量）
	if result == InspectionPass {
		r.ActualRefundAmount = r.RefundAmount
	} else if result == InspectionPartial {
		// 按比例计算
		totalQty := acceptedQty + rejectedQty
		if totalQty > 0 {
			r.ActualRefundAmount = r.RefundAmount * int64(acceptedQty) / int64(totalQty)
		}
	}
	// InspectionFail 不退款

	r.addLog("COMPLETE_INSPECTION", r.InspectionInfo.InspectorName, ReverseStatusInspecting.String(), ReverseStatusInspected.String(), fmt.Sprintf("Result: %v, Accepted: %d, Rejected: %d", result, acceptedQty, rejectedQty))
	return nil
}

// ProcessRefund 处理退款
func (r *ReverseLogistics) ProcessRefund(ctx context.Context) error {
	if r.Status != ReverseStatusInspected {
		return errors.New("can only process refund for inspected status")
	}
	if r.InspectionInfo.OverallResult == InspectionFail {
		return errors.New("cannot refund: inspection failed")
	}

	r.Status = ReverseStatusRefunding
	r.addLog("PROCESS_REFUND", "System", ReverseStatusInspected.String(), ReverseStatusRefunding.String(), fmt.Sprintf("Refund amount: %d", r.ActualRefundAmount))
	return nil
}

// ConfirmRefund 确认退款完成
func (r *ReverseLogistics) ConfirmRefund(ctx context.Context, transactionID string) error {
	if r.Status != ReverseStatusRefunding {
		return errors.New("can only confirm refund for refunding status")
	}

	now := time.Now()
	switch r.Type {
	case ReverseTypeReturn:
		r.Status = ReverseStatusRefunded
		r.RefundedAt = &now
		r.CompletedAt = &now
		r.addLog("REFUND_COMPLETED", "System", ReverseStatusRefunding.String(), ReverseStatusRefunded.String(), fmt.Sprintf("Transaction: %s", transactionID))
	case ReverseTypeExchange:
		r.Status = ReverseStatusResending
		r.RefundedAt = &now
		r.addLog("REFUND_COMPLETED_EXCHANGE", "System", ReverseStatusRefunding.String(), ReverseStatusResending.String(), "Preparing to resend")
	}
	return nil
}

// ConfirmResend 确认换货发出
func (r *ReverseLogistics) ConfirmResend(ctx context.Context, newLogisticsID uint64) error {
	if r.Status != ReverseStatusResending {
		return errors.New("can only confirm resend for resending status")
	}

	now := time.Now()
	r.Status = ReverseStatusCompleted
	r.NewLogisticsID = newLogisticsID
	r.CompletedAt = &now

	r.addLog("RESEND_COMPLETED", "System", ReverseStatusResending.String(), ReverseStatusCompleted.String(), fmt.Sprintf("New logistics: %d", newLogisticsID))
	return nil
}

// Cancel 取消逆向物流
func (r *ReverseLogistics) Cancel(ctx context.Context, reason string, operator string) error {
	if r.Status >= ReverseStatusReceived {
		return errors.New("cannot cancel after package received")
	}

	oldStatus := r.Status
	r.Status = ReverseStatusCancelled
	r.addLog("CANCEL", operator, oldStatus.String(), ReverseStatusCancelled.String(), reason)
	return nil
}

// addLog 添加操作日志
func (r *ReverseLogistics) addLog(action, operator, oldStatus, newStatus, remark string) {
	r.Logs = append(r.Logs, &ReverseLogisticsLog{
		Action:    action,
		Operator:  operator,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
	})
}

// String 返回状态字符串
func (s ReverseLogisticsStatus) String() string {
	statusNames := []string{"", "PENDING", "APPROVED", "PICKUP_WAIT", "PICKED_UP", "IN_TRANSIT", "RECEIVED", "INSPECTING", "INSPECTED", "REFUNDING", "REFUNDED", "RESENDING", "COMPLETED", "REJECTED", "CANCELLED"}
	if int(s) < len(statusNames) {
		return statusNames[s]
	}
	return "UNKNOWN"
}

// --- 逆向物流仓储接口 ---

// ReverseLogisticsRepository 逆向物流仓储接口
type ReverseLogisticsRepository interface {
	Save(ctx context.Context, reverse *ReverseLogistics) error
	Update(ctx context.Context, reverse *ReverseLogistics) error
	FindByID(ctx context.Context, id uint64) (*ReverseLogistics, error)
	FindByReverseNo(ctx context.Context, reverseNo string) (*ReverseLogistics, error)
	FindByOrderID(ctx context.Context, orderID uint64) ([]*ReverseLogistics, error)
	FindByUserID(ctx context.Context, userID uint64, status []ReverseLogisticsStatus) ([]*ReverseLogistics, error)
	FindPendingPickup(ctx context.Context) ([]*ReverseLogistics, error)
	FindPendingInspection(ctx context.Context) ([]*ReverseLogistics, error)
}

// --- 拒收处理 ---

// RefusalHandler 拒收处理器
type RefusalHandler struct{}

// HandleRefusal 处理拒收
func (h *RefusalHandler) HandleRefusal(ctx context.Context, orderID uint64, orderNo, reason string, userID, merchantID uint64, items []*ReverseItem, reverseNoGenerator func() string) (*ReverseLogistics, error) {
	reverse := NewReverseLogistics(
		reverseNoGenerator(),
		orderNo,
		orderID,
		userID,
		merchantID,
		ReverseTypeRefusal,
		reason,
		"REFUSAL",
		"Customer refused delivery",
		nil,
	)

	for _, item := range items {
		reverse.Items = append(reverse.Items, item)
		reverse.RefundAmount += item.RefundAmount
	}

	// 拒收直接进入取件状态（由快递员带回）
	reverse.Status = ReverseStatusPickedUp
	reverse.addLog("REFUSAL_INIT", "System", "", ReverseStatusPickedUp.String(), reason)

	return reverse, nil
}

// --- 召回处理 ---

// RecallHandler 召回处理器
type RecallHandler struct{}

// CreateRecall 创建召回单（批量）
func (h *RecallHandler) CreateRecall(ctx context.Context, merchantID uint64, reason string, affectedOrders []uint64, reverseNoGenerator func() string) ([]*ReverseLogistics, error) {
	var recalls []*ReverseLogistics
	for _, orderID := range affectedOrders {
		reverse := NewReverseLogistics(
			reverseNoGenerator(),
			"",
			orderID,
			0, // 召回不指定用户
			merchantID,
			ReverseTypeRecall,
			reason,
			"RECALL",
			"Product recall initiated by manufacturer",
			nil,
		)
		reverse.Status = ReverseStatusApproved // 召回自动审批
		recalls = append(recalls, reverse)
	}
	return recalls, nil
}
