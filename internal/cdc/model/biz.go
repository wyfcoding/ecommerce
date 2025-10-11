package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChangeEvent represents a captured database change event in the business logic layer.
type ChangeEvent struct {
	ID              uint
	EventID         string
	TableName       string
	OperationType   string
	PrimaryKeyValue string
	OldData         string // JSON string
	NewData         string // JSON string
	EventTimestamp  time.Time
}

// CdcRepo defines the interface for CDC data access.
type CdcRepo interface {
	CaptureChangeEvent(ctx context.Context, event *ChangeEvent) (*ChangeEvent, error)
}

// CdcUsecase is the business logic for CDC.
type CdcUsecase struct {
	repo CdcRepo
}

// NewCdcUsecase creates a new CdcUsecase.
func NewCdcUsecase(repo CdcRepo) *CdcUsecase {
	return &CdcUsecase{repo: repo}
}

// CaptureChangeEvent captures a database change event.
func (uc *CdcUsecase) CaptureChangeEvent(ctx context.Context, tableName, operationType, primaryKeyValue, oldData, newData string) (*ChangeEvent, error) {
	eventID := uuid.New().String()
	event := &ChangeEvent{
		EventID:         eventID,
		TableName:       tableName,
		OperationType:   operationType,
		PrimaryKeyValue: primaryKeyValue,
		OldData:         oldData,
		NewData:         newData,
		EventTimestamp:  time.Now(),
	}
	return uc.repo.CaptureChangeEvent(ctx, event)
}
