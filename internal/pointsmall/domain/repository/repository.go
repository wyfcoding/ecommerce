package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity"
)

// PointsRepository 积分商城仓储接口
type PointsRepository interface {
	// 商品
	SaveProduct(ctx context.Context, product *entity.PointsProduct) error
	GetProduct(ctx context.Context, id uint64) (*entity.PointsProduct, error)
	ListProducts(ctx context.Context, status *entity.PointsProductStatus, offset, limit int) ([]*entity.PointsProduct, int64, error)

	// 订单
	SaveOrder(ctx context.Context, order *entity.PointsOrder) error
	GetOrder(ctx context.Context, id uint64) (*entity.PointsOrder, error)
	ListOrders(ctx context.Context, userID uint64, status *entity.PointsOrderStatus, offset, limit int) ([]*entity.PointsOrder, int64, error)

	// 账户 & 流水
	GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error)
	SaveAccount(ctx context.Context, account *entity.PointsAccount) error
	SaveTransaction(ctx context.Context, tx *entity.PointsTransaction) error
	ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error)
}
