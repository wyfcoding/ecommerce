package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/search/domain"
)

// SearchQuery 处理搜索模块的查询操作。
type SearchQuery struct {
	repo domain.SearchRepository
}

// NewSearchQuery 创建并返回一个新的 SearchQuery 实例。
func NewSearchQuery(repo domain.SearchRepository) *SearchQuery {
	return &SearchQuery{repo: repo}
}

// Search 执行搜索操作。
func (q *SearchQuery) Search(ctx context.Context, filter *domain.SearchFilter) (*domain.SearchResult, error) {
	return q.repo.Search(ctx, filter)
}

// Suggest 提供搜索建议。
func (q *SearchQuery) Suggest(ctx context.Context, keyword string, limit int) ([]*domain.Suggestion, error) {
	return q.repo.Suggest(ctx, keyword, limit)
}

// GetHotKeywords 获取热门搜索词。
func (q *SearchQuery) GetHotKeywords(ctx context.Context, limit int) ([]*domain.HotKeyword, error) {
	return q.repo.GetHotKeywords(ctx, limit)
}

// ListHistory 获取用户的搜索历史。
func (q *SearchQuery) ListHistory(ctx context.Context, userID uint64, limit int) ([]*domain.SearchHistory, error) {
	return q.repo.ListSearchHistory(ctx, userID, limit)
}
