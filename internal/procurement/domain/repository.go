package domain

import (
	"context"
)

type ProcurementRepository interface {
	SavePurchaseRequest(ctx context.Context, pr *PurchaseRequest) error
	GetPurchaseRequest(ctx context.Context, id string) (*PurchaseRequest, error)

	SavePurchaseOrder(ctx context.Context, po *PurchaseOrder) error
	GetPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error)
}
