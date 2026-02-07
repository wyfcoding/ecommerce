package domain

import "context"

// PointsProductSearchRepository 定义积分商品搜索仓储接口（Elasticsearch）。
type PointsProductSearchRepository interface {
	Index(ctx context.Context, product *PointsProduct) error
	Delete(ctx context.Context, productID uint64) error
	Search(ctx context.Context, query *PointsProductQuery, offset, limit int) ([]*PointsProduct, int64, error)
}
