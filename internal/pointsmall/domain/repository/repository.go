package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity" // 导入积分商城领域的实体定义。
)

// PointsRepository 是积分商城模块的仓储接口。
// 它定义了对积分商品、积分订单、积分账户和积分交易实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type PointsRepository interface {
	// --- 商品管理 (Product methods) ---

	// SaveProduct 将积分商品实体保存到数据存储中。
	// 如果商品已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// product: 待保存的积分商品实体。
	SaveProduct(ctx context.Context, product *entity.PointsProduct) error
	// GetProduct 根据ID获取积分商品实体。
	GetProduct(ctx context.Context, id uint64) (*entity.PointsProduct, error)
	// ListProducts 列出所有积分商品实体，支持通过状态过滤和分页。
	ListProducts(ctx context.Context, status *entity.PointsProductStatus, offset, limit int) ([]*entity.PointsProduct, int64, error)

	// --- 订单管理 (Order methods) ---

	// SaveOrder 将积分订单实体保存到数据存储中。
	SaveOrder(ctx context.Context, order *entity.PointsOrder) error
	// GetOrder 根据ID获取积分订单实体。
	GetOrder(ctx context.Context, id uint64) (*entity.PointsOrder, error)
	// ListOrders 列出指定用户ID的所有积分订单实体，支持通过状态过滤和分页。
	ListOrders(ctx context.Context, userID uint64, status *entity.PointsOrderStatus, offset, limit int) ([]*entity.PointsOrder, int64, error)

	// --- 账户与流水管理 (Account & Transaction methods) ---

	// GetAccount 根据用户ID获取积分账户实体。
	GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error)
	// SaveAccount 将积分账户实体保存到数据存储中。
	SaveAccount(ctx context.Context, account *entity.PointsAccount) error
	// SaveTransaction 将积分交易流水实体保存到数据存储中。
	SaveTransaction(ctx context.Context, tx *entity.PointsTransaction) error
	// ListTransactions 列出指定用户ID的所有积分交易流水实体，支持分页。
	ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error)
}
