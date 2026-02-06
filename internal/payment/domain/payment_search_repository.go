// 生成摘要：定义支付搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引可按 user_id、status、payment_no 过滤并按 CreatedAt 排序。
package domain

import (
	"context"
	"time"
)

// PaymentSearchRepository 定义支付搜索仓储接口。
type PaymentSearchRepository interface {
	// Index 将支付写入搜索索引。
	Index(ctx context.Context, payment *Payment) error
	// Delete 从索引中删除支付。
	Delete(ctx context.Context, paymentID uint64) error
	// Search 按条件检索支付（支持用户与状态过滤、分页）。
	Search(ctx context.Context, userID *uint64, status *PaymentStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Payment, int64, error)
	// FindByPaymentNo 通过支付单号检索支付。
	FindByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error)
	// FindByOrderID 通过订单ID检索支付。
	FindByOrderID(ctx context.Context, orderID uint64) (*Payment, error)
}
