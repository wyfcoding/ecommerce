package mysql

import (
	"context"
	"encoding/json"

	pb "github.com/wyfcoding/ecommerce/go-api/returns/v1"
	"github.com/wyfcoding/ecommerce/internal/returns/domain"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"gorm.io/gorm"
)

// ReturnRequestModel GORM 模型.
type ReturnRequestModel struct {
	gorm.Model
	ReturnID       string `gorm:"column:return_id;uniqueIndex;type:varchar(64);not null"`
	OrderID        string `gorm:"column:order_id;index;type:varchar(64);not null"`
	UserID         string `gorm:"column:user_id;index;type:varchar(64);not null"`
	ItemsJSON      string `gorm:"column:items_json;type:longtext"`
	Status         int32  `gorm:"column:status;index"`
	RMANumber      string `gorm:"column:rma_number;uniqueIndex;type:varchar(64)"`
	TrackingNumber string `gorm:"column:tracking_number;type:varchar(128)"`
	WarehouseNotes string `gorm:"column:warehouse_notes;type:text"`
}

func (ReturnRequestModel) TableName() string { return "returns_requests" }

type returnRepository struct {
	db     *database.DB
	logger *logging.Logger
}

func NewReturnRepository(db *database.DB, logger *logging.Logger) domain.ReturnRepository {
	return &returnRepository{db: db, logger: logger}
}

func (r *returnRepository) Save(ctx context.Context, req *domain.ReturnRequest) error {
	items, _ := json.Marshal(req.Items)
	m := &ReturnRequestModel{
		ReturnID:       req.ID,
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		ItemsJSON:      string(items),
		Status:         int32(req.Status),
		RMANumber:      req.RMANumber,
		TrackingNumber: req.TrackingNumber,
		WarehouseNotes: req.WarehouseNotes,
	}
	m.CreatedAt = req.CreatedAt
	m.UpdatedAt = req.UpdatedAt

	return r.db.RawDB().WithContext(ctx).Save(m).Error
}

func (r *returnRepository) GetByID(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	var m ReturnRequestModel
	if err := r.db.RawDB().WithContext(ctx).Where("return_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toReturnDomain(&m), nil
}

func (r *returnRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.ReturnRequest, int, error) {
	var models []*ReturnRequestModel
	var total int64
	db := r.db.RawDB().WithContext(ctx).Model(&ReturnRequestModel{}).Where("user_id = ?", userID)
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*domain.ReturnRequest, len(models))
	for i, m := range models {
		res[i] = toReturnDomain(m)
	}
	return res, int(total), nil
}

func (r *returnRepository) GetByRMA(ctx context.Context, rma string) (*domain.ReturnRequest, error) {
	var m ReturnRequestModel
	if err := r.db.RawDB().WithContext(ctx).Where("rma_number = ?", rma).First(&m).Error; err != nil {
		return nil, err
	}
	return toReturnDomain(&m), nil
}

func toReturnDomain(m *ReturnRequestModel) *domain.ReturnRequest {
	var items []domain.ReturnItem
	_ = json.Unmarshal([]byte(m.ItemsJSON), &items)
	return &domain.ReturnRequest{
		ID:             m.ReturnID,
		OrderID:        m.OrderID,
		UserID:         m.UserID,
		Items:          items,
		Status:         pb.ReturnStatus(m.Status),
		RMANumber:      m.RMANumber,
		TrackingNumber: m.TrackingNumber,
		WarehouseNotes: m.WarehouseNotes,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
