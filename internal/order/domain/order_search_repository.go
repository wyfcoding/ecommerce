// 生成摘要：定义订单搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引可按 user_id、status、order_no 过滤并按 CreatedAt 排序。
package domain

import "context"

// OrderSearchRepository 定义订单搜索仓储接口。
type OrderSearchRepository interface {
	// Index 将订单写入搜索索引。
	Index(ctx context.Context, order *Order) error
	// Delete 从索引中删除订单。
	Delete(ctx context.Context, orderID uint64) error
	// Search 按条件检索订单（支持用户与状态过滤、分页）。
	Search(ctx context.Context, userID *uint64, status *int, offset, limit int) ([]*Order, int64, error)
	// FindByOrderNo 通过订单号检索订单。
	FindByOrderNo(ctx context.Context, orderNo string) (*Order, error)
}
