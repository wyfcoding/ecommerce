package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/search/domain/repository"
	"time"

	"log/slog"
)

type SearchService struct {
	repo   repository.SearchRepository
	logger *slog.Logger
}

func NewSearchService(repo repository.SearchRepository, logger *slog.Logger) *SearchService {
	return &SearchService{
		repo:   repo,
		logger: logger,
	}
}

// Search 执行搜索
func (s *SearchService) Search(ctx context.Context, userID uint64, filter *entity.SearchFilter) (*entity.SearchResult, error) {
	start := time.Now()

	// 1. Execute Search
	result, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 2. Async Log (Fire and forget in production, sync here for simplicity)
	if filter.Keyword != "" {
		_ = s.repo.SaveSearchLog(ctx, &entity.SearchLog{
			UserID:      userID,
			Keyword:     filter.Keyword,
			ResultCount: int(result.Total),
			Duration:    time.Since(start).Milliseconds(),
		})

		if userID > 0 {
			_ = s.repo.SaveSearchHistory(ctx, &entity.SearchHistory{
				UserID:    userID,
				Keyword:   filter.Keyword,
				Timestamp: time.Now(),
			})
		}
	}

	return result, nil
}

// GetHotKeywords 获取热搜词
func (s *SearchService) GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error) {
	return s.repo.GetHotKeywords(ctx, limit)
}

// GetSearchHistory 获取搜索历史
func (s *SearchService) GetSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error) {
	return s.repo.ListSearchHistory(ctx, userID, limit)
}

// ClearSearchHistory 清空搜索历史
func (s *SearchService) ClearSearchHistory(ctx context.Context, userID uint64) error {
	return s.repo.DeleteSearchHistory(ctx, userID)
}

// Suggest 搜索建议
func (s *SearchService) Suggest(ctx context.Context, keyword string) ([]*entity.Suggestion, error) {
	return s.repo.Suggest(ctx, keyword, 10)
}
