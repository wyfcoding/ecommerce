package model

import "time"

// RefundType 退款类型
type RefundType string

const (
	RefundTypeOnlyRefund RefundType = "ONLY_REFUND" // 仅退款
	RefundTypeRefundGoods RefundType = "REFUND_GOODS" // 退货退款
)

// RefundStatus 退款状态
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "PENDING"    // 待审核
	RefundStatusApproved  RefundStatus = "APPROVED"   // 已同意
	RefundStatusRejected  RefundStatus = "REJECTED"   // 已拒绝
	RefundStatusReturning RefundStatus = "RETURNING"  // 退货中
	RefundStatusReturned  RefundStatus = "RETURNED"   // 已退货
	RefundStatusRefunding RefundStatus = "REFUNDING"  // 退款中
	RefundStatusCompleted RefundStatus = "COMPLETED"  // 已完成
	RefundStatusCancelled RefundStatus = "CANCELLED"  // 已取消
)

// RefundOrder 退款订单
type RefundOrder struct {
	ID              uint64       `gorm:"primarykey" json:"id"`
	RefundNo        string       `gorm:"uniqueIndex;type:varchar(100);not null;comment:退款单号" json:"refundNo"`
	OrderID         uint64       `gorm:"index;not null;comment:订单ID" json:"orderId"`
	OrderNo         string       `gorm:"index;type:varchar(100);not null;comment:订单号" json:"orderNo"`
	UserID          uint64       `gorm:"index;not null;comment:用户ID" json:"userId"`
	Type            RefundType   `gorm:"type:varchar(20);not null;comment:退款类型" json:"type"`
	Status          RefundStatus `gorm:"type:varchar(20);not null;comment:退款状态" json:"status"`
	RefundAmount    uint64       `gorm:"not null;comment:退款金额(分)" json:"refundAmount"`
	RefundReason    string       `gorm:"type:varchar(500);not null;comment:退款原因" json:"refundReason"`
	RefundDesc      string       `gorm:"type:text;comment:退款说明" json:"refundDesc"`
	Images          string       `gorm:"type:text;comment:凭证图片,逗号分隔" json:"images"`
	
	// 审核信息
	ReviewerID      uint64       `gorm:"comment:审核人ID" json:"reviewerId"`
	ReviewTime      *time.Time   `gorm:"comment:审核时间" json:"reviewTime"`
	ReviewRemark    string       `gorm:"type:varchar(500);comment:审核备注" json:"reviewRemark"`
	
	// 退货信息
	ReturnLogistics string       `gorm:"type:varchar(100);comment:退货物流公司" json:"returnLogistics"`
	ReturnTrackingNo string      `gorm:"type:varchar(100);comment:退货物流单号" json:"returnTrackingNo"`
	ReturnTime      *time.Time   `gorm:"comment:退货时间" json:"returnTime"`
	
	// 退款信息
	RefundTime      *time.Time   `gorm:"comment:退款时间" json:"refundTime"`
	RefundChannel   string       `gorm:"type:varchar(50);comment:退款渠道" json:"refundChannel"`
	
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (RefundOrder) TableName() string {
	return "refund_orders"
}

// ExchangeStatus 换货状态
type ExchangeStatus string

const (
	ExchangeStatusPending   ExchangeStatus = "PENDING"    // 待审核
	ExchangeStatusApproved  ExchangeStatus = "APPROVED"   // 已同意
	ExchangeStatusRejected  ExchangeStatus = "REJECTED"   // 已拒绝
	ExchangeStatusReturning ExchangeStatus = "RETURNING"  // 退货中
	ExchangeStatusReturned  ExchangeStatus = "RETURNED"   // 已退货
	ExchangeStatusShipping  ExchangeStatus = "SHIPPING"   // 发货中
	ExchangeStatusCompleted ExchangeStatus = "COMPLETED"  // 已完成
	ExchangeStatusCancelled ExchangeStatus = "CANCELLED"  // 已取消
)

// ExchangeOrder 换货订单
type ExchangeOrder struct {
	ID                uint64         `gorm:"primarykey" json:"id"`
	ExchangeNo        string         `gorm:"uniqueIndex;type:varchar(100);not null;comment:换货单号" json:"exchangeNo"`
	OrderID           uint64         `gorm:"index;not null;comment:订单ID" json:"orderId"`
	OrderNo           string         `gorm:"index;type:varchar(100);not null;comment:订单号" json:"orderNo"`
	UserID            uint64         `gorm:"index;not null;comment:用户ID" json:"userId"`
	Status            ExchangeStatus `gorm:"type:varchar(20);not null;comment:换货状态" json:"status"`
	
	// 原商品信息
	OldProductID      uint64         `gorm:"not null;comment:原商品ID" json:"oldProductId"`
	OldSKUID          uint64         `gorm:"not null;comment:原SKU ID" json:"oldSkuId"`
	OldProductName    string         `gorm:"type:varchar(255);comment:原商品名称" json:"oldProductName"`
	
	// 新商品信息
	NewProductID      uint64         `gorm:"not null;comment:新商品ID" json:"newProductId"`
	NewSKUID          uint64         `gorm:"not null;comment:新SKU ID" json:"newSkuId"`
	NewProductName    string         `gorm:"type:varchar(255);comment:新商品名称" json:"newProductName"`
	
	Quantity          uint32         `gorm:"not null;comment:换货数量" json:"quantity"`
	ExchangeReason    string         `gorm:"type:varchar(500);not null;comment:换货原因" json:"exchangeReason"`
	ExchangeDesc      string         `gorm:"type:text;comment:换货说明" json:"exchangeDesc"`
	Images            string         `gorm:"type:text;comment:凭证图片,逗号分隔" json:"images"`
	
	// 审核信息
	ReviewerID        uint64         `gorm:"comment:审核人ID" json:"reviewerId"`
	ReviewTime        *time.Time     `gorm:"comment:审核时间" json:"reviewTime"`
	ReviewRemark      string         `gorm:"type:varchar(500);comment:审核备注" json:"reviewRemark"`
	
	// 退货信息
	ReturnLogistics   string         `gorm:"type:varchar(100);comment:退货物流公司" json:"returnLogistics"`
	ReturnTrackingNo  string         `gorm:"type:varchar(100);comment:退货物流单号" json:"returnTrackingNo"`
	ReturnTime        *time.Time     `gorm:"comment:退货时间" json:"returnTime"`
	
	// 发货信息
	ShipLogistics     string         `gorm:"type:varchar(100);comment:发货物流公司" json:"shipLogistics"`
	ShipTrackingNo    string         `gorm:"type:varchar(100);comment:发货物流单号" json:"shipTrackingNo"`
	ShipTime          *time.Time     `gorm:"comment:发货时间" json:"shipTime"`
	
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (ExchangeOrder) TableName() string {
	return "exchange_orders"
}

// RepairStatus 维修状态
type RepairStatus string

const (
	RepairStatusPending   RepairStatus = "PENDING"    // 待审核
	RepairStatusApproved  RepairStatus = "APPROVED"   // 已同意
	RepairStatusRejected  RepairStatus = "REJECTED"   // 已拒绝
	RepairStatusReturning RepairStatus = "RETURNING"  // 寄回中
	RepairStatusRepairing RepairStatus = "REPAIRING"  // 维修中
	RepairStatusShipping  RepairStatus = "SHIPPING"   // 寄出中
	RepairStatusCompleted RepairStatus = "COMPLETED"  // 已完成
	RepairStatusCancelled RepairStatus = "CANCELLED"  // 已取消
)

// RepairOrder 维修订单
type RepairOrder struct {
	ID               uint64       `gorm:"primarykey" json:"id"`
	RepairNo         string       `gorm:"uniqueIndex;type:varchar(100);not null;comment:维修单号" json:"repairNo"`
	OrderID          uint64       `gorm:"index;not null;comment:订单ID" json:"orderId"`
	OrderNo          string       `gorm:"index;type:varchar(100);not null;comment:订单号" json:"orderNo"`
	UserID           uint64       `gorm:"index;not null;comment:用户ID" json:"userId"`
	Status           RepairStatus `gorm:"type:varchar(20);not null;comment:维修状态" json:"status"`
	
	ProductID        uint64       `gorm:"not null;comment:商品ID" json:"productId"`
	SKUID            uint64       `gorm:"not null;comment:SKU ID" json:"skuId"`
	ProductName      string       `gorm:"type:varchar(255);comment:商品名称" json:"productName"`
	
	FaultDesc        string       `gorm:"type:text;not null;comment:故障描述" json:"faultDesc"`
	Images           string       `gorm:"type:text;comment:故障图片,逗号分隔" json:"images"`
	
	// 审核信息
	ReviewerID       uint64       `gorm:"comment:审核人ID" json:"reviewerId"`
	ReviewTime       *time.Time   `gorm:"comment:审核时间" json:"reviewTime"`
	ReviewRemark     string       `gorm:"type:varchar(500);comment:审核备注" json:"reviewRemark"`
	
	// 寄回信息
	ReturnLogistics  string       `gorm:"type:varchar(100);comment:寄回物流公司" json:"returnLogistics"`
	ReturnTrackingNo string       `gorm:"type:varchar(100);comment:寄回物流单号" json:"returnTrackingNo"`
	ReturnTime       *time.Time   `gorm:"comment:寄回时间" json:"returnTime"`
	
	// 维修信息
	RepairStartTime  *time.Time   `gorm:"comment:维修开始时间" json:"repairStartTime"`
	RepairEndTime    *time.Time   `gorm:"comment:维修结束时间" json:"repairEndTime"`
	RepairResult     string       `gorm:"type:text;comment:维修结果" json:"repairResult"`
	RepairCost       uint64       `gorm:"comment:维修费用(分)" json:"repairCost"`
	
	// 寄出信息
	ShipLogistics    string       `gorm:"type:varchar(100);comment:寄出物流公司" json:"shipLogistics"`
	ShipTrackingNo   string       `gorm:"type:varchar(100);comment:寄出物流单号" json:"shipTrackingNo"`
	ShipTime         *time.Time   `gorm:"comment:寄出时间" json:"shipTime"`
	
	CreatedAt        time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (RepairOrder) TableName() string {
	return "repair_orders"
}

// AfterSalesTicketStatus 售后工单状态
type AfterSalesTicketStatus string

const (
	TicketStatusOpen       AfterSalesTicketStatus = "OPEN"        // 待处理
	TicketStatusProcessing AfterSalesTicketStatus = "PROCESSING"  // 处理中
	TicketStatusResolved   AfterSalesTicketStatus = "RESOLVED"    // 已解决
	TicketStatusClosed     AfterSalesTicketStatus = "CLOSED"      // 已关闭
)

// AfterSalesTicket 售后工单
type AfterSalesTicket struct {
	ID          uint64                 `gorm:"primarykey" json:"id"`
	TicketNo    string                 `gorm:"uniqueIndex;type:varchar(100);not null;comment:工单号" json:"ticketNo"`
	UserID      uint64                 `gorm:"index;not null;comment:用户ID" json:"userId"`
	OrderID     uint64                 `gorm:"index;comment:关联订单ID" json:"orderId"`
	Type        string                 `gorm:"type:varchar(50);not null;comment:工单类型" json:"type"`
	Status      AfterSalesTicketStatus `gorm:"type:varchar(20);not null;comment:工单状态" json:"status"`
	Priority    string                 `gorm:"type:varchar(20);comment:优先级" json:"priority"`
	Subject     string                 `gorm:"type:varchar(255);not null;comment:工单主题" json:"subject"`
	Description string                 `gorm:"type:text;not null;comment:问题描述" json:"description"`
	Images      string                 `gorm:"type:text;comment:附件图片,逗号分隔" json:"images"`
	
	// 处理信息
	AgentID     uint64                 `gorm:"index;comment:客服ID" json:"agentId"`
	AssignTime  *time.Time             `gorm:"comment:分配时间" json:"assignTime"`
	ResolveTime *time.Time             `gorm:"comment:解决时间" json:"resolveTime"`
	CloseTime   *time.Time             `gorm:"comment:关闭时间" json:"closeTime"`
	Resolution  string                 `gorm:"type:text;comment:解决方案" json:"resolution"`
	
	CreatedAt   time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (AfterSalesTicket) TableName() string {
	return "aftersales_tickets"
}
