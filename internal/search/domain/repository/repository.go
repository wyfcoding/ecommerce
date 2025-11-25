package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity"
)

// SearchRepository 搜索仓储接口
type SearchRepository interface {
	// 搜索日志
	SaveSearchLog(ctx context.Context, log *entity.SearchLog) error

	// 搜索历史
	SaveSearchHistory(ctx context.Context, history *entity.SearchHistory) error
	ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error)
	DeleteSearchHistory(ctx context.Context, userID uint64) error

	// 热门搜索
	GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error)

	// 核心搜索功能 (通常对接Elasticsearch，这里先定义接口)
	Search(ctx context.Context, filter *entity.SearchFilter) (*entity.SearchResult, error)
	Suggest(ctx context.Context, keyword string, limit int) ([]*entity.Suggestion, error)
}
