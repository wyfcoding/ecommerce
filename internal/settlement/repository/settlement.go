package repository

import (
	"context"
	"time"

	"ecommerce/internal/settlement/model"
)

// SettlementRepo defines the interface for settlement data access.
type SettlementRepo interface {
	CreateSettlementRecord(ctx context.Context, record *model.SettlementRecord) (*model.SettlementRecord, error)
	GetSettlementRecordByID(ctx context.Context, recordID string) (*model.SettlementRecord, error)
	ListSettlementRecords(ctx context.Context, merchantID uint64, status string, pageSize, pageNum uint32) ([]*model.SettlementRecord, uint64, error)
	UpdateSettlementRecordStatus(ctx context.Context, recordID string, newStatus string, settledAt *time.Time) error
}