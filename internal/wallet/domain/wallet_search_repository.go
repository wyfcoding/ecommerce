// Package domain 钱包搜索仓储接口（CQRS 查询侧 - Elasticsearch）
// 生成摘要：
// 1) 定义钱包交易流水的全文搜索仓储接口
// 2) 支持按时间范围、交易类型、金额范围等多维度检索
// 3) 用于运营后台的交易审计、对账、风控分析
package domain

import (
	"context"
	"time"
)

// TransactionSearchQuery 交易搜索查询条件
type TransactionSearchQuery struct {
	WalletID      uint64     `json:"wallet_id,omitempty"`
	UserID        uint64     `json:"user_id,omitempty"`
	TransactionNo string     `json:"transaction_no,omitempty"`
	Type          string     `json:"type,omitempty"`
	Status        string     `json:"status,omitempty"`
	MinAmount     int64      `json:"min_amount,omitempty"`
	MaxAmount     int64      `json:"max_amount,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Keyword       string     `json:"keyword,omitempty"`
	Page          int        `json:"page"`
	PageSize      int        `json:"page_size"`
	SortBy        string     `json:"sort_by,omitempty"`
	SortOrder     string     `json:"sort_order,omitempty"`
}

// TransactionSearchResult 交易搜索结果
type TransactionSearchResult struct {
	Items    []*TransactionReadModel `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// WalletSearchRepository 钱包搜索仓储接口（Elasticsearch）
type WalletSearchRepository interface {
	// IndexTransaction 索引交易记录到 ES。
	IndexTransaction(ctx context.Context, tx *TransactionReadModel) error
	// SearchTransactions 多维度搜索交易记录。
	SearchTransactions(ctx context.Context, query *TransactionSearchQuery) (*TransactionSearchResult, error)
	// AggregateByType 按交易类型聚合统计。
	AggregateByType(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error)
	// AggregateDaily 按日聚合交易金额。
	AggregateDaily(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error)
}
