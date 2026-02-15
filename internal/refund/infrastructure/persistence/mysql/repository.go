// Package mysql 退款服务 MySQL 仓储实现
package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/refund/domain"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

type RefundRequestModel struct {
	ID            uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	RefundNo      string     `gorm:"column:refund_no;type:varchar(64);uniqueIndex;not null"`
	OrderID       string     `gorm:"column:order_id;type:varchar(64);index;not null"`
	OrderNo       string     `gorm:"column:order_no;type:varchar(64);not null"`
	UserID        uint64     `gorm:"column:user_id;index;not null"`
	MerchantID    uint64     `gorm:"column:merchant_id;index;not null"`
	Amount        int64      `gorm:"column:amount;not null"` // 分
	Reason        string     `gorm:"column:reason;type:varchar(255)"`
	Description   string     `gorm:"column:description;type:text"`
	Type          int8       `gorm:"column:type;type:tinyint;not null"`
	Status        int8       `gorm:"column:status;type:tinyint;not null;index"`
	PaymentID     string     `gorm:"column:payment_id;type:varchar(64)"`
	TransactionID string     `gorm:"column:transaction_id;type:varchar(64)"`
	RejectReason  string     `gorm:"column:reject_reason;type:varchar(255)"`
	OperatorID    uint64     `gorm:"column:operator_id"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	CompletedAt   *time.Time `gorm:"column:completed_at"`

	Items []RefundItemModel `gorm:"foreignKey:RefundID"`
}

func (RefundRequestModel) TableName() string { return "refund_requests" }

type RefundItemModel struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	RefundID     uint64 `gorm:"column:refund_id;index;not null"`
	OrderItemID  string `gorm:"column:order_item_id;type:varchar(64);not null"`
	SkuID        uint64 `gorm:"column:sku_id;not null"`
	Quantity     int32  `gorm:"column:quantity;not null"`
	RefundAmount int64  `gorm:"column:refund_amount;not null"`
}

func (RefundItemModel) TableName() string { return "refund_items" }

type RefundRepositoryImpl struct {
	db *database.DB
}

func NewRefundRepository(db *database.DB) domain.RefundRepository {
	return &RefundRepositoryImpl{db: db}
}

func (r *RefundRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := contextx.GetTx(ctx).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return r.db.DB.WithContext(ctx)
}

func (r *RefundRepositoryImpl) Save(ctx context.Context, refund *domain.RefundRequest) error {
	model := toModel(refund)
	if err := r.getDB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(model).Error; err != nil {
		return err
	}
	// 更新回主键
	refund.ID = model.ID
	return nil
}

func (r *RefundRepositoryImpl) GetByID(ctx context.Context, id uint64) (*domain.RefundRequest, error) {
	var model RefundRequestModel
	if err := r.getDB(ctx).Preload("Items").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *RefundRepositoryImpl) GetByRefundNo(ctx context.Context, no string) (*domain.RefundRequest, error) {
	var model RefundRequestModel
	if err := r.getDB(ctx).Preload("Items").Where("refund_no = ?", no).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *RefundRepositoryImpl) ListByOrder(ctx context.Context, orderID string) ([]*domain.RefundRequest, error) {
	var models []RefundRequestModel
	if err := r.getDB(ctx).Preload("Items").Where("order_id = ?", orderID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainList(models), nil
}

func (r *RefundRepositoryImpl) ListByMerchant(ctx context.Context, merchantID uint64, status domain.RefundStatus, page, size int) ([]*domain.RefundRequest, int64, error) {
	var models []RefundRequestModel
	var total int64

	db := r.getDB(ctx).Model(&RefundRequestModel{}).Where("merchant_id = ?", merchantID)
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Preload("Items").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return toDomainList(models), total, nil
}

func (r *RefundRepositoryImpl) ListPending(ctx context.Context, limit int) ([]*domain.RefundRequest, error) {
	var models []RefundRequestModel
	// 查找 waiting approval 的记录
	if err := r.getDB(ctx).
		Where("status IN ?", []int{int(domain.RefundStatusApproved)}). // Approved means ready to be refunded by system
		Order("updated_at ASC").
		Limit(limit).
		Preload("Items").
		Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainList(models), nil
}

func toModel(d *domain.RefundRequest) *RefundRequestModel {
	items := make([]RefundItemModel, len(d.Items))
	for i, item := range d.Items {
		items[i] = RefundItemModel{
			OrderItemID:  item.OrderItemID,
			SkuID:        item.SkuID,
			Quantity:     item.Quantity,
			RefundAmount: item.RefundAmount,
		}
		if item.ID > 0 {
			items[i].ID = item.ID
		}
		// RefundID will be handled by GORM association
	}

	return &RefundRequestModel{
		ID:            d.ID,
		RefundNo:      d.RefundNo,
		OrderID:       d.OrderID,
		OrderNo:       d.OrderNo,
		UserID:        d.UserID,
		MerchantID:    d.MerchantID,
		Amount:        d.Amount,
		Reason:        d.Reason,
		Description:   d.Description,
		Type:          int8(d.Type),
		Status:        int8(d.Status),
		PaymentID:     d.PaymentID,
		TransactionID: d.TransactionID,
		RejectReason:  d.RejectReason,
		OperatorID:    d.OperatorID,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		CompletedAt:   d.CompletedAt,
		Items:         items,
	}
}

func toDomain(m *RefundRequestModel) *domain.RefundRequest {
	items := make([]domain.RefundItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = domain.RefundItem{
			ID:           item.ID,
			RefundID:     item.RefundID,
			OrderItemID:  item.OrderItemID,
			SkuID:        item.SkuID,
			Quantity:     item.Quantity,
			RefundAmount: item.RefundAmount,
		}
	}

	return &domain.RefundRequest{
		ID:            m.ID,
		RefundNo:      m.RefundNo,
		OrderID:       m.OrderID,
		OrderNo:       m.OrderNo,
		UserID:        m.UserID,
		MerchantID:    m.MerchantID,
		Amount:        m.Amount,
		Reason:        m.Reason,
		Description:   m.Description,
		Type:          domain.RefundType(m.Type),
		Status:        domain.RefundStatus(m.Status),
		PaymentID:     m.PaymentID,
		TransactionID: m.TransactionID,
		RejectReason:  m.RejectReason,
		OperatorID:    m.OperatorID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		CompletedAt:   m.CompletedAt,
		Items:         items,
	}
}

func toDomainList(models []RefundRequestModel) []*domain.RefundRequest {
	result := make([]*domain.RefundRequest, len(models))
	for i, m := range models {
		result[i] = toDomain(&m)
	}
	return result
}
