package domain

import (
	"errors"
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// 定义AfterSales模块的业务错误。
var (
	ErrAfterSalesNotFound = errors.New("after-sales record not found") // 售后记录未找到。
	ErrInvalidStatus      = errors.New("invalid status for operation") // 操作状态无效。
)

// AfterSalesType 定义了售后请求的类型。
type AfterSalesType int8

const (
	AfterSalesTypeReturnGoods AfterSalesType = 1 // 退货：客户将商品退回商家。
	AfterSalesTypeExchange    AfterSalesType = 2 // 换货：客户要求更换商品。
	AfterSalesTypeRefund      AfterSalesType = 3 // 退款：客户要求退回款项。
	AfterSalesTypeRepair      AfterSalesType = 4 // 维修：商品需要维修。
	AfterSalesTypeComplaint   AfterSalesType = 5 // 投诉：客户对商品或服务不满。
)

// AfterSalesStatus 定义了售后请求的生命周期状态。
type AfterSalesStatus int8

const (
	AfterSalesStatusPending    AfterSalesStatus = 1 // 待处理：申请已提交，等待商家审核。
	AfterSalesStatusApproved   AfterSalesStatus = 2 // 已批准：商家已同意售后申请。
	AfterSalesStatusRejected   AfterSalesStatus = 3 // 已拒绝：商家已拒绝售后申请。
	AfterSalesStatusInProgress AfterSalesStatus = 4 // 处理中：售后流程正在进行，例如退货物流中。
	AfterSalesStatusCompleted  AfterSalesStatus = 5 // 已完成：售后流程已结束。
	AfterSalesStatusCancelled  AfterSalesStatus = 6 // 已取消：售后申请被取消。
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
	case AfterSalesStatusCompleted:
		return "Completed"
	case AfterSalesStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// AfterSales 实体是售后模块的聚合根。
// 它代表一个完整的售后申请，包含售后单号、订单信息、用户、类型、状态、原因、商品列表和操作日志等。
type AfterSales struct {
	gorm.Model                        // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	AfterSalesNo    string            `gorm:"type:varchar(64);uniqueIndex;not null;comment:售后单号" json:"after_sales_no"` // 售后单的唯一编号，唯一索引。
	OrderID         uint64            `gorm:"not null;index;comment:订单ID" json:"order_id"`                              // 关联的订单ID，索引字段。
	OrderNo         string            `gorm:"type:varchar(64);not null;comment:订单编号" json:"order_no"`                   // 关联的订单编号。
	UserID          uint64            `gorm:"not null;index;comment:用户ID" json:"user_id"`                               // 发起售后申请的用户ID，索引字段。
	Type            AfterSalesType    `gorm:"type:tinyint;not null;comment:售后类型" json:"type"`                           // 售后类型。
	Status          AfterSalesStatus  `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`                 // 售后单状态，默认为待处理。
	Reason          string            `gorm:"type:varchar(255);not null;comment:申请原因" json:"reason"`                    // 客户提交的申请原因。
	Description     string            `gorm:"type:text;comment:详细描述" json:"description"`                                // 详细的售后描述。
	Images          []string          `gorm:"type:json;serializer:json;comment:凭证图片" json:"images"`                     // 客户提供的凭证图片URL列表。
	RefundAmount    int64             `gorm:"not null;default:0;comment:退款金额(分)" json:"refund_amount"`                  // 订单中实际产生的退款金额（总额），GORM会忽略，仅在代码中记录。
	ApprovalAmount  int64             `gorm:"not null;default:0;comment:批准金额(分)" json:"approval_amount"`                // 实际批准的退款金额或补偿金额。
	ApprovedBy      string            `gorm:"type:varchar(64);comment:批准人" json:"approved_by"`                          // 批准售后申请的操作人员。
	RejectionReason string            `gorm:"type:varchar(255);comment:拒绝原因" json:"rejection_reason"`                   // 拒绝售后申请的原因。
	ApprovedAt      *time.Time        `gorm:"comment:批准时间" json:"approved_at"`                                          // 售后申请被批准的时间。
	RejectedAt      *time.Time        `gorm:"comment:拒绝时间" json:"rejected_at"`                                          // 售后申请被拒绝的时间。
	CompletedAt     *time.Time        `gorm:"comment:完成时间" json:"completed_at"`                                         // 售后流程完成的时间。
	CancelledAt     *time.Time        `gorm:"comment:取消时间" json:"cancelled_at"`                                         // 售后申请被取消的时间。
	Items           []*AfterSalesItem `gorm:"foreignKey:AfterSalesID" json:"items"`                                     // 售后申请包含的商品列表，一对多关系。
	Logs            []*AfterSalesLog  `gorm:"foreignKey:AfterSalesID" json:"logs"`                                      // 售后申请的操作日志列表，一对多关系。
}

// AfterSalesItem 实体代表售后申请中的一个商品项。
type AfterSalesItem struct {
	gorm.Model          // 嵌入gorm.Model。
	AfterSalesID uint64 `gorm:"not null;index;comment:售后单ID" json:"after_sales_id"`          // 关联的售后单ID，索引字段。
	ProductID    uint64 `gorm:"not null;comment:商品ID" json:"product_id"`                     // 商品ID。
	SkuID        uint64 `gorm:"not null;comment:SKU ID" json:"sku_id"`                       // SKU ID。
	ProductName  string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"` // 商品名称。
	SkuName      string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`    // SKU名称（例如，颜色、尺码）。
	Quantity     int32  `gorm:"not null;comment:数量" json:"quantity"`                         // 申请售后的商品数量。
	Price        int64  `gorm:"not null;comment:单价(分)" json:"price"`                         // 商品的单价（单位：分）。
	TotalPrice   int64  `gorm:"not null;comment:总价(分)" json:"total_price"`                   // 商品项的总价（单价 * 数量）。
}

// AfterSalesLog 实体代表售后单的某次操作日志。
// 它记录了操作人、操作动作、状态变更和备注等信息，用于追踪售后流程。
type AfterSalesLog struct {
	gorm.Model          // 嵌入gorm.Model。
	AfterSalesID uint64 `gorm:"not null;index;comment:售后单ID" json:"after_sales_id"`    // 关联的售后单ID，索引字段。
	Operator     string `gorm:"type:varchar(64);not null;comment:操作人" json:"operator"` // 执行操作的人员（例如，用户或管理员）。
	Action       string `gorm:"type:varchar(64);not null;comment:动作" json:"action"`    // 执行的操作类型，例如 "Create", "Approve", "Reject"。
	OldStatus    string `gorm:"type:varchar(32);comment:旧状态" json:"old_status"`        // 操作前的售后单状态。
	NewStatus    string `gorm:"type:varchar(32);comment:新状态" json:"new_status"`        // 操作后的售后单状态。
	Remark       string `gorm:"type:varchar(255);comment:备注" json:"remark"`            // 操作的备注信息。
}

// NewAfterSales 创建并返回一个新的 AfterSales 实体实例。
// afterSalesNo: 售后单号。
// orderID, orderNo, userID: 关联的订单和用户信息。
// afterSalesType: 售后类型。
// reason, description, images: 申请详情。
func NewAfterSales(afterSalesNo string, orderID uint64, orderNo string, userID uint64, afterSalesType AfterSalesType, reason, description string, images []string) *AfterSales {
	return &AfterSales{
		AfterSalesNo: afterSalesNo,
		OrderID:      orderID,
		OrderNo:      orderNo,
		UserID:       userID,
		Type:         afterSalesType,
		Status:       AfterSalesStatusPending, // 新创建的售后单默认为待处理状态。
		Reason:       reason,
		Description:  description,
		Images:       images,
		Items:        []*AfterSalesItem{}, // 初始化商品列表。
		Logs:         []*AfterSalesLog{},  // 初始化日志列表。
	}
}

// Approve 批准售后申请，更新售后单状态为“已批准”，并记录批准人、批准金额和批准时间。
func (a *AfterSales) Approve(operator string, amount int64) {
	a.Status = AfterSalesStatusApproved // 状态更新为“已批准”。
	a.ApprovedBy = operator             // 记录批准人。
	a.ApprovalAmount = amount           // 记录批准金额。
	now := time.Now()
	a.ApprovedAt = &now // 记录批准时间。
}

// Reject 拒绝售后申请，更新售后单状态为“已拒绝”，并记录拒绝人和拒绝原因。
func (a *AfterSales) Reject(operator, reason string) {
	a.Status = AfterSalesStatusRejected // 状态更新为“已拒绝”。
	a.RejectionReason = reason          // 记录拒绝原因。
	now := time.Now()
	a.RejectedAt = &now // 记录拒绝时间。
}

// Process 开始处理售后申请，更新售后单状态为“处理中”。
func (a *AfterSales) Process() {
	a.Status = AfterSalesStatusInProgress // 状态更新为“处理中”。
}

// Complete 完成售后申请，更新售后单状态为“已完成”，并记录完成时间。
func (a *AfterSales) Complete() {
	a.Status = AfterSalesStatusCompleted // 状态更新为“已完成”。
	now := time.Now()
	a.CompletedAt = &now // 记录完成时间。
}

// Cancel 取消售后申请，更新售后单状态为“已取消”，并记录取消时间。
// Cancel 取消售后申请，更新售后单状态为“已取消”，并记录取消时间。
func (a *AfterSales) Cancel() {
	a.Status = AfterSalesStatusCancelled // 状态更新为“已取消”。
	now := time.Now()
	a.CancelledAt = &now // 记录取消时间。
}

// SupportTicketStatus 定义工单状态。
type SupportTicketStatus int8

const (
	SupportTicketStatusOpen     SupportTicketStatus = 1
	SupportTicketStatusPending  SupportTicketStatus = 2
	SupportTicketStatusResolved SupportTicketStatus = 3
	SupportTicketStatusClosed   SupportTicketStatus = 4
)

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
	gorm.Model
	TicketNo    string                  `gorm:"type:varchar(64);uniqueIndex;not null;comment:工单编号"`
	UserID      uint64                  `gorm:"not null;index;comment:用户ID"`
	OrderID     uint64                  `gorm:"index;comment:关联订单ID"`
	Subject     string                  `gorm:"type:varchar(255);not null;comment:主题"`
	Description string                  `gorm:"type:text;comment:描述"`
	Status      SupportTicketStatus     `gorm:"type:tinyint;not null;default:1;comment:状态"`
	Priority    int8                    `gorm:"type:tinyint;default:1;comment:优先级"` // 1: Low, 2: Medium, 3: High
	Category    string                  `gorm:"type:varchar(64);comment:分类"`
	Messages    []*SupportTicketMessage `gorm:"foreignKey:TicketID"`
}

// SupportTicketMessage 工单消息实体。
type SupportTicketMessage struct {
	gorm.Model
	TicketID   uint64 `gorm:"not null;index;comment:工单ID"`
	SenderID   uint64 `gorm:"not null;comment:发送者ID"`                  // 0 表示系统/客服，>0 表示用户
	SenderType string `gorm:"type:varchar(32);not null;comment:发送者类型"` // User, Agent, System
	Content    string `gorm:"type:text;not null;comment:消息内容"`
	IsRead     bool   `gorm:"default:false;comment:是否已读"`
}

// AfterSalesConfig 售后配置实体。
type AfterSalesConfig struct {
	gorm.Model
	Key         string `gorm:"type:varchar(128);uniqueIndex;not null;comment:配置键"`
	Value       string `gorm:"type:text;comment:配置值"`
	Description string `gorm:"type:varchar(255);comment:描述"`
}
