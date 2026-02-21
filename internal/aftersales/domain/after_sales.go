// 变更说明：
// 1. 【合并】将 returns 和 refund 服务的核心逻辑合并到 aftersales 统一售后服务
// 2. 【增强】新增退款路由（原路返回/退至钱包/线下打款/退为优惠券）
// 3. 【增强】新增退货物流追踪（寄回/签收/质检）完整流程
// 4. 【增强】所有状态变更增加状态机校验，防止非法状态流转
// 5. 【增强】新增操作日志自动记录
// 6. 【增强】新增部分退款类型
package domain

import (
	"errors"
	"time"
)

// 定义AfterSales模块的业务错误。
var (
	ErrAfterSalesNotFound    = errors.New("after-sales record not found")    // 售后记录未找到。
	ErrInvalidStatus         = errors.New("invalid status for operation")    // 操作状态无效。
	ErrRefundAmountExceed    = errors.New("refund amount exceeds order total") // 退款金额超出订单总额。
	ErrRefundChannelFailed   = errors.New("refund channel execution failed")  // 退款渠道执行失败。
	ErrReturnShipmentExpired = errors.New("return shipment deadline expired")  // 退货寄回超时。
	ErrDuplicateAfterSales   = errors.New("duplicate after-sales request")    // 重复售后申请。
)

// AfterSalesType 定义了售后请求的类型。
type AfterSalesType int8

const (
	AfterSalesTypeReturnGoods   AfterSalesType = 1 // 退货：客户将商品退回商家。
	AfterSalesTypeExchange      AfterSalesType = 2 // 换货：客户要求更换商品。
	AfterSalesTypeRefund        AfterSalesType = 3 // 退款：客户要求退回款项。
	AfterSalesTypeRepair        AfterSalesType = 4 // 维修：商品需要维修。
	AfterSalesTypeComplaint     AfterSalesType = 5 // 投诉：客户对商品或服务不满。
	AfterSalesTypePartialRefund AfterSalesType = 6 // 部分退款：仅退款不退货（如商品瑕疵补偿）。
)

// AfterSalesStatus 定义了售后请求的生命周期状态。
type AfterSalesStatus int8

const (
	AfterSalesStatusPending        AfterSalesStatus = 1  // 待处理：申请已提交，等待商家审核。
	AfterSalesStatusApproved       AfterSalesStatus = 2  // 已批准：商家已同意售后申请。
	AfterSalesStatusRejected       AfterSalesStatus = 3  // 已拒绝：商家已拒绝售后申请。
	AfterSalesStatusInProgress     AfterSalesStatus = 4  // 处理中：售后流程正在进行。
	AfterSalesStatusQCPassed       AfterSalesStatus = 5  // 质检通过。
	AfterSalesStatusQCFailed       AfterSalesStatus = 6  // 质检失败。
	AfterSalesStatusCompleted      AfterSalesStatus = 7  // 已完成：售后流程已结束。
	AfterSalesStatusCancelled      AfterSalesStatus = 8  // 已取消：售后申请被取消。
	AfterSalesStatusRefunding      AfterSalesStatus = 9  // 退款中：退款请求已提交至支付渠道。
	AfterSalesStatusRefundSuccess  AfterSalesStatus = 10 // 退款成功：资金已退回用户。
	AfterSalesStatusRefundFailed   AfterSalesStatus = 11 // 退款失败：渠道退款异常，需人工介入。
	AfterSalesStatusReturnShipping AfterSalesStatus = 12 // 退货寄回中：用户已寄出退货包裹。
	AfterSalesStatusReturnReceived AfterSalesStatus = 13 // 退货已签收：仓库已签收退货包裹。
)

// String 方法返回 AfterSalesStatus 的字符串表示。
func (s AfterSalesStatus) String() string {
	switch s {
	case AfterSalesStatusPending:
		return "Pending"
	case AfterSalesStatusApproved:
		return "Approved"
	case AfterSalesStatusRejected:
		return "Rejected"
	case AfterSalesStatusInProgress:
		return "InProgress"
	case AfterSalesStatusQCPassed:
		return "QCPassed"
	case AfterSalesStatusQCFailed:
		return "QCFailed"
	case AfterSalesStatusCompleted:
		return "Completed"
	case AfterSalesStatusCancelled:
		return "Cancelled"
	case AfterSalesStatusRefunding:
		return "Refunding"
	case AfterSalesStatusRefundSuccess:
		return "RefundSuccess"
	case AfterSalesStatusRefundFailed:
		return "RefundFailed"
	case AfterSalesStatusReturnShipping:
		return "ReturnShipping"
	case AfterSalesStatusReturnReceived:
		return "ReturnReceived"
	default:
		return "Unknown"
	}
}

// RefundChannel 退款渠道类型。
type RefundChannel string

const (
	// RefundChannelOriginal 原路返回：退款至原支付渠道。
	RefundChannelOriginal RefundChannel = "ORIGINAL"
	// RefundChannelWallet 退至钱包：退款至用户平台钱包余额。
	RefundChannelWallet RefundChannel = "WALLET"
	// RefundChannelOffline 线下打款：通过银行转账等线下方式退款。
	RefundChannelOffline RefundChannel = "OFFLINE"
	// RefundChannelCoupon 退为优惠券：将退款金额转为平台优惠券。
	RefundChannelCoupon RefundChannel = "COUPON"
)

// RefundStatus 退款执行状态。
type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "PENDING"    // 待执行。
	RefundStatusProcessing RefundStatus = "PROCESSING" // 处理中。
	RefundStatusSuccess    RefundStatus = "SUCCESS"     // 成功。
	RefundStatusFailed     RefundStatus = "FAILED"      // 失败。
	RefundStatusSuspended  RefundStatus = "SUSPENDED"   // 挂起，需人工介入。
)

// AfterSales 实体是售后模块的聚合根。
// 它代表一个完整的售后申请，包含售后单号、订单信息、用户、类型、状态、原因、商品列表和操作日志等。
type AfterSales struct {
	ID              uint64            `json:"id"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	AfterSalesNo    string            `json:"after_sales_no"`   // 售后单的唯一编号，唯一索引。
	OrderID         uint64            `json:"order_id"`         // 关联的订单ID，索引字段。
	OrderNo         string            `json:"order_no"`         // 关联的订单编号。
	UserID          uint64            `json:"user_id"`          // 发起售后申请的用户ID，索引字段。
	Type            AfterSalesType    `json:"type"`             // 售后类型。
	Status          AfterSalesStatus  `json:"status"`           // 售后单状态，默认为待处理。
	Reason          string            `json:"reason"`           // 客户提交的申请原因。
	Description     string            `json:"description"`      // 详细的售后描述。
	Images          []string          `json:"images"`           // 客户提供的凭证图片URL列表。
	RefundAmount    int64             `json:"refund_amount"`    // 订单中实际产生的退款金额（总额）。
	ApprovalAmount  int64             `json:"approval_amount"`  // 实际批准的退款金额或补偿金额。
	ApprovedBy      string            `json:"approved_by"`      // 批准售后申请的操作人员。
	RejectionReason string            `json:"rejection_reason"` // 拒绝售后申请的原因。
	ApprovedAt      *time.Time        `json:"approved_at"`      // 售后申请被批准的时间。
	RejectedAt      *time.Time        `json:"rejected_at"`      // 售后申请被拒绝的时间。
	CompletedAt     *time.Time        `json:"completed_at"`     // 售后流程完成的时间。
	CancelledAt     *time.Time        `json:"cancelled_at"`     // 售后申请被取消的时间。
	RMANumber       string            `json:"rma_number"`       // 退货授权号。
	TrackingNumber  string            `json:"tracking_number"`  // 退货物流单号。
	WarehouseNotes  string            `json:"warehouse_notes"`  // 仓库备注。
	Items           []*AfterSalesItem `json:"items"`            // 售后申请包含的商品列表，一对多关系。
	Logs            []*AfterSalesLog  `json:"logs"`             // 售后申请的操作日志列表，一对多关系。
	RefundInfo      *RefundInfo       `json:"refund_info"`      // 退款详情（退款类售后单必填）。
	ReturnInfo      *ReturnInfo       `json:"return_info"`      // 退货物流详情（退货类售后单必填）。
}

// RefundInfo 退款详情值对象。
// 记录退款渠道、支付流水、退款执行状态等信息。
type RefundInfo struct {
	RefundNo     string        `json:"refund_no"`      // 退款单号。
	PaymentID    string        `json:"payment_id"`     // 原支付单号。
	Channel      RefundChannel `json:"channel"`        // 退款渠道。
	RefundStatus RefundStatus  `json:"refund_status"`  // 退款执行状态。
	ChannelTxID  string        `json:"channel_tx_id"`  // 支付渠道退款流水号。
	RefundAmount int64         `json:"refund_amount"`  // 退款金额（分）。
	Currency     string        `json:"currency"`       // 币种。
	RetryCount   int           `json:"retry_count"`    // 重试次数。
	MaxRetry     int           `json:"max_retry"`      // 最大重试次数。
	ErrorMessage string        `json:"error_message"`  // 失败原因。
	SubmittedAt  *time.Time    `json:"submitted_at"`   // 提交时间。
	CompletedAt  *time.Time    `json:"completed_at"`   // 完成时间。
}

// ReturnInfo 退货物流详情值对象。
// 记录退货寄回的物流信息、签收状态、质检结果等。
type ReturnInfo struct {
	LogisticsCompany string     `json:"logistics_company"` // 物流公司名称。
	TrackingNumber   string     `json:"tracking_number"`   // 物流单号。
	ShippedAt        *time.Time `json:"shipped_at"`        // 用户寄出时间。
	ReceivedAt       *time.Time `json:"received_at"`       // 仓库签收时间。
	Deadline         *time.Time `json:"deadline"`          // 寄回截止时间。
	ReturnAddress    string     `json:"return_address"`    // 退货地址。
	QCResult         string     `json:"qc_result"`         // 质检结果（PASS/FAIL/PARTIAL）。
	QCNotes          string     `json:"qc_notes"`          // 质检备注。
	QCImages         []string   `json:"qc_images"`         // 质检图片。
}

// AfterSalesItem 实体代表售后申请中的一个商品项。
type AfterSalesItem struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	AfterSalesID uint64    `json:"after_sales_id"` // 关联的售后单ID，索引字段。
	ProductID    uint64    `json:"product_id"`     // 商品ID。
	SkuID        uint64    `json:"sku_id"`         // SKU ID。
	ProductName  string    `json:"product_name"`   // 商品名称。
	SkuName      string    `json:"sku_name"`       // SKU名称（例如，颜色、尺码）。
	Quantity     int32     `json:"quantity"`        // 申请售后的商品数量。
	Price        int64     `json:"price"`           // 商品的单价（单位：分）。
	TotalPrice   int64     `json:"total_price"`     // 商品项的总价（单价 * 数量）。
	Reason       string    `json:"reason"`          // 售后原因。
	Images       []string  `json:"images"`          // 商品图片。
}

// AfterSalesLog 实体代表售后单的某次操作日志。
type AfterSalesLog struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	AfterSalesID uint64    `json:"after_sales_id"` // 关联的售后单ID，索引字段。
	Operator     string    `json:"operator"`       // 执行操作的人员。
	Action       string    `json:"action"`         // 执行的操作类型。
	OldStatus    string    `json:"old_status"`     // 操作前的售后单状态。
	NewStatus    string    `json:"new_status"`     // 操作后的售后单状态。
	Remark       string    `json:"remark"`         // 操作的备注信息。
}

// NewAfterSales 创建并返回一个新的 AfterSales 实体实例。
func NewAfterSales(afterSalesNo string, orderID uint64, orderNo string, userID uint64, afterSalesType AfterSalesType, reason, description string, images []string) *AfterSales {
	return &AfterSales{
		AfterSalesNo: afterSalesNo,
		OrderID:      orderID,
		OrderNo:      orderNo,
		UserID:       userID,
		Type:         afterSalesType,
		Status:       AfterSalesStatusPending,
		Reason:       reason,
		Description:  description,
		Images:       images,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Items:        []*AfterSalesItem{},
		Logs:         []*AfterSalesLog{},
	}
}

// Approve 批准售后申请。
// 状态机校验：仅 Pending 状态可批准。
func (a *AfterSales) Approve(operator string, amount int64) error {
	if a.Status != AfterSalesStatusPending {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusApproved
	a.ApprovedBy = operator
	a.ApprovalAmount = amount
	now := time.Now()
	a.ApprovedAt = &now
	a.addLog(operator, "Approve", oldStatus, AfterSalesStatusApproved, "")
	return nil
}

// Reject 拒绝售后申请。
// 状态机校验：仅 Pending 状态可拒绝。
func (a *AfterSales) Reject(operator, reason string) error {
	if a.Status != AfterSalesStatusPending {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusRejected
	a.RejectionReason = reason
	now := time.Now()
	a.RejectedAt = &now
	a.addLog(operator, "Reject", oldStatus, AfterSalesStatusRejected, reason)
	return nil
}

// Process 开始处理售后申请。
// 状态机校验：仅 Approved 状态可开始处理。
func (a *AfterSales) Process(operator string) error {
	if a.Status != AfterSalesStatusApproved {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusInProgress
	a.addLog(operator, "Process", oldStatus, AfterSalesStatusInProgress, "")
	return nil
}

// SubmitReturnShipment 用户提交退货物流信息。
// 状态机校验：仅 Approved / InProgress 状态可提交退货物流。
func (a *AfterSales) SubmitReturnShipment(operator, company, trackingNo, returnAddr string, deadline time.Time) error {
	allowed := map[AfterSalesStatus]bool{
		AfterSalesStatusApproved:   true,
		AfterSalesStatusInProgress: true,
	}
	if !allowed[a.Status] {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusReturnShipping
	now := time.Now()
	a.ReturnInfo = &ReturnInfo{
		LogisticsCompany: company,
		TrackingNumber:   trackingNo,
		ShippedAt:        &now,
		Deadline:         &deadline,
		ReturnAddress:    returnAddr,
	}
	a.TrackingNumber = trackingNo
	a.addLog(operator, "SubmitReturnShipment", oldStatus, AfterSalesStatusReturnShipping, trackingNo)
	return nil
}

// ConfirmReturnReceived 仓库确认收到退货包裹。
// 状态机校验：仅 ReturnShipping 状态可确认签收。
func (a *AfterSales) ConfirmReturnReceived(operator string) error {
	if a.Status != AfterSalesStatusReturnShipping {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusReturnReceived
	now := time.Now()
	if a.ReturnInfo != nil {
		a.ReturnInfo.ReceivedAt = &now
	}
	a.addLog(operator, "ConfirmReturnReceived", oldStatus, AfterSalesStatusReturnReceived, "")
	return nil
}

// SetQCResult 设置质检结果。
// 状态机校验：仅 ReturnReceived 状态可设置质检结果。
func (a *AfterSales) SetQCResult(operator string, passed bool, notes string, images []string) error {
	if a.Status != AfterSalesStatusReturnReceived {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	if passed {
		a.Status = AfterSalesStatusQCPassed
	} else {
		a.Status = AfterSalesStatusQCFailed
	}
	a.WarehouseNotes = notes
	a.UpdatedAt = time.Now()
	if a.ReturnInfo != nil {
		result := "PASS"
		if !passed {
			result = "FAIL"
		}
		a.ReturnInfo.QCResult = result
		a.ReturnInfo.QCNotes = notes
		a.ReturnInfo.QCImages = images
	}
	a.addLog(operator, "QCResult", oldStatus, a.Status, notes)
	return nil
}

// InitiateRefund 发起退款。
// 状态机校验：QCPassed / Approved（仅退款不退货场景）可发起退款。
func (a *AfterSales) InitiateRefund(operator, refundNo, paymentID string, amount int64, currency string, channel RefundChannel) error {
	allowed := map[AfterSalesStatus]bool{
		AfterSalesStatusQCPassed: true,
		AfterSalesStatusApproved: true,
	}
	if !allowed[a.Status] {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusRefunding
	now := time.Now()
	a.RefundInfo = &RefundInfo{
		RefundNo:     refundNo,
		PaymentID:    paymentID,
		Channel:      channel,
		RefundStatus: RefundStatusProcessing,
		RefundAmount: amount,
		Currency:     currency,
		MaxRetry:     3,
		SubmittedAt:  &now,
	}
	a.addLog(operator, "InitiateRefund", oldStatus, AfterSalesStatusRefunding, refundNo)
	return nil
}

// ConfirmRefundSuccess 确认退款成功（支付渠道回调）。
func (a *AfterSales) ConfirmRefundSuccess(channelTxID string) error {
	if a.Status != AfterSalesStatusRefunding {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusRefundSuccess
	now := time.Now()
	if a.RefundInfo != nil {
		a.RefundInfo.RefundStatus = RefundStatusSuccess
		a.RefundInfo.ChannelTxID = channelTxID
		a.RefundInfo.CompletedAt = &now
	}
	a.addLog("SYSTEM", "RefundSuccess", oldStatus, AfterSalesStatusRefundSuccess, channelTxID)
	return nil
}

// MarkRefundFailed 标记退款失败。
// 未达最大重试次数时保持 Refunding 状态等待重试，达到后标记为 RefundFailed。
func (a *AfterSales) MarkRefundFailed(errorMsg string) error {
	if a.Status != AfterSalesStatusRefunding {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	if a.RefundInfo != nil {
		a.RefundInfo.RetryCount++
		if a.RefundInfo.RetryCount >= a.RefundInfo.MaxRetry {
			a.Status = AfterSalesStatusRefundFailed
			a.RefundInfo.RefundStatus = RefundStatusFailed
			a.RefundInfo.ErrorMessage = errorMsg
		} else {
			a.RefundInfo.RefundStatus = RefundStatusPending
			a.RefundInfo.ErrorMessage = errorMsg
			return nil
		}
	} else {
		a.Status = AfterSalesStatusRefundFailed
	}
	a.addLog("SYSTEM", "RefundFailed", oldStatus, a.Status, errorMsg)
	return nil
}

// Complete 完成售后申请。
// 状态机校验：RefundSuccess / QCPassed / InProgress 可标记完成。
func (a *AfterSales) Complete(operator string) error {
	allowed := map[AfterSalesStatus]bool{
		AfterSalesStatusRefundSuccess: true,
		AfterSalesStatusQCPassed:      true,
		AfterSalesStatusInProgress:    true,
	}
	if !allowed[a.Status] {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusCompleted
	now := time.Now()
	a.CompletedAt = &now
	a.addLog(operator, "Complete", oldStatus, AfterSalesStatusCompleted, "")
	return nil
}

// Cancel 取消售后申请。
// 状态机校验：仅 Pending / Approved 状态可取消。
func (a *AfterSales) Cancel(operator string) error {
	allowed := map[AfterSalesStatus]bool{
		AfterSalesStatusPending:  true,
		AfterSalesStatusApproved: true,
	}
	if !allowed[a.Status] {
		return ErrInvalidStatus
	}
	oldStatus := a.Status
	a.Status = AfterSalesStatusCancelled
	now := time.Now()
	a.CancelledAt = &now
	a.addLog(operator, "Cancel", oldStatus, AfterSalesStatusCancelled, "")
	return nil
}

// IsReturnRequired 判断该售后类型是否需要退货。
func (a *AfterSales) IsReturnRequired() bool {
	return a.Type == AfterSalesTypeReturnGoods || a.Type == AfterSalesTypeExchange
}

// IsRefundRequired 判断该售后类型是否需要退款。
func (a *AfterSales) IsRefundRequired() bool {
	return a.Type == AfterSalesTypeReturnGoods ||
		a.Type == AfterSalesTypeRefund ||
		a.Type == AfterSalesTypePartialRefund
}

// CanCancel 判断当前状态是否可取消。
func (a *AfterSales) CanCancel() bool {
	return a.Status == AfterSalesStatusPending || a.Status == AfterSalesStatusApproved
}

// IsTerminal 判断是否为终态。
func (a *AfterSales) IsTerminal() bool {
	return a.Status == AfterSalesStatusCompleted ||
		a.Status == AfterSalesStatusCancelled ||
		a.Status == AfterSalesStatusRejected
}

// addLog 添加操作日志。
func (a *AfterSales) addLog(operator, action string, oldStatus, newStatus AfterSalesStatus, remark string) {
	a.Logs = append(a.Logs, &AfterSalesLog{
		AfterSalesID: a.ID,
		Operator:     operator,
		Action:       action,
		OldStatus:    oldStatus.String(),
		NewStatus:    newStatus.String(),
		Remark:       remark,
		CreatedAt:    time.Now(),
	})
	a.UpdatedAt = time.Now()
}

// SupportTicketStatus 定义工单状态。
type SupportTicketStatus int8

const (
	SupportTicketStatusOpen     SupportTicketStatus = 1 // 开启。
	SupportTicketStatusPending  SupportTicketStatus = 2 // 待处理。
	SupportTicketStatusResolved SupportTicketStatus = 3 // 已解决。
	SupportTicketStatusClosed   SupportTicketStatus = 4 // 已关闭。
)

// String 返回工单状态的字符串表示。
func (s SupportTicketStatus) String() string {
	switch s {
	case SupportTicketStatusOpen:
		return "Open"
	case SupportTicketStatusPending:
		return "Pending"
	case SupportTicketStatusResolved:
		return "Resolved"
	case SupportTicketStatusClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// SupportTicket 客服工单实体。
type SupportTicket struct {
	ID          uint64                  `json:"id"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	TicketNo    string                  `json:"ticket_no"`
	UserID      uint64                  `json:"user_id"`
	OrderID     uint64                  `json:"order_id"`
	Subject     string                  `json:"subject"`
	Description string                  `json:"description"`
	Status      SupportTicketStatus     `json:"status"`
	Priority    int8                    `json:"priority"` // 1: Low, 2: Medium, 3: High
	Category    string                  `json:"category"`
	Messages    []*SupportTicketMessage `json:"messages"`
}

// SupportTicketMessage 工单消息实体。
type SupportTicketMessage struct {
	ID         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	TicketID   uint64    `json:"ticket_id"`
	SenderID   uint64    `json:"sender_id"`   // 0 表示系统/客服，>0 表示用户。
	SenderType string    `json:"sender_type"` // User, Agent, System。
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
}

// AfterSalesConfig 售后配置实体。
type AfterSalesConfig struct {
	ID          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
}
