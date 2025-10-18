package repository

import (
	"context"

	"ecommerce/internal/cdc/model"
)

// CdcRepo defines the interface for CDC data access.
type CdcRepo interface {
	CaptureChangeEvent(ctx context.Context, event *model.ChangeEvent) (*model.ChangeEvent, error)
}