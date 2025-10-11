package data

import (
	"context"
	"ecommerce/internal/settlement/biz"
	"ecommerce/internal/settlement/data/model"
	"time"

	"gorm.io/gorm"
)

type settlementRepo struct {
	data *Data
}

// NewSettlementRepo creates a new SettlementRepo.
func NewSettlementRepo(data *Data) biz.SettlementRepo {
	return &settlementRepo{data: data}
}

// CreateSettlementRecord creates a new settlement record.
func (r *settlementRepo) CreateSettlementRecord(ctx context.Context, record *biz.SettlementRecord) (*biz.SettlementRecord, error) {
	po := &model.SettlementRecord{
		RecordID:         record.RecordID,
		OrderID:          record.OrderID,
		MerchantID:       record.MerchantID,
		TotalAmount:      record.TotalAmount,
		PlatformFee:      record.PlatformFee,
		SettlementAmount: record.SettlementAmount,
		Status:           record.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	record.ID = po.ID
	return record, nil
}

// GetSettlementRecordByID retrieves a settlement record by its ID.
func (r *settlementRepo) GetSettlementRecordByID(ctx context.Context, recordID string) (*biz.SettlementRecord, error) {
	var po model.SettlementRecord
	if err := r.data.db.WithContext(ctx).Where("record_id = ?", recordID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Record not found
		}
		return nil, err
	}
	return &biz.SettlementRecord{
		ID:               po.ID,
		RecordID:         po.RecordID,
		OrderID:          po.OrderID,
		MerchantID:       po.MerchantID,
		TotalAmount:      po.TotalAmount,
		PlatformFee:      po.PlatformFee,
		SettlementAmount: po.SettlementAmount,
		Status:           po.Status,
		CreatedAt:        po.CreatedAt,
		SettledAt:        po.SettledAt,
	}, nil
}

// ListSettlementRecords lists settlement records based on filters.
func (r *settlementRepo) ListSettlementRecords(ctx context.Context, merchantID uint64, status string, pageSize, pageNum uint32) ([]*biz.SettlementRecord, uint64, error) {
	var records []*model.SettlementRecord
	var total int64
	query := r.data.db.WithContext(ctx).Model(&model.SettlementRecord{})

	if merchantID != 0 {
		query = query.Where("merchant_id = ?", merchantID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	bizRecords := make([]*biz.SettlementRecord, len(records))
	for i, rec := range records {
		bizRecords[i] = &biz.SettlementRecord{
			ID:               rec.ID,
			RecordID:         rec.RecordID,
			OrderID:          rec.OrderID,
			MerchantID:       rec.MerchantID,
			TotalAmount:      rec.TotalAmount,
			PlatformFee:      rec.PlatformFee,
			SettlementAmount: rec.SettlementAmount,
			Status:           rec.Status,
			CreatedAt:        rec.CreatedAt,
			SettledAt:        rec.SettledAt,
		}
	}
	return bizRecords, uint64(total), nil
}

// UpdateSettlementRecordStatus updates the status of a settlement record.
func (r *settlementRepo) UpdateSettlementRecordStatus(ctx context.Context, recordID string, newStatus string, settledAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     newStatus,
		"settled_at": settledAt,
	}
	return r.data.db.WithContext(ctx).Model(&model.SettlementRecord{}).Where("record_id = ?", recordID).Updates(updates).Error
}
