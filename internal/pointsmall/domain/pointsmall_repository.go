package domain

import (
	"context"
)

// PointsRepository 是积分商城模块的仓储接口。
type PointsRepository interface {
	// Product
	SaveProduct(ctx context.Context, product *PointsProduct) error
	GetProduct(ctx context.Context, id uint64) (*PointsProduct, error)
	ListProducts(ctx context.Context, status *PointsProductStatus, offset, limit int) ([]*PointsProduct, int64, error)

	// Order
	SaveOrder(ctx context.Context, order *PointsOrder) error
	GetOrder(ctx context.Context, id uint64) (*PointsOrder, error)
	ListOrders(ctx context.Context, userID uint64, status *PointsOrderStatus, offset, limit int) ([]*PointsOrder, int64, error)

	// Account & Transaction
	GetAccount(ctx context.Context, userID uint64) (*PointsAccount, error)
	SaveAccount(ctx context.Context, account *PointsAccount) error
	SaveTransaction(ctx context.Context, tx *PointsTransaction) error
	ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*PointsTransaction, int64, error)
}
