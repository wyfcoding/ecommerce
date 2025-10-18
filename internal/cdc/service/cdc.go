package service

import (
	"context"
	"time"

	"ecommerce/internal/cdc/model"
	"ecommerce/internal/cdc/repository"
	"github.com/google/uuid"
)

// CdcService is the business logic for CDC.
type CdcService struct {
	repo repository.CdcRepo
}

// NewCdcService creates a new CdcService.
func NewCdcService(repo repository.CdcRepo) *CdcService {
	return &CdcService{repo: repo}
}

// CaptureChangeEvent captures a database change event.
func (s *CdcService) CaptureChangeEvent(ctx context.Context, tableName, operationType, primaryKeyValue, oldData, newData string) (*model.ChangeEvent, error) {
	eventID := uuid.New().String()
	event := &model.ChangeEvent{
		EventID:         eventID,
		TableName:       tableName,
		OperationType:   operationType,
		PrimaryKeyValue: primaryKeyValue,
		OldData:         oldData,
		NewData:         newData,
		EventTimestamp:  time.Now(),
	}
	return s.repo.CaptureChangeEvent(ctx, event)
}
