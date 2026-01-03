package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/search/domain"
	pkgsearch "github.com/wyfcoding/pkg/search"
)

// SearchManager 处理搜索模块的写操作、历史记录管理和核心业务逻辑。
type SearchManager struct {
	repo     domain.SearchRepository
	esClient *pkgsearch.Client
	logger   *slog.Logger
}

// NewSearchManager 创建并返回一个新的 SearchManager 实例。
func NewSearchManager(repo domain.SearchRepository, esClient *pkgsearch.Client, logger *slog.Logger) *SearchManager {
	return &SearchManager{
		repo:     repo,
		esClient: esClient,
		logger:   logger.With("module", "search_manager"),
	}
}

// SyncProductIndex 处理来自 MQ 的商品同步事件，更新 ES 索引
func (m *SearchManager) SyncProductIndex(ctx context.Context, event map[string]any) error {
	action := event["action"].(string)
	productID := fmt.Sprintf("%v", event["product_id"])
	indexName := "products" // 索引名通常在配置中定义

	m.logger.Info("syncing product index", "action", action, "product_id", productID)

	switch action {
	case "create", "update":
		// 执行索引或更新
		// 消息体中的 event 数据直接作为文档存入 ES
		if err := m.esClient.Index(ctx, indexName, productID, event); err != nil {
			m.logger.Error("failed to index product", "product_id", productID, "error", err)
			return err
		}
	case "delete":
		// 执行删除
		if err := m.esClient.Delete(ctx, indexName, productID); err != nil {
			m.logger.Error("failed to delete product index", "product_id", productID, "error", err)
			return err
		}
	default:
		m.logger.Warn("unknown sync action", "action", action)
	}

	return nil
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
