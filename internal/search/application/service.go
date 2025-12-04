package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/search/domain/entity"     // 导入搜索领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/search/domain/repository" // 导入搜索领域的仓储接口。

	"log/slog"
)

// SearchService 结构体定义了商品搜索相关的应用服务。
// 它协调领域层和基础设施层，处理搜索请求的执行、搜索行为的记录、搜索历史的维护和搜索建议的提供等业务逻辑。
type SearchService struct {
	repo   repository.SearchRepository // 依赖SearchRepository接口，用于数据持久化操作。
	logger *slog.Logger                // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewSearchService 创建并返回一个新的 SearchService 实例。
func NewSearchService(repo repository.SearchRepository, logger *slog.Logger) *SearchService {
	return &SearchService{
		repo:   repo,
		logger: logger,
	}
}

// Search 执行搜索操作，并记录搜索日志和搜索历史。
// ctx: 上下文。
// userID: 搜索用户ID。
// filter: 搜索过滤器，包含关键词、分页、排序等。
// 返回搜索结果SearchResult实体和可能发生的错误。
func (s *SearchService) Search(ctx context.Context, userID uint64, filter *entity.SearchFilter) (*entity.SearchResult, error) {
	start := time.Now() // 记录搜索开始时间，用于计算耗时。

	// 1. 执行实际搜索操作。
	result, err := s.repo.Search(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute search", "keyword", filter.Keyword, "error", err)
		return nil, err
	}

	// 2. 异步记录搜索日志和搜索历史。
	// 在生产环境中，这些记录操作通常是异步“即发即弃”（Fire and forget）的，以避免影响主搜索流程的性能。
	// 此处为简化，直接调用并忽略错误。
	if filter.Keyword != "" { // 只有当有关键词时才记录。
		// 保存搜索日志。
		_ = s.repo.SaveSearchLog(ctx, &entity.SearchLog{
			UserID:      userID,
			Keyword:     filter.Keyword,
			ResultCount: int(result.Total),                // 搜索结果总数。
			Duration:    time.Since(start).Milliseconds(), // 搜索耗时。
		})

		if userID > 0 { // 只有登录用户才记录搜索历史。
			// 保存搜索历史。
			_ = s.repo.SaveSearchHistory(ctx, &entity.SearchHistory{
				UserID:    userID,
				Keyword:   filter.Keyword,
				Timestamp: time.Now(),
			})
		}
	}
	s.logger.InfoContext(ctx, "search completed", "keyword", filter.Keyword, "user_id", userID, "result_count", result.Total, "duration_ms", time.Since(start).Milliseconds())

	return result, nil
}

// GetHotKeywords 获取热搜词列表。
// ctx: 上下文。
// limit: 限制返回的热搜词数量。
// 返回热搜词实体列表和可能发生的错误。
func (s *SearchService) GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error) {
	return s.repo.GetHotKeywords(ctx, limit)
}

// GetSearchHistory 获取指定用户的搜索历史。
// ctx: 上下文。
// userID: 用户ID。
// limit: 限制返回的搜索历史数量。
// 返回搜索历史实体列表和可能发生的错误。
func (s *SearchService) GetSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error) {
	return s.repo.ListSearchHistory(ctx, userID, limit)
}

// ClearSearchHistory 清空指定用户的搜索历史。
// ctx: 上下文。
// userID: 用户ID。
// 返回可能发生的错误。
func (s *SearchService) ClearSearchHistory(ctx context.Context, userID uint64) error {
	return s.repo.DeleteSearchHistory(ctx, userID)
}

// Suggest 提供搜索建议。
// ctx: 上下文。
// keyword: 用户输入的关键词前缀。
// 返回搜索建议实体列表和可能发生的错误。
func (s *SearchService) Suggest(ctx context.Context, keyword string) ([]*entity.Suggestion, error) {
	// 默认返回10条建议。
	return s.repo.Suggest(ctx, keyword, 10)
}
