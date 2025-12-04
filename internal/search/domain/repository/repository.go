package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity" // 导入搜索领域的实体定义。
)

// SearchRepository 是搜索模块的仓储接口。
// 它定义了对搜索日志、搜索历史、热门搜索词以及核心搜索和建议功能进行数据持久化和检索的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type SearchRepository interface {
	// --- 搜索日志 (SearchLog methods) ---

	// SaveSearchLog 将搜索日志实体保存到数据存储中。
	// ctx: 上下文。
	// log: 待保存的搜索日志实体。
	SaveSearchLog(ctx context.Context, log *entity.SearchLog) error

	// --- 搜索历史 (SearchHistory methods) ---

	// SaveSearchHistory 将搜索历史实体保存到数据存储中。
	SaveSearchHistory(ctx context.Context, history *entity.SearchHistory) error
	// ListSearchHistory 列出指定用户ID的搜索历史实体，支持数量限制。
	ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error)
	// DeleteSearchHistory 删除指定用户ID的所有搜索历史实体。
	DeleteSearchHistory(ctx context.Context, userID uint64) error

	// --- 热门搜索 (HotKeyword methods) ---

	// GetHotKeywords 获取热门搜索词实体列表，支持数量限制。
	GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error)

	// --- 核心搜索功能 (Search & Suggest methods) ---

	// Search 执行实际的搜索操作，通常对接Elasticsearch等搜索引擎。
	// filter: 搜索过滤器，包含关键词、分页、排序等条件。
	// 返回搜索结果实体。
	Search(ctx context.Context, filter *entity.SearchFilter) (*entity.SearchResult, error)
	// Suggest 提供搜索建议，通常对接Elasticsearch或其他Suggest服务。
	// keyword: 用户输入的关键词前缀。
	// limit: 建议数量限制。
	// 返回搜索建议实体列表。
	Suggest(ctx context.Context, keyword string, limit int) ([]*entity.Suggestion, error)
}
