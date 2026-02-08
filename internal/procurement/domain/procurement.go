package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PurchaseRequestStatus 采购申请状态
type PurchaseRequestStatus int8

const (
	PRStatusPending  PurchaseRequestStatus = 1 // 待审批
	PRStatusApproved PurchaseRequestStatus = 2 // 已批准
	PRStatusRejected PurchaseRequestStatus = 3 // 已拒绝
	PRStatusOrdered  PurchaseRequestStatus = 4 // 已下单
)

// PurchaseOrderStatus 采购单状态
type PurchaseOrderStatus int8

const (
	POStatusPending   PurchaseOrderStatus = 1 // 待确认
	POStatusConfirmed PurchaseOrderStatus = 2 // 供应商已通过
	POStatusShipped   PurchaseOrderStatus = 3 // 供应商已发货
	POStatusReceived  PurchaseOrderStatus = 4 // 仓库已入库
	POStatusCompleted PurchaseOrderStatus = 5 // 已完成
	POStatusCancelled PurchaseOrderStatus = 6 // 已取消
)

func (s PurchaseOrderStatus) String() string {
	switch s {
	case POStatusPending:
		return "PENDING"
	case POStatusConfirmed:
		return "CONFIRMED"
	case POStatusShipped:
		return "SHIPPED"
	case POStatusReceived:
		return "RECEIVED"
	case POStatusCompleted:
		return "COMPLETED"
	case POStatusCancelled:
		return "CANCELLED"
	}
	return "UNKNOWN"
}

// PurchaseRequest 采购申请
type PurchaseRequest struct {
	gorm.Model
	RequestID   string                `gorm:"column:request_id;type:varchar(32);unique_index;not null"`
	ApplicantID string                `gorm:"column:applicant_id;type:varchar(32);index;not null"`
	Reason      string                `gorm:"column:reason;type:varchar(255)"`
	Status      PurchaseRequestStatus `gorm:"column:status;type:tinyint;not null;default:1"`
	ApproverID  string                `gorm:"column:approver_id;type:varchar(32)"`
	Comment     string                `gorm:"column:comment;type:varchar(255)"`
	ApprovedAt  *time.Time            `gorm:"column:approved_at"`

	Items []PurchaseRequestItem `gorm:"foreignKey:RequestID;references:RequestID"`
}

type PurchaseRequestItem struct {
	gorm.Model
	RequestID    string `gorm:"column:request_id;type:varchar(32);index;not null"`
	SKUID        string `gorm:"column:sku_id;type:varchar(64);not null"`
	ProductName  string `gorm:"column:product_name;type:varchar(255)"`
	Quantity     int32  `gorm:"column:quantity;not null"`
	ExpectedDate string `gorm:"column:expected_date;type:varchar(10)"` // YYYY-MM-DD
}

// PurchaseOrder 采购单
type PurchaseOrder struct {
	gorm.Model
	OrderID           string              `gorm:"column:order_id;type:varchar(32);unique_index;not null"`
	PurchaseRequestID string              `gorm:"column:purchase_request_id;type:varchar(32);index"`
	SupplierID        string              `gorm:"column:supplier_id;type:varchar(32);index;not null"`
	WarehouseID       string              `gorm:"column:warehouse_id;type:varchar(32);not null"`
	TotalAmount       decimal.Decimal     `gorm:"column:total_amount;type:decimal(20,2);not null"`
	Status            PurchaseOrderStatus `gorm:"column:status;type:tinyint;not null;default:1"`
	Remark            string              `gorm:"column:remark;type:text"`

	Items []PurchaseOrderItem `gorm:"foreignKey:OrderID;references:OrderID"`

	domainEvents []DomainEvent `gorm:"-"`
}

type PurchaseOrderItem struct {
	gorm.Model
	OrderID     string          `gorm:"column:order_id;type:varchar(32);index;not null"`
	SKUID       string          `gorm:"column:sku_id;type:varchar(64);not null"`
	ProductName string          `gorm:"column:product_name;type:varchar(255)"`
	Quantity    int32           `gorm:"column:quantity;not null"`
	UnitPrice   decimal.Decimal `gorm:"column:unit_price;type:decimal(20,2);not null"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(20,2);not null"`
}

func (PurchaseRequest) TableName() string     { return "purchase_requests" }
func (PurchaseRequestItem) TableName() string { return "purchase_request_items" }
func (PurchaseOrder) TableName() string       { return "purchase_orders" }
func (PurchaseOrderItem) TableName() string   { return "purchase_order_items" }

// NewPurchaseOrder 创建采购单
func NewPurchaseOrder(id, prID, supplierID, warehouseID, remark string) *PurchaseOrder {
	return &PurchaseOrder{
		OrderID:           id,
		PurchaseRequestID: prID,
		SupplierID:        supplierID,
		WarehouseID:       warehouseID,
		Status:            POStatusPending,
		Remark:            remark,
		TotalAmount:       decimal.Zero,
	}
}

func (po *PurchaseOrder) AddItem(skuID, name string, qty int32, price decimal.Decimal) {
	total := price.Mul(decimal.NewFromInt32(qty))
	po.Items = append(po.Items, PurchaseOrderItem{
		OrderID:     po.OrderID,
		SKUID:       skuID,
		ProductName: name,
		Quantity:    qty,
		UnitPrice:   price,
		TotalAmount: total,
	})
	po.TotalAmount = po.TotalAmount.Add(total)
}

func (po *PurchaseOrder) Confirm() error {
	if po.Status != POStatusPending {
		return errors.New("invalid status transition")
	}
	po.Status = POStatusConfirmed
	po.addEvent(&PurchaseOrderConfirmedEvent{OrderID: po.OrderID, Timestamp: time.Now()})
	return nil
}

func (po *PurchaseOrder) Ship() error {
	if po.Status != POStatusConfirmed {
		return errors.New("invalid status transition")
	}
	po.Status = POStatusShipped
	return nil
}

func (po *PurchaseOrder) Receive() error {
	if po.Status != POStatusShipped {
		return errors.New("invalid status transition")
	}
	po.Status = POStatusReceived
	po.addEvent(&PurchaseOrderReceivedEvent{OrderID: po.OrderID, Timestamp: time.Now()})
	return nil
}

func (po *PurchaseOrder) addEvent(e DomainEvent) {
	po.domainEvents = append(po.domainEvents, e)
}

func (po *PurchaseOrder) GetDomainEvents() []DomainEvent {
	return po.domainEvents
}

func (po *PurchaseOrder) ClearDomainEvents() {
	po.domainEvents = nil
}
