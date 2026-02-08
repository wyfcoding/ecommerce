package domain

import (
	"context"
)

type SettlementRepository interface {
	Save(ctx context.Context, s *Settlement) error
	GetByID(ctx context.Context, id string) (*Settlement, error)
	ListByMerchant(ctx context.Context, merchantID string, status string) ([]*Settlement, error)
}
