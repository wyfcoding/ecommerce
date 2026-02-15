package domain

import (
	"context"
)

// SearchRepository 是搜索模块的仓储接口。
// 它定义了对搜索日志、搜索历史、热门搜索词以及核心搜索和建议功能进行数据持久化和检索的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
// SearchRepository 处理搜索模块的基础数据持久化（日志、历史、热词等）。
// 通常对接 MySQL 等关系型数据库。
type SearchRepository interface {
	SaveSearchLog(ctx context.Context, log *SearchLog) error
	SaveSearchHistory(ctx context.Context, history *SearchHistory) error
	ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*SearchHistory, error)
	DeleteSearchHistory(ctx context.Context, userID uint64) error
	GetHotKeywords(ctx context.Context, limit int) ([]*HotKeyword, error)
}

// SearchEngine 处理商品搜索的核心引擎操作。
// 通常对接 Elasticsearch 或 Vector DB。
type SearchEngine interface {
	// Search 执行商品搜索操作。
	Search(ctx context.Context, filter *SearchFilter) (*SearchResult, error)
	// Suggest 提供搜索建议。
	Suggest(ctx context.Context, keyword string, limit int) ([]*Suggestion, error)
	// Index 将数据同步到搜索索引。
	Index(ctx context.Context, indexName, documentID string, data any) error
	// Delete 从搜索索引中删除数据。
	Delete(ctx context.Context, indexName, documentID string) error
}
