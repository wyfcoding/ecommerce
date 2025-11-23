package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrAfterSalesNotFound = errors.New("after-sales record not found")
	ErrInvalidStatus      = errors.New("invalid status for operation")
)

// AfterSalesType 售后类型
type AfterSalesType int8

const (
	AfterSalesTypeReturnGoods AfterSalesType = 1 // 退货
	AfterSalesTypeExchange    AfterSalesType = 2 // 换货
	AfterSalesTypeRefund      AfterSalesType = 3 // 退款
	AfterSalesTypeRepair      AfterSalesType = 4 // 维修
	AfterSalesTypeComplaint   AfterSalesType = 5 // 投诉
)

// AfterSalesStatus 售后状态
type AfterSalesStatus int8

const (
	AfterSalesStatusPending    AfterSalesStatus = 1 // 待处理
	AfterSalesStatusApproved   AfterSalesStatus = 2 // 已批准
	AfterSalesStatusRejected   AfterSalesStatus = 3 // 已拒绝
	AfterSalesStatusInProgress AfterSalesStatus = 4 // 处理中
	AfterSalesStatusCompleted  AfterSalesStatus = 5 // 已完成
	AfterSalesStatusCancelled  AfterSalesStatus = 6 // 已取消
)

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

// AfterSales 售后聚合根
type AfterSales struct {
	gorm.Model
	AfterSalesNo    string            `gorm:"type:varchar(64);uniqueIndex;not null;comment:售后单号" json:"after_sales_no"`
	OrderID         uint64            `gorm:"not null;index;comment:订单ID" json:"order_id"`
	OrderNo         string            `gorm:"type:varchar(64);not null;comment:订单编号" json:"order_no"`
	UserID          uint64            `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Type            AfterSalesType    `gorm:"type:tinyint;not null;comment:售后类型" json:"type"`
	Status          AfterSalesStatus  `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	Reason          string            `gorm:"type:varchar(255);not null;comment:申请原因" json:"reason"`
	Description     string            `gorm:"type:text;comment:详细描述" json:"description"`
	Images          []string          `gorm:"type:json;serializer:json;comment:凭证图片" json:"images"`
	RefundAmount    int64             `gorm:"not null;default:0;comment:退款金额(分)" json:"refund_amount"`
	ApprovalAmount  int64             `gorm:"not null;default:0;comment:批准金额(分)" json:"approval_amount"`
	ApprovedBy      string            `gorm:"type:varchar(64);comment:批准人" json:"approved_by"`
	RejectionReason string            `gorm:"type:varchar(255);comment:拒绝原因" json:"rejection_reason"`
	ApprovedAt      *time.Time        `gorm:"comment:批准时间" json:"approved_at"`
	RejectedAt      *time.Time        `gorm:"comment:拒绝时间" json:"rejected_at"`
	CompletedAt     *time.Time        `gorm:"comment:完成时间" json:"completed_at"`
	CancelledAt     *time.Time        `gorm:"comment:取消时间" json:"cancelled_at"`
	Items           []*AfterSalesItem `gorm:"foreignKey:AfterSalesID" json:"items"`
	Logs            []*AfterSalesLog  `gorm:"foreignKey:AfterSalesID" json:"logs"`
}

// AfterSalesItem 售后项目实体
type AfterSalesItem struct {
	gorm.Model
	AfterSalesID uint64 `gorm:"not null;index;comment:售后单ID" json:"after_sales_id"`
	ProductID    uint64 `gorm:"not null;comment:商品ID" json:"product_id"`
	SkuID        uint64 `gorm:"not null;comment:SKU ID" json:"sku_id"`
	ProductName  string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`
	SkuName      string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`
	Quantity     int32  `gorm:"not null;comment:数量" json:"quantity"`
	Price        int64  `gorm:"not null;comment:单价(分)" json:"price"`
	TotalPrice   int64  `gorm:"not null;comment:总价(分)" json:"total_price"`
}

// AfterSalesLog 售后日志值对象
type AfterSalesLog struct {
	gorm.Model
	AfterSalesID uint64 `gorm:"not null;index;comment:售后单ID" json:"after_sales_id"`
	Operator     string `gorm:"type:varchar(64);not null;comment:操作人" json:"operator"`
	Action       string `gorm:"type:varchar(64);not null;comment:动作" json:"action"`
	OldStatus    string `gorm:"type:varchar(32);comment:旧状态" json:"old_status"`
	NewStatus    string `gorm:"type:varchar(32);comment:新状态" json:"new_status"`
	Remark       string `gorm:"type:varchar(255);comment:备注" json:"remark"`
}

// NewAfterSales 创建售后聚合根
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
		Items:        []*AfterSalesItem{},
		Logs:         []*AfterSalesLog{},
	}
}

// Approve 批准售后
func (a *AfterSales) Approve(operator string, amount int64) {
	a.Status = AfterSalesStatusApproved
	a.ApprovedBy = operator
	a.ApprovalAmount = amount
	now := time.Now()
	a.ApprovedAt = &now
}

// Reject 拒绝售后
func (a *AfterSales) Reject(operator, reason string) {
	a.Status = AfterSalesStatusRejected
	a.RejectionReason = reason
	now := time.Now()
	a.RejectedAt = &now
}

// Process 开始处理
func (a *AfterSales) Process() {
	a.Status = AfterSalesStatusInProgress
}

// Complete 完成售后
func (a *AfterSales) Complete() {
	a.Status = AfterSalesStatusCompleted
	now := time.Now()
	a.CompletedAt = &now
}

// Cancel 取消售后
func (a *AfterSales) Cancel() {
	a.Status = AfterSalesStatusCancelled
	now := time.Now()
	a.CancelledAt = &now
}
