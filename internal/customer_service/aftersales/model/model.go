package model

import (
	"time"

	"gorm.io/gorm"
)

// ApplicationType 定义了售后申请的类型。
// 包括退货、换货和维修等。
type ApplicationType string

const (
	TypeReturn   ApplicationType = "RETURN"   // 退货：用户将商品退回商家并获得退款。
	TypeExchange ApplicationType = "EXCHANGE" // 换货：用户将商品退回商家并更换为其他商品。
	TypeRepair   ApplicationType = "REPAIR"   // 维修：商品出现故障，需要商家提供维修服务。
)

// ApplicationStatus 定义了售后申请的当前状态。
// 描述了申请从提交到完成的整个生命周期。
type ApplicationStatus string

const (
	StatusPendingApproval  ApplicationStatus = "PENDING_APPROVAL"  // 待审核：申请已提交，等待管理员审核。
	StatusApproved         ApplicationStatus = "APPROVED"          // 审核通过：申请已通过审核，等待用户寄回商品或进行下一步处理。
	StatusRejected         ApplicationStatus = "REJECTED"          // 审核拒绝：申请未通过审核。
	StatusGoodsReceived    ApplicationStatus = "GOODS_RECEIVED"    // 已收到退货：商家已收到用户寄回的商品。
	StatusProcessing       ApplicationStatus = "PROCESSING"        // 处理中：退款、换货或维修正在进行中。
	StatusCompleted        ApplicationStatus = "COMPLETED"         // 已完成：售后流程已全部完成。
	StatusCancelled        ApplicationStatus = "CANCELLED"         // 已取消：申请在处理前被用户或管理员取消。
)

// AftersalesApplication 售后申请主模型。
// 记录了用户提交的售后请求的详细信息。
type AftersalesApplication struct {
	ID              uint              `gorm:"primarykey" json:"id"`                                         // 售后申请的唯一标识符
	ApplicationSN   string            `gorm:"type:varchar(100);uniqueIndex;not null" json:"application_sn"` // 售后申请单号，唯一
	UserID          uint              `gorm:"not null;index" json:"user_id"`                                // 提交申请的用户ID
	OrderID         uint              `gorm:"not null;index" json:"order_id"`                               // 关联的订单ID
	OrderSN         string            `gorm:"type:varchar(100);index" json:"order_sn"`                      // 关联的订单号
	Type            ApplicationType   `gorm:"type:varchar(20);not null" json:"type"`                        // 售后申请类型 (退货、换货、维修)
	Status          ApplicationStatus `gorm:"type:varchar(30);not null" json:"status"`                      // 售后申请当前状态
	Reason          string            `gorm:"type:varchar(255);not null" json:"reason"`                     // 用户提交的申请原因
	UserRemarks     string            `gorm:"type:text" json:"user_remarks"`                                // 用户备注信息
	AdminRemarks    string            `gorm:"type:text" json:"admin_remarks"`                               // 管理员审核或处理时的备注信息
	RefundAmount    float64           `gorm:"type:decimal(10,2)" json:"refund_amount"`                      // 最终退款金额 (仅退货/退款申请相关)
	CreatedAt       time.Time         `json:"created_at"`                                                   // 申请创建时间
	UpdatedAt       time.Time         `json:"updated_at"`                                                   // 申请最后更新时间

	Items []AftersalesItem `gorm:"foreignKey:ApplicationID" json:"items"` // 关联的售后商品项列表
}

// AftersalesItem 售后申请的商品项。
// 记录了售后申请中涉及的具体商品及其数量。
type AftersalesItem struct {
	ID            uint   `gorm:"primarykey" json:"id"`                 // 售后商品项的唯一标识符
	ApplicationID uint   `gorm:"not null;index" json:"application_id"` // 所属售后申请的ID
	OrderItemID   uint   `gorm:"not null" json:"order_item_id"`        // 原始订单项的ID
	ProductID     uint   `gorm:"not null" json:"product_id"`           // 商品ID
	ProductSKU    string `gorm:"type:varchar(100)" json:"product_sku"` // 商品SKU
	Quantity      int    `gorm:"not null" json:"quantity"`             // 售后数量
}

// TableName 自定义 AftersalesApplication 对应的表名。
func (AftersalesApplication) TableName() string {
	return "aftersales_applications"
}

// TableName 自定义 AftersalesItem 对应的表名。
func (AftersalesItem) TableName() string {
	return "aftersales_items"
}