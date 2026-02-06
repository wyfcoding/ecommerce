package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"gorm.io/gorm"
)

// PaymentModel 支付写模型（持久化专用）。
type PaymentModel struct {
	gorm.Model
	PaymentNo      string               `gorm:"type:varchar(64);uniqueIndex;not null;comment:支付单号"`
	OrderID        uint64               `gorm:"index;comment:订单ID"`
	OrderNo        string               `gorm:"type:varchar(64);comment:订单号"`
	UserID         uint64               `gorm:"index;comment:用户ID"`
	Amount         int64                `gorm:"not null;comment:总金额"`
	CapturedAmount int64                `gorm:"default:0;comment:捕获金额"`
	Currency       string               `gorm:"type:varchar(10);default:'CNY';comment:币种"`
	PaymentMethod  string               `gorm:"type:varchar(32);comment:支付方式"`
	GatewayType    domain.GatewayType   `gorm:"type:varchar(32);comment:网关类型"`
	Status         domain.PaymentStatus `gorm:"type:tinyint;comment:支付状态"`
	TransactionID  string               `gorm:"type:varchar(128);comment:网关交易号"`
	ThirdPartyNo   string               `gorm:"type:varchar(128);comment:三方单号"`
	CallbackData   string               `gorm:"type:text;comment:回调原始数据"`
	FailureReason  string               `gorm:"type:varchar(255);comment:失败原因"`
	PaidAt         *time.Time           `gorm:"comment:支付时间"`
	CancelledAt    *time.Time           `gorm:"comment:取消时间"`
	RefundedAt     *time.Time           `gorm:"comment:退款时间"`
	Version        int64                `gorm:"column:version;default:1;comment:乐观锁版本"`

	Logs   []PaymentLogModel   `gorm:"foreignKey:PaymentID"`
	Splits []PaymentSplitModel `gorm:"foreignKey:PaymentID"`
}

func (PaymentModel) TableName() string {
	return "payments"
}

// PaymentSplitModel 资金拆分（持久化专用）。
type PaymentSplitModel struct {
	gorm.Model
	PaymentID     uint64 `gorm:"index;not null;comment:支付ID"`
	RecipientID   uint64 `gorm:"index;comment:接收者ID(商家或平台)"`
	RecipientType string `gorm:"type:varchar(32);comment:MERCHANT, PLATFORM, TAX"`
	Amount        int64  `gorm:"not null"`
	Status        string `gorm:"type:varchar(32);default:'PENDING';comment:PENDING, SETTLED"`
}

func (PaymentSplitModel) TableName() string {
	return "payment_splits"
}

// PaymentLogModel 支付日志（持久化专用）。
type PaymentLogModel struct {
	gorm.Model
	PaymentID uint64 `gorm:"index;not null;comment:支付ID"`
	UserID    uint64 `gorm:"index;not null;comment:用户ID"`
	Action    string `gorm:"type:varchar(64);comment:动作"`
	OldStatus string `gorm:"type:varchar(32);comment:旧状态"`
	NewStatus string `gorm:"type:varchar(32);comment:新状态"`
	Remark    string `gorm:"type:varchar(255);comment:备注"`
}

func (PaymentLogModel) TableName() string {
	return "payment_logs"
}

// RefundModel 退款单（持久化专用）。
type RefundModel struct {
	gorm.Model
	RefundNo        string               `gorm:"type:varchar(64);uniqueIndex;not null"`
	PaymentID       uint64               `gorm:"index;not null"`
	PaymentNo       string               `gorm:"type:varchar(64)"`
	OrderID         uint64               `gorm:"index"`
	OrderNo         string               `gorm:"type:varchar(64)"`
	UserID          uint64               `gorm:"index"`
	RefundAmount    int64                `gorm:"not null"`
	Reason          string               `gorm:"type:varchar(255)"`
	Status          domain.PaymentStatus `gorm:"type:tinyint"`
	ThirdPartyNo    string               `gorm:"type:varchar(128)"`
	GatewayRefundID string               `gorm:"type:varchar(128)"`
	FailureReason   string               `gorm:"type:varchar(255)"`
	RefundedAt      *time.Time
}

func (RefundModel) TableName() string {
	return "refunds"
}

// ChannelConfigModel 渠道配置（持久化专用）。
type ChannelConfigModel struct {
	gorm.Model
	Code        string             `gorm:"type:varchar(64);uniqueIndex;not null"`
	Type        domain.ChannelType `gorm:"type:varchar(32);not null"`
	Name        string             `gorm:"type:varchar(64)"`
	Priority    int                `gorm:"default:0"`
	Enabled     bool               `gorm:"default:true"`
	ConfigJSON  string             `gorm:"type:text"`
	RatePercent float64            `gorm:"default:0"`
	Description string             `gorm:"type:varchar(255)"`
}

func (ChannelConfigModel) TableName() string {
	return "channel_configs"
}

// ReconciliationRecordModel 对账记录（持久化专用）。
type ReconciliationRecordModel struct {
	gorm.Model
	OrderNo       string `gorm:"type:varchar(64)"`
	PaymentID     uint64 `gorm:"index"`
	GatewayAmount int64
	SystemAmount  int64
	DiffAmount    int64
	Status        string `gorm:"type:varchar(32)"`
	Remark        string `gorm:"type:varchar(255)"`
}

func (ReconciliationRecordModel) TableName() string {
	return "reconciliation_records"
}

func toPaymentModel(payment *domain.Payment) *PaymentModel {
	if payment == nil {
		return nil
	}

	model := &PaymentModel{
		Model: gorm.Model{
			ID:        payment.ID,
			CreatedAt: payment.CreatedAt,
			UpdatedAt: payment.UpdatedAt,
		},
		PaymentNo:      payment.PaymentNo,
		OrderID:        payment.OrderID,
		OrderNo:        payment.OrderNo,
		UserID:         payment.UserID,
		Amount:         payment.Amount,
		CapturedAmount: payment.CapturedAmount,
		Currency:       payment.Currency,
		PaymentMethod:  payment.PaymentMethod,
		GatewayType:    payment.GatewayType,
		Status:         payment.Status,
		TransactionID:  payment.TransactionID,
		ThirdPartyNo:   payment.ThirdPartyNo,
		CallbackData:   payment.CallbackData,
		FailureReason:  payment.FailureReason,
		PaidAt:         payment.PaidAt,
		CancelledAt:    payment.CancelledAt,
		RefundedAt:     payment.RefundedAt,
		Version:        payment.PersistenceVer,
	}

	model.Logs = make([]PaymentLogModel, 0, len(payment.Logs))
	for _, log := range payment.Logs {
		if log == nil {
			continue
		}
		model.Logs = append(model.Logs, PaymentLogModel{
			Model: gorm.Model{
				ID:        log.ID,
				CreatedAt: log.CreatedAt,
				UpdatedAt: log.UpdatedAt,
			},
			PaymentID: log.PaymentID,
			UserID:    log.UserID,
			Action:    log.Action,
			OldStatus: log.OldStatus,
			NewStatus: log.NewStatus,
			Remark:    log.Remark,
		})
	}

	model.Splits = make([]PaymentSplitModel, 0, len(payment.Splits))
	for _, split := range payment.Splits {
		model.Splits = append(model.Splits, PaymentSplitModel{
			Model: gorm.Model{
				ID:        split.ID,
				CreatedAt: split.CreatedAt,
				UpdatedAt: split.UpdatedAt,
			},
			PaymentID:     split.PaymentID,
			RecipientID:   split.RecipientID,
			RecipientType: split.RecipientType,
			Amount:        split.Amount,
			Status:        split.Status,
		})
	}

	return model
}

func toDomainPayment(model *PaymentModel) *domain.Payment {
	if model == nil {
		return nil
	}

	payment := &domain.Payment{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		PaymentNo:      model.PaymentNo,
		OrderID:        model.OrderID,
		OrderNo:        model.OrderNo,
		UserID:         model.UserID,
		Amount:         model.Amount,
		CapturedAmount: model.CapturedAmount,
		Currency:       model.Currency,
		PaymentMethod:  model.PaymentMethod,
		GatewayType:    model.GatewayType,
		Status:         model.Status,
		TransactionID:  model.TransactionID,
		ThirdPartyNo:   model.ThirdPartyNo,
		CallbackData:   model.CallbackData,
		FailureReason:  model.FailureReason,
		PaidAt:         model.PaidAt,
		CancelledAt:    model.CancelledAt,
		RefundedAt:     model.RefundedAt,
		PersistenceVer: model.Version,
	}

	payment.Logs = make([]*domain.PaymentLog, 0, len(model.Logs))
	for _, log := range model.Logs {
		logCopy := log
		payment.Logs = append(payment.Logs, &domain.PaymentLog{
			ID:        logCopy.ID,
			CreatedAt: logCopy.CreatedAt,
			UpdatedAt: logCopy.UpdatedAt,
			PaymentID: logCopy.PaymentID,
			UserID:    logCopy.UserID,
			Action:    logCopy.Action,
			OldStatus: logCopy.OldStatus,
			NewStatus: logCopy.NewStatus,
			Remark:    logCopy.Remark,
		})
	}

	payment.Splits = make([]domain.PaymentSplit, 0, len(model.Splits))
	for _, split := range model.Splits {
		splitCopy := split
		payment.Splits = append(payment.Splits, domain.PaymentSplit{
			ID:            splitCopy.ID,
			CreatedAt:     splitCopy.CreatedAt,
			UpdatedAt:     splitCopy.UpdatedAt,
			PaymentID:     splitCopy.PaymentID,
			RecipientID:   splitCopy.RecipientID,
			RecipientType: splitCopy.RecipientType,
			Amount:        splitCopy.Amount,
			Status:        splitCopy.Status,
		})
	}

	payment.SetID(payment.PaymentNo)
	payment.SetVersion(payment.PersistenceVer)
	payment.InitFSM()

	return payment
}

func toRefundModel(refund *domain.Refund) *RefundModel {
	if refund == nil {
		return nil
	}
	return &RefundModel{
		Model: gorm.Model{
			ID:        refund.ID,
			CreatedAt: refund.CreatedAt,
			UpdatedAt: refund.UpdatedAt,
		},
		RefundNo:        refund.RefundNo,
		PaymentID:       refund.PaymentID,
		PaymentNo:       refund.PaymentNo,
		OrderID:         refund.OrderID,
		OrderNo:         refund.OrderNo,
		UserID:          refund.UserID,
		RefundAmount:    refund.RefundAmount,
		Reason:          refund.Reason,
		Status:          refund.Status,
		ThirdPartyNo:    refund.ThirdPartyNo,
		GatewayRefundID: refund.GatewayRefundID,
		FailureReason:   refund.FailureReason,
		RefundedAt:      refund.RefundedAt,
	}
}

func toDomainRefund(model *RefundModel) *domain.Refund {
	if model == nil {
		return nil
	}
	return &domain.Refund{
		ID:              model.ID,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		RefundNo:        model.RefundNo,
		PaymentID:       model.PaymentID,
		PaymentNo:       model.PaymentNo,
		OrderID:         model.OrderID,
		OrderNo:         model.OrderNo,
		UserID:          model.UserID,
		RefundAmount:    model.RefundAmount,
		Reason:          model.Reason,
		Status:          model.Status,
		ThirdPartyNo:    model.ThirdPartyNo,
		GatewayRefundID: model.GatewayRefundID,
		FailureReason:   model.FailureReason,
		RefundedAt:      model.RefundedAt,
	}
}

func toChannelConfigModel(cfg *domain.ChannelConfig) *ChannelConfigModel {
	if cfg == nil {
		return nil
	}
	return &ChannelConfigModel{
		Model: gorm.Model{
			ID:        cfg.ID,
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		},
		Code:        cfg.Code,
		Type:        cfg.Type,
		Name:        cfg.Name,
		Priority:    cfg.Priority,
		Enabled:     cfg.Enabled,
		ConfigJSON:  cfg.ConfigJSON,
		RatePercent: cfg.RatePercent,
		Description: cfg.Description,
	}
}

func toDomainChannelConfig(model *ChannelConfigModel) *domain.ChannelConfig {
	if model == nil {
		return nil
	}
	return &domain.ChannelConfig{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Code:        model.Code,
		Type:        model.Type,
		Name:        model.Name,
		Priority:    model.Priority,
		Enabled:     model.Enabled,
		ConfigJSON:  model.ConfigJSON,
		RatePercent: model.RatePercent,
		Description: model.Description,
	}
}

func toReconciliationRecordModel(record *domain.ReconciliationRecord) *ReconciliationRecordModel {
	if record == nil {
		return nil
	}
	return &ReconciliationRecordModel{
		Model: gorm.Model{
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		},
		OrderNo:       record.OrderNo,
		PaymentID:     record.PaymentID,
		GatewayAmount: record.GatewayAmount,
		SystemAmount:  record.SystemAmount,
		DiffAmount:    record.DiffAmount,
		Status:        record.Status,
		Remark:        record.Remark,
	}
}
