package persistence

import (
	"context"
	"time" // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/search/domain" // 导入搜索领域的领域定义。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type searchRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSearchRepository 创建并返回一个新的 searchRepository 实例。
func NewSearchRepository(db *gorm.DB) domain.SearchRepository {
	return &searchRepository{db: db}
}

// --- 搜索日志 (SearchLog methods) ---

// SaveSearchLog 将搜索日志实体保存到数据库。
func (r *searchRepository) SaveSearchLog(ctx context.Context, log *domain.SearchLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// --- 搜索历史 (SearchHistory methods) ---

// SaveSearchHistory 将搜索历史实体保存到数据库。
// 如果相同用户和关键词的记录已存在，则更新其时间戳；否则创建新记录。
func (r *searchRepository) SaveSearchHistory(ctx context.Context, history *domain.SearchHistory) error {
	var existing domain.SearchHistory
	// 尝试查找现有记录。
	err := r.db.WithContext(ctx).Where("user_id = ? AND keyword = ?", history.UserID, history.Keyword).First(&existing).Error
	if err == nil {
		// 如果找到现有记录，则更新其时间戳。
		existing.Timestamp = time.Now()
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	// 如果未找到记录，则创建新记录。
	return r.db.WithContext(ctx).Create(history).Error
}

// ListSearchHistory 从数据库列出指定用户ID的搜索历史记录，按时间降序排列，并应用数量限制。
func (r *searchRepository) ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*domain.SearchHistory, error) {
	var list []*domain.SearchHistory
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("timestamp desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteSearchHistory 从数据库删除指定用户ID的所有搜索历史记录。
func (r *searchRepository) DeleteSearchHistory(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.SearchHistory{}).Error
}

// --- 热门搜索 (HotKeyword methods) ---

// GetHotKeywords 从搜索日志中聚合计算热门搜索词列表。
func (r *searchRepository) GetHotKeywords(ctx context.Context, limit int) ([]*domain.HotKeyword, error) {
	var results []*domain.HotKeyword
	err := r.db.WithContext(ctx).Model(&domain.SearchLog{}).
		Select("keyword, count(*) as search_count"). // 选择关键词和搜索计数。
		Group("keyword").                            // 按关键词分组。
		Order("search_count desc").                  // 按搜索计数降序排列。
		Limit(limit).                                // 应用数量限制。
		Scan(&results).Error                         // 将结果扫描到HotKeyword结构体中。
	if err != nil {
		return nil, err
	}
	return results, nil
}

// --- 核心搜索功能 (Search & Suggest methods) ---

// Search 执行基于数据库的模糊搜索。
func (r *searchRepository) Search(ctx context.Context, filter *domain.SearchFilter) (*domain.SearchResult, error) {
	// 真实化实现：从产品表中搜索
	var products []struct {
		ID          uint64 `gorm:"column:id"`
		Name        string `gorm:"column:name"`
		Description string `gorm:"column:description"`
	}
	var total int64

	db := r.db.WithContext(ctx).Table("products")

	if filter.Keyword != "" {
		likeQuery := "%" + filter.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR category LIKE ?", likeQuery, likeQuery, likeQuery)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := db.Offset(offset).Limit(filter.PageSize).Find(&products).Error
	if err != nil {
		return nil, err
	}

	// 转换为通用的 any 列表返回
	items := make([]any, len(products))
	for i, p := range products {
		items[i] = p
	}

	return &domain.SearchResult{
		Total: total,
		Items: items,
	}, nil
}

// Suggest 提供搜索建议。
func (r *searchRepository) Suggest(ctx context.Context, keyword string, limit int) ([]*domain.Suggestion, error) {
	// 真实实现：从 SearchLog 中查找以指定关键词开头的关键词，并按频率排序。
	var suggestions []*domain.Suggestion
	err := r.db.WithContext(ctx).Model(&domain.SearchLog{}).
		Select("keyword, COUNT(*) as score").
		Where("keyword LIKE ?", keyword+"%").
		Group("keyword").
		Order("score DESC").
		Limit(limit).
		Scan(&suggestions).Error
	if err != nil {
		return nil, err
	}
	return suggestions, nil
}
