package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/search/domain"
)

// Search 结构体定义了商品搜索相关的应用服务（外观模式）。
// 它协调 SearchManager 和 SearchQuery 处理搜索请求的执行、行为记录、历史维护和搜索建议。
type Search struct {
	manager *SearchManager
	query   *SearchQuery
}

// NewSearch 创建并返回一个新的 Search 实例。
func NewSearch(manager *SearchManager, query *SearchQuery) *Search {
	return &Search{
		manager: manager,
		query:   query,
	}
}

// Search 执行搜索操作，并记录搜索日志和搜索历史。
func (s *Search) Search(ctx context.Context, userID uint64, filter *domain.SearchFilter) (*domain.SearchResult, error) {
	start := time.Now()

	// 1. 执行实际搜索操作 (Query)。
	result, err := s.query.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 2. 异步记录搜索日志和搜索历史 (Manager)。
	if filter.Keyword != "" {
		// 保存搜索日志。
		_ = s.manager.SaveLog(ctx, &domain.SearchLog{
			UserID:      userID,
			Keyword:     filter.Keyword,
			ResultCount: int(result.Total),
			Duration:    time.Since(start).Milliseconds(),
		})

		if userID > 0 {
			// 保存搜索历史。
			_ = s.manager.SaveHistory(ctx, &domain.SearchHistory{
				UserID:    userID,
				Keyword:   filter.Keyword,
				Timestamp: time.Now(),
			})
		}
	}

	return result, nil
}

// GetHotKeywords 获取热搜词列表。
func (s *Search) GetHotKeywords(ctx context.Context, limit int) ([]*domain.HotKeyword, error) {
	return s.query.GetHotKeywords(ctx, limit)
}

// GetSearchHistory 获取指定用户的搜索历史。
func (s *Search) GetSearchHistory(ctx context.Context, userID uint64, limit int) ([]*domain.SearchHistory, error) {
	return s.query.ListHistory(ctx, userID, limit)
}

// ClearSearchHistory 清空指定用户的搜索历史。
func (s *Search) ClearSearchHistory(ctx context.Context, userID uint64) error {
	return s.manager.DeleteHistory(ctx, userID)
}

// Suggest 提供搜索建议。
func (s *Search) Suggest(ctx context.Context, keyword string) ([]*domain.Suggestion, error) {
	return s.query.Suggest(ctx, keyword, 10)
}
