package domain

import (
	"context"
)

type SupplierRepository interface {
	Save(ctx context.Context, supplier *Supplier) error
	Get(ctx context.Context, id string) (*Supplier, error)
}
