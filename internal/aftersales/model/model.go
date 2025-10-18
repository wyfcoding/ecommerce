package model

import (
	"time"

	"gorm.io/gorm"
)

// ApplicationType 定义了售后申请的类型
type ApplicationType string

const (
	TypeReturn   ApplicationType = "RETURN"   // 退货
	TypeExchange ApplicationType = "EXCHANGE" // 换货
	TypeRepair   ApplicationType = "REPAIR"   // 维修
)

// ApplicationStatus 定义了售后申请的当前状态
type ApplicationStatus string

const (
	StatusPendingApproval  ApplicationStatus = "PENDING_APPROVAL"  // 待审核
	StatusApproved         ApplicationStatus = "APPROVED"          // 审核通过 (待用户寄回)
	StatusRejected         ApplicationStatus = "REJECTED"          // 审核拒绝
	StatusGoodsReceived    ApplicationStatus = "GOODS_RECEIVED"    // 已收到退货
	StatusProcessing       ApplicationStatus = "PROCESSING"        // 处理中 (如换货中、退款中)
	StatusCompleted        ApplicationStatus = "COMPLETED"         // 已完成
)

// AftersalesApplication 售后申请主模型
type AftersalesApplication struct {
	ID              uint              `gorm:"primarykey" json:"id"`
	ApplicationSN   string            `gorm:"type:varchar(100);uniqueIndex;not null" json:"application_sn"` // 售后申请单号
	UserID          uint              `gorm:"not null;index" json:"user_id"`
	OrderID         uint              `gorm:"not null;index" json:"order_id"`
	OrderSN         string            `gorm:"type:varchar(100);index" json:"order_sn"`
	Type            ApplicationType   `gorm:"type:varchar(20);not null" json:"type"`
	Status          ApplicationStatus `gorm:"type:varchar(30);not null" json:"status"`
	Reason          string            `gorm:"type:varchar(255);not null" json:"reason"` // 申请原因
	UserRemarks     string            `gorm:"type:text" json:"user_remarks"`             // 用户备注
	AdminRemarks    string            `gorm:"type:text" json:"admin_remarks"`            // 管理员备注 (审核意见)
	RefundAmount    float64           `gorm:"type:decimal(10,2)" json:"refund_amount"`   // 最终退款金额
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`

	Items []AftersalesItem `gorm:"foreignKey:ApplicationID" json:"items"`
}

// AftersalesItem 售后申请的商品项
type AftersalesItem struct {
	ID            uint `gorm:"primarykey" json:"id"`
	ApplicationID uint `gorm:"not null;index" json:"application_id"`
	OrderItemID   uint `gorm:"not null" json:"order_item_id"` // 原始订单项ID
	ProductID     uint `gorm:"not null" json:"product_id"`
	ProductSKU    string `gorm:"type:varchar(100)" json:"product_sku"`
	Quantity      int  `gorm:"not null" json:"quantity"`
}

// TableName 自定义表名
func (AftersalesApplication) TableName() string {
	return "aftersales_applications"
}

func (AftersalesItem) TableName() string {
	return "aftersales_items"
}
