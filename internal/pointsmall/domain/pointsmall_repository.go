package domain

import (
	"context"
)

// PointsRepository 是积分商城模块的写模型仓储接口。
type PointsRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// Product
	SaveProduct(ctx context.Context, product *PointsProduct) error
	SaveProductInTx(ctx context.Context, tx any, product *PointsProduct) error
	GetProduct(ctx context.Context, id uint64) (*PointsProduct, error)
	ListProducts(ctx context.Context, query *PointsProductQuery) ([]*PointsProduct, int64, error)

	// Order
	SaveOrder(ctx context.Context, order *PointsOrder) error
	SaveOrderInTx(ctx context.Context, tx any, order *PointsOrder) error
	GetOrder(ctx context.Context, id uint64) (*PointsOrder, error)
	ListOrders(ctx context.Context, query *PointsOrderQuery) ([]*PointsOrder, int64, error)

	// Account & Transaction
	GetAccount(ctx context.Context, userID uint64) (*PointsAccount, error)
	SaveAccount(ctx context.Context, account *PointsAccount) error
	SaveAccountInTx(ctx context.Context, tx any, account *PointsAccount) error
	SaveTransaction(ctx context.Context, tx *PointsTransaction) error
	SaveTransactionInTx(ctx context.Context, tx any, transaction *PointsTransaction) error
	ListTransactions(ctx context.Context, query *PointsTransactionQuery) ([]*PointsTransaction, int64, error)
}

// PointsProductQuery 积分商品查询条件。
type PointsProductQuery struct {
	Status   *PointsProductStatus
	Keyword  string
	Page     int
	PageSize int
}

// PointsOrderQuery 积分订单查询条件。
type PointsOrderQuery struct {
	UserID   uint64
	Status   *PointsOrderStatus
	OrderNo  string
	Page     int
	PageSize int
}

// PointsTransactionQuery 积分流水查询条件。
type PointsTransactionQuery struct {
	UserID   uint64
	Type     string
	RefID    string
	Page     int
	PageSize int
}
