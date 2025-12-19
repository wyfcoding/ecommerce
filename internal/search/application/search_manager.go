package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/search/domain"
)

// SearchManager 处理搜索模块的写操作、历史记录管理和核心业务逻辑。
type SearchManager struct {
	repo   domain.SearchRepository
	logger *slog.Logger
}

// NewSearchManager 创建并返回一个新的 SearchManager 实例。
func NewSearchManager(repo domain.SearchRepository, logger *slog.Logger) *SearchManager {
	return &SearchManager{
		repo:   repo,
		logger: logger,
	}
}

// SaveLog 保存搜索日志。
func (m *SearchManager) SaveLog(ctx context.Context, log *domain.SearchLog) error {
	if err := m.repo.SaveSearchLog(ctx, log); err != nil {
		m.logger.Error("failed to save search log", "error", err, "user_id", log.UserID, "keyword", log.Keyword)
		return err
	}
	return nil
}

// SaveHistory 保存搜索历史。
func (m *SearchManager) SaveHistory(ctx context.Context, history *domain.SearchHistory) error {
	if err := m.repo.SaveSearchHistory(ctx, history); err != nil {
		m.logger.Error("failed to save search history", "error", err, "user_id", history.UserID, "keyword", history.Keyword)
		return err
	}
	return nil
}

// DeleteHistory 删除搜索历史。
func (m *SearchManager) DeleteHistory(ctx context.Context, userID uint64) error {
	if err := m.repo.DeleteSearchHistory(ctx, userID); err != nil {
		m.logger.Error("failed to delete search history", "error", err, "user_id", userID)
		return err
	}
	return nil
}
