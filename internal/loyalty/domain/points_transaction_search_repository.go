// 生成摘要：定义积分交易搜索仓储接口（Elasticsearch）。
package domain

import "context"

// PointsTransactionSearchRepository 定义积分交易搜索的访问接口。
type PointsTransactionSearchRepository interface {
	IndexTransaction(ctx context.Context, tx *PointsTransaction) error
	DeleteTransaction(ctx context.Context, transactionID uint64) error
	SearchTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*PointsTransaction, int64, error)
}
