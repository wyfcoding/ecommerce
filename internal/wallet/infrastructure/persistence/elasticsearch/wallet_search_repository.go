// Package elasticsearch 钱包搜索仓储实现（Elasticsearch）
// 生成摘要：
// 1) 实现 WalletSearchRepository 接口，使用 Elasticsearch 存储交易记录全文索引
// 2) 支持多维度搜索、聚合统计、时间范围查询
// 3) 通过事件投影保持索引与 MySQL 写模型的最终一致性
// 关键改动: 将 search.ElasticsearchClient 改为 search.Client 以匹配 pkg/search 包的实际导出类型
package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/search"
)

// esSearchResponse ES 搜索响应结构（用于反序列化 search.Client.Search 的结果）
type esSearchResponse struct {
	Hits struct {
		TotalHits struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []esHit `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]json.RawMessage `json:"aggregations"`
}

// esHit 单条搜索命中
type esHit struct {
	Source json.RawMessage `json:"_source"`
}

// WalletSearchRepositoryImpl Elasticsearch 钱包搜索仓储实现
type WalletSearchRepositoryImpl struct {
	esClient *search.Client
	logger   *slog.Logger
}

// NewWalletSearchRepository 创建 Elasticsearch 钱包搜索仓储实例
func NewWalletSearchRepository(esClient *search.Client, logger *slog.Logger) domain.WalletSearchRepository {
	return &WalletSearchRepositoryImpl{
		esClient: esClient,
		logger:   logger.With("module", "wallet_search_repository"),
	}
}

// IndexTransaction 索引交易记录到 ES
func (r *WalletSearchRepositoryImpl) IndexTransaction(ctx context.Context, tx *domain.TransactionReadModel) error {
	start := time.Now()

	indexName := "wallet_transactions"
	docID := fmt.Sprintf("%s_%d", tx.TransactionNo, tx.ID)

	err := r.esClient.Index(ctx, indexName, docID, tx)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to index transaction",
			"tx_no", tx.TransactionNo, "error", err, "duration", time.Since(start))
		return fmt.Errorf("index transaction: %w", err)
	}

	r.logger.DebugContext(ctx, "transaction indexed to es",
		"tx_no", tx.TransactionNo, "duration", time.Since(start))
	return nil
}

// SearchTransactions 多维度搜索交易记录
func (r *WalletSearchRepositoryImpl) SearchTransactions(ctx context.Context, query *domain.TransactionSearchQuery) (*domain.TransactionSearchResult, error) {
	start := time.Now()

	indexName := "wallet_transactions"
	searchQuery := r.buildSearchQuery(query)

	result := &domain.TransactionSearchResult{
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	// 执行搜索
	var searchResult esSearchResponse
	if err := r.esClient.Search(ctx, indexName, searchQuery, &searchResult); err != nil {
		r.logger.ErrorContext(ctx, "failed to search transactions",
			"error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("search transactions: %w", err)
	}

	// 解析结果
	for _, hit := range searchResult.Hits.Hits {
		var tx domain.TransactionReadModel
		if err := json.Unmarshal(hit.Source, &tx); err != nil {
			r.logger.WarnContext(ctx, "failed to unmarshal transaction hit", "error", err)
			continue
		}
		result.Items = append(result.Items, &tx)
	}

	result.Total = searchResult.Hits.TotalHits.Value

	r.logger.DebugContext(ctx, "transactions search completed",
		"total", result.Total, "duration", time.Since(start))
	return result, nil
}

// AggregateByType 按交易类型聚合统计
func (r *WalletSearchRepositoryImpl) AggregateByType(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error) {
	indexName := "wallet_transactions"

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"wallet_id": walletID,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"created_at": map[string]interface{}{
								"gte": startTime.Format(time.RFC3339),
								"lte": endTime.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"by_type": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "type.keyword",
				},
				"aggs": map[string]interface{}{
					"total_amount": map[string]interface{}{
						"sum": map[string]interface{}{
							"field": "amount",
						},
					},
				},
			},
		},
		"size": 0,
	}

	var searchResult esSearchResponse
	if err := r.esClient.Search(ctx, indexName, query, &searchResult); err != nil {
		return nil, fmt.Errorf("aggregate by type: %w", err)
	}

	return r.parseTermsAggregation(searchResult.Aggregations, "by_type", "total_amount"), nil
}

// AggregateDaily 按日聚合交易金额
func (r *WalletSearchRepositoryImpl) AggregateDaily(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error) {
	indexName := "wallet_transactions"

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"wallet_id": walletID,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"created_at": map[string]interface{}{
								"gte": startTime.Format(time.RFC3339),
								"lte": endTime.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"by_date": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":             "created_at",
					"calendar_interval": "day",
					"format":            "yyyy-MM-dd",
				},
				"aggs": map[string]interface{}{
					"daily_amount": map[string]interface{}{
						"sum": map[string]interface{}{
							"field": "amount",
						},
					},
				},
			},
		},
		"size": 0,
	}

	var searchResult esSearchResponse
	if err := r.esClient.Search(ctx, indexName, query, &searchResult); err != nil {
		return nil, fmt.Errorf("aggregate daily: %w", err)
	}

	return r.parseDateHistogramAggregation(searchResult.Aggregations, "by_date", "daily_amount"), nil
}

// parseTermsAggregation 解析 terms 聚合结果
func (r *WalletSearchRepositoryImpl) parseTermsAggregation(aggs map[string]json.RawMessage, aggName, metricName string) map[string]int64 {
	result := make(map[string]int64)
	raw, ok := aggs[aggName]
	if !ok {
		return result
	}

	var aggResult struct {
		Buckets []struct {
			Key    string `json:"key"`
			Metric struct {
				Value float64 `json:"value"`
			} `json:"total_amount"`
		} `json:"buckets"`
	}
	if err := json.Unmarshal(raw, &aggResult); err != nil {
		r.logger.Error("failed to unmarshal terms aggregation", "error", err)
		return result
	}

	for _, bucket := range aggResult.Buckets {
		result[bucket.Key] = int64(bucket.Metric.Value)
	}
	return result
}

// parseDateHistogramAggregation 解析日期直方图聚合结果
func (r *WalletSearchRepositoryImpl) parseDateHistogramAggregation(aggs map[string]json.RawMessage, aggName, metricName string) map[string]int64 {
	result := make(map[string]int64)
	raw, ok := aggs[aggName]
	if !ok {
		return result
	}

	var aggResult struct {
		Buckets []struct {
			KeyAsString string `json:"key_as_string"`
			Metric      struct {
				Value float64 `json:"value"`
			} `json:"daily_amount"`
		} `json:"buckets"`
	}
	if err := json.Unmarshal(raw, &aggResult); err != nil {
		r.logger.Error("failed to unmarshal date histogram aggregation", "error", err)
		return result
	}

	for _, bucket := range aggResult.Buckets {
		result[bucket.KeyAsString] = int64(bucket.Metric.Value)
	}
	return result
}

// buildSearchQuery 构建 Elasticsearch 搜索查询
func (r *WalletSearchRepositoryImpl) buildSearchQuery(query *domain.TransactionSearchQuery) map[string]interface{} {
	mustClauses := []interface{}{}

	// 基础条件
	if query.WalletID > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"wallet_id": query.WalletID,
			},
		})
	}

	if query.UserID > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"user_id": query.UserID,
			},
		})
	}

	if query.TransactionNo != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"transaction_no.keyword": query.TransactionNo,
			},
		})
	}

	if query.Type != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"type.keyword": query.Type,
			},
		})
	}

	if query.Status != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"status.keyword": query.Status,
			},
		})
	}

	// 金额范围
	if query.MinAmount > 0 || query.MaxAmount > 0 {
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"amount": map[string]interface{}{},
			},
		}
		if query.MinAmount > 0 {
			rangeQuery["range"].(map[string]interface{})["amount"].(map[string]interface{})["gte"] = query.MinAmount
		}
		if query.MaxAmount > 0 {
			rangeQuery["range"].(map[string]interface{})["amount"].(map[string]interface{})["lte"] = query.MaxAmount
		}
		mustClauses = append(mustClauses, rangeQuery)
	}

	// 时间范围
	if query.StartTime != nil || query.EndTime != nil {
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]interface{}{},
			},
		}
		if query.StartTime != nil {
			rangeQuery["range"].(map[string]interface{})["created_at"].(map[string]interface{})["gte"] = query.StartTime.Format(time.RFC3339)
		}
		if query.EndTime != nil {
			rangeQuery["range"].(map[string]interface{})["created_at"].(map[string]interface{})["lte"] = query.EndTime.Format(time.RFC3339)
		}
		mustClauses = append(mustClauses, rangeQuery)
	}

	// 关键词搜索（备注、交易号等）
	if query.Keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Keyword,
				"fields": []string{"remark", "transaction_no"},
			},
		})
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"from": (query.Page - 1) * query.PageSize,
		"size": query.PageSize,
	}

	// 排序
	if query.SortBy != "" {
		sortOrder := query.SortOrder
		if sortOrder == "" {
			sortOrder = "desc"
		}
		searchQuery["sort"] = []map[string]interface{}{
			{
				query.SortBy: map[string]interface{}{
					"order": sortOrder,
				},
			},
		}
	} else {
		// 默认按创建时间倒序
		searchQuery["sort"] = []map[string]interface{}{
			{
				"created_at": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	}

	return searchQuery
}
