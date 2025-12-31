package application

import (
	"context"
	"sort"

	"github.com/wyfcoding/ecommerce/internal/search/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// SearchQuery 处理搜索模块的查询操作。
type SearchQuery struct {
	repo           domain.SearchRepository
	suggestionTrie *algorithm.Trie
}

// NewSearchQuery 创建并返回一个新的 SearchQuery 实例。
func NewSearchQuery(repo domain.SearchRepository) *SearchQuery {
	return &SearchQuery{
		repo:           repo,
		suggestionTrie: algorithm.NewTrie(),
	}
}

// Search 执行搜索操作。
func (q *SearchQuery) Search(ctx context.Context, filter *domain.SearchFilter) (*domain.SearchResult, error) {
	return q.repo.Search(ctx, filter)
}

// Suggest 提供搜索建议。
func (q *SearchQuery) Suggest(ctx context.Context, keyword string, limit int) ([]*domain.Suggestion, error) {
	// 1. 尝试从内存 Trie 中获取建议 (高性能)
	trieResults := q.suggestionTrie.StartsWith(keyword)
	if len(trieResults) > 0 {
		suggestions := make([]*domain.Suggestion, 0, len(trieResults))
		for _, res := range trieResults {
			if s, ok := res.(*domain.Suggestion); ok {
				suggestions = append(suggestions, s)
			}
		}

		// 按分数降序排序
		sort.Slice(suggestions, func(i, j int) bool {
			return suggestions[i].Score > suggestions[j].Score
		})

		if len(suggestions) > limit {
			suggestions = suggestions[:limit]
		}
		return suggestions, nil
	}

	// 2. 如果 Trie 中没有，回退到 Repo (ES/DB)
	return q.repo.Suggest(ctx, keyword, limit)
}

// GetHotKeywords 获取热门搜索词。
func (q *SearchQuery) GetHotKeywords(ctx context.Context, limit int) ([]*domain.HotKeyword, error) {
	return q.repo.GetHotKeywords(ctx, limit)
}

// LoadHotKeywordsToTrie 加载热门搜索词到内存 Trie 中 (通常由定时任务调用)。
func (q *SearchQuery) LoadHotKeywordsToTrie(ctx context.Context) error {
	hotKeywords, err := q.repo.GetHotKeywords(ctx, 1000) // 加载前1000个热词
	if err != nil {
		return err
	}

	// 重建一个新的 Trie 或清空旧的 (这里简化为直接插入，实际可能需要并发控制或替换 Trie 实例)
	// 为了线程安全且原子性，建议创建一个新 Trie 然后替换指针，但 SearchQuery struct 中 trie 是值类型/指针?
	// 这里直接插入，Trie 内部有锁。
	newTrie := algorithm.NewTrie()

	for _, k := range hotKeywords {
		suggestion := &domain.Suggestion{
			Keyword: k.Keyword,
			Score:   k.SearchCount,
			Type:    "hot",
		}
		newTrie.Insert(k.Keyword, suggestion)
	}

	// 替换旧 Trie
	q.suggestionTrie = newTrie
	return nil
}

// ListHistory 获取用户的搜索历史。
func (q *SearchQuery) ListHistory(ctx context.Context, userID uint64, limit int) ([]*domain.SearchHistory, error) {
	return q.repo.ListSearchHistory(ctx, userID, limit)
}
