// 生成摘要：实现积分交易搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.PointsTransaction 的 JSON 映射一致。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"github.com/wyfcoding/pkg/search"
)

type pointsTransactionSearchRepository struct {
	client *search.Client
	index  string
}

// NewPointsTransactionSearchRepository 创建积分交易搜索仓储实现。
func NewPointsTransactionSearchRepository(client *search.Client, index string) domain.PointsTransactionSearchRepository {
	if index == "" {
		index = "points_transactions"
	}
	return &pointsTransactionSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *pointsTransactionSearchRepository) IndexTransaction(ctx context.Context, tx *domain.PointsTransaction) error {
	if tx == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(tx.ID), tx)
}

func (r *pointsTransactionSearchRepository) DeleteTransaction(ctx context.Context, transactionID uint64) error {
	if transactionID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(transactionID))
}

func (r *pointsTransactionSearchRepository) SearchTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*domain.PointsTransaction, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"term": map[string]any{"user_id": userID},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.PointsTransaction `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.PointsTransaction, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *pointsTransactionSearchRepository) documentID(transactionID uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", transactionID))
}
