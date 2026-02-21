package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"gorm.io/gorm"
)

// StringArray JSON 序列化字符串数组。
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// AfterSalesModel 售后写模型。
type AfterSalesModel struct {
	gorm.Model
	AfterSalesNo    string                  `gorm:"type:varchar(64);uniqueIndex;not null;comment:售后单号"`
	OrderID         uint64                  `gorm:"not null;index;comment:订单ID"`
	OrderNo         string                  `gorm:"type:varchar(64);not null;comment:订单编号"`
	UserID          uint64                  `gorm:"not null;index;comment:用户ID"`
	Type            domain.AfterSalesType   `gorm:"type:tinyint;not null;comment:售后类型"`
	Status          domain.AfterSalesStatus `gorm:"type:tinyint;not null;default:1;comment:状态"`
	Reason          string                  `gorm:"type:varchar(255);not null;comment:申请原因"`
	Description     string                  `gorm:"type:text;comment:详细描述"`
	Images          StringArray             `gorm:"type:json;comment:凭证图片"`
	RefundAmount    int64                   `gorm:"not null;default:0;comment:退款金额(分)"`
	ApprovalAmount  int64                   `gorm:"not null;default:0;comment:批准金额(分)"`
	ApprovedBy      string                  `gorm:"type:varchar(64);comment:批准人"`
	RejectionReason string                  `gorm:"type:varchar(255);comment:拒绝原因"`
	ApprovedAt      *time.Time              `gorm:"comment:批准时间"`
	RejectedAt      *time.Time              `gorm:"comment:拒绝时间"`
	CompletedAt     *time.Time              `gorm:"comment:完成时间"`
	CancelledAt     *time.Time              `gorm:"comment:取消时间"`
	TrackingNumber  string                  `gorm:"type:varchar(64);comment:物流单号"`
	WarehouseNotes  string                  `gorm:"type:text;comment:仓库备注"`
	RMANumber       string                  `gorm:"type:varchar(64);comment:退货授权号"`
}

// AfterSalesItemModel 售后商品项写模型。
type AfterSalesItemModel struct {
	gorm.Model
	AfterSalesID uint64      `gorm:"not null;index;comment:售后单ID"`
	ProductID    uint64      `gorm:"not null;comment:商品ID"`
	SkuID        uint64      `gorm:"not null;comment:SKU ID"`
	ProductName  string      `gorm:"type:varchar(255);not null;comment:商品名称"`
	SkuName      string      `gorm:"type:varchar(255);not null;comment:SKU名称"`
	Quantity     int32       `gorm:"not null;comment:数量"`
	Price        int64       `gorm:"not null;comment:单价(分)"`
	TotalPrice   int64       `gorm:"not null;comment:总价(分)"`
	Reason       string      `gorm:"type:varchar(255);comment:售后原因"`
	Images       StringArray `gorm:"type:json;comment:商品图片"`
}

// AfterSalesLogModel 售后操作日志写模型。
type AfterSalesLogModel struct {
	gorm.Model
	AfterSalesID uint64 `gorm:"not null;index;comment:售后单ID"`
	Operator     string `gorm:"type:varchar(64);not null;comment:操作人"`
	Action       string `gorm:"type:varchar(64);not null;comment:动作"`
	OldStatus    string `gorm:"type:varchar(32);comment:旧状态"`
	NewStatus    string `gorm:"type:varchar(32);comment:新状态"`
	Remark       string `gorm:"type:varchar(255);comment:备注"`
}

// SupportTicketModel 客服工单写模型。
type SupportTicketModel struct {
	gorm.Model
	TicketNo    string                     `gorm:"type:varchar(64);uniqueIndex;not null;comment:工单编号"`
	UserID      uint64                     `gorm:"not null;index;comment:用户ID"`
	OrderID     uint64                     `gorm:"index;comment:关联订单ID"`
	Subject     string                     `gorm:"type:varchar(255);not null;comment:主题"`
	Description string                     `gorm:"type:text;comment:描述"`
	Status      domain.SupportTicketStatus `gorm:"type:tinyint;not null;default:1;comment:状态"`
	Priority    int8                       `gorm:"type:tinyint;default:1;comment:优先级"`
	Category    string                     `gorm:"type:varchar(64);comment:分类"`
}

// SupportTicketMessageModel 客服工单消息写模型。
type SupportTicketMessageModel struct {
	gorm.Model
	TicketID   uint64 `gorm:"not null;index;comment:工单ID"`
	SenderID   uint64 `gorm:"not null;comment:发送者ID"`
	SenderType string `gorm:"type:varchar(32);not null;comment:发送者类型"`
	Content    string `gorm:"type:text;not null;comment:消息内容"`
	IsRead     bool   `gorm:"default:false;comment:是否已读"`
}

// AfterSalesConfigModel 售后配置写模型。
type AfterSalesConfigModel struct {
	gorm.Model
	Key         string `gorm:"type:varchar(128);uniqueIndex;not null;comment:配置键"`
	Value       string `gorm:"type:text;comment:配置值"`
	Description string `gorm:"type:varchar(255);comment:描述"`
}

func (AfterSalesModel) TableName() string           { return "after_sales" }
func (AfterSalesItemModel) TableName() string       { return "after_sales_items" }
func (AfterSalesLogModel) TableName() string        { return "after_sales_logs" }
func (SupportTicketModel) TableName() string        { return "support_tickets" }
func (SupportTicketMessageModel) TableName() string { return "support_ticket_messages" }
func (AfterSalesConfigModel) TableName() string     { return "after_sales_configs" }

func toAfterSalesModel(a *domain.AfterSales) *AfterSalesModel {
	if a == nil {
		return nil
	}
	return &AfterSalesModel{
		Model: gorm.Model{
			ID:        uint(a.ID),
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		},
		AfterSalesNo:    a.AfterSalesNo,
		OrderID:         a.OrderID,
		OrderNo:         a.OrderNo,
		UserID:          a.UserID,
		Type:            a.Type,
		Status:          a.Status,
		Reason:          a.Reason,
		Description:     a.Description,
		Images:          StringArray(a.Images),
		RefundAmount:    a.RefundAmount,
		ApprovalAmount:  a.ApprovalAmount,
		ApprovedBy:      a.ApprovedBy,
		RejectionReason: a.RejectionReason,
		ApprovedAt:      a.ApprovedAt,
		RejectedAt:      a.RejectedAt,
		CompletedAt:     a.CompletedAt,
		CancelledAt:     a.CancelledAt,
		TrackingNumber:  a.TrackingNumber,
		WarehouseNotes:  a.WarehouseNotes,
		RMANumber:       a.RMANumber,
	}
}

func toAfterSales(model *AfterSalesModel) *domain.AfterSales {
	if model == nil {
		return nil
	}
	items := make([]*domain.AfterSalesItem, 0)
	logs := make([]*domain.AfterSalesLog, 0)
	return &domain.AfterSales{
		ID:              uint64(model.ID),
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		AfterSalesNo:    model.AfterSalesNo,
		OrderID:         model.OrderID,
		OrderNo:         model.OrderNo,
		UserID:          model.UserID,
		Type:            model.Type,
		Status:          model.Status,
		Reason:          model.Reason,
		Description:     model.Description,
		Images:          []string(model.Images),
		RefundAmount:    model.RefundAmount,
		ApprovalAmount:  model.ApprovalAmount,
		ApprovedBy:      model.ApprovedBy,
		RejectionReason: model.RejectionReason,
		ApprovedAt:      model.ApprovedAt,
		RejectedAt:      model.RejectedAt,
		CompletedAt:     model.CompletedAt,
		CancelledAt:     model.CancelledAt,
		TrackingNumber:  model.TrackingNumber,
		WarehouseNotes:  model.WarehouseNotes,
		RMANumber:       model.RMANumber,
		Items:           items,
		Logs:            logs,
	}
}

func toAfterSalesItemModel(item *domain.AfterSalesItem) *AfterSalesItemModel {
	if item == nil {
		return nil
	}
	return &AfterSalesItemModel{
		Model: gorm.Model{
			ID:        uint(item.ID),
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		AfterSalesID: item.AfterSalesID,
		ProductID:    item.ProductID,
		SkuID:        item.SkuID,
		ProductName:  item.ProductName,
		SkuName:      item.SkuName,
		Quantity:     item.Quantity,
		Price:        item.Price,
		TotalPrice:   item.TotalPrice,
		Reason:       item.Reason,
		Images:       StringArray(item.Images),
	}
}

func toAfterSalesItem(model *AfterSalesItemModel) *domain.AfterSalesItem {
	if model == nil {
		return nil
	}
	return &domain.AfterSalesItem{
		ID:           uint64(model.ID),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		AfterSalesID: model.AfterSalesID,
		ProductID:    model.ProductID,
		SkuID:        model.SkuID,
		ProductName:  model.ProductName,
		SkuName:      model.SkuName,
		Quantity:     model.Quantity,
		Price:        model.Price,
		TotalPrice:   model.TotalPrice,
		Reason:       model.Reason,
		Images:       []string(model.Images),
	}
}

func toAfterSalesLogModel(log *domain.AfterSalesLog) *AfterSalesLogModel {
	if log == nil {
		return nil
	}
	return &AfterSalesLogModel{
		Model: gorm.Model{
			ID:        uint(log.ID),
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		AfterSalesID: log.AfterSalesID,
		Operator:     log.Operator,
		Action:       log.Action,
		OldStatus:    log.OldStatus,
		NewStatus:    log.NewStatus,
		Remark:       log.Remark,
	}
}

func toAfterSalesLog(model *AfterSalesLogModel) *domain.AfterSalesLog {
	if model == nil {
		return nil
	}
	return &domain.AfterSalesLog{
		ID:           uint64(model.ID),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		AfterSalesID: model.AfterSalesID,
		Operator:     model.Operator,
		Action:       model.Action,
		OldStatus:    model.OldStatus,
		NewStatus:    model.NewStatus,
		Remark:       model.Remark,
	}
}

func toSupportTicketModel(t *domain.SupportTicket) *SupportTicketModel {
	if t == nil {
		return nil
	}
	return &SupportTicketModel{
		Model: gorm.Model{
			ID:        uint(t.ID),
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		},
		TicketNo:    t.TicketNo,
		UserID:      t.UserID,
		OrderID:     t.OrderID,
		Subject:     t.Subject,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Category:    t.Category,
	}
}

func toSupportTicket(model *SupportTicketModel) *domain.SupportTicket {
	if model == nil {
		return nil
	}
	return &domain.SupportTicket{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		TicketNo:    model.TicketNo,
		UserID:      model.UserID,
		OrderID:     model.OrderID,
		Subject:     model.Subject,
		Description: model.Description,
		Status:      model.Status,
		Priority:    model.Priority,
		Category:    model.Category,
		Messages:    []*domain.SupportTicketMessage{},
	}
}

func toSupportTicketMessageModel(m *domain.SupportTicketMessage) *SupportTicketMessageModel {
	if m == nil {
		return nil
	}
	return &SupportTicketMessageModel{
		Model: gorm.Model{
			ID:        uint(m.ID),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		TicketID:   m.TicketID,
		SenderID:   m.SenderID,
		SenderType: m.SenderType,
		Content:    m.Content,
		IsRead:     m.IsRead,
	}
}

func toSupportTicketMessage(model *SupportTicketMessageModel) *domain.SupportTicketMessage {
	if model == nil {
		return nil
	}
	return &domain.SupportTicketMessage{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		TicketID:   model.TicketID,
		SenderID:   model.SenderID,
		SenderType: model.SenderType,
		Content:    model.Content,
		IsRead:     model.IsRead,
	}
}

func toConfigModel(cfg *domain.AfterSalesConfig) *AfterSalesConfigModel {
	if cfg == nil {
		return nil
	}
	return &AfterSalesConfigModel{
		Model: gorm.Model{
			ID:        uint(cfg.ID),
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		},
		Key:         cfg.Key,
		Value:       cfg.Value,
		Description: cfg.Description,
	}
}

func toConfig(model *AfterSalesConfigModel) *domain.AfterSalesConfig {
	if model == nil {
		return nil
	}
	return &domain.AfterSalesConfig{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Key:         model.Key,
		Value:       model.Value,
		Description: model.Description,
	}
}
