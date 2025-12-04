package persistence

import (
	"context"
	"time" // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/search/domain/entity"     // 导入搜索领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/search/domain/repository" // 导入搜索领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type searchRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSearchRepository 创建并返回一个新的 searchRepository 实例。
func NewSearchRepository(db *gorm.DB) repository.SearchRepository {
	return &searchRepository{db: db}
}

// --- 搜索日志 (SearchLog methods) ---

// SaveSearchLog 将搜索日志实体保存到数据库。
func (r *searchRepository) SaveSearchLog(ctx context.Context, log *entity.SearchLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// --- 搜索历史 (SearchHistory methods) ---

// SaveSearchHistory 将搜索历史实体保存到数据库。
// 如果相同用户和关键词的记录已存在，则更新其时间戳；否则创建新记录。
func (r *searchRepository) SaveSearchHistory(ctx context.Context, history *entity.SearchHistory) error {
	var existing entity.SearchHistory
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
func (r *searchRepository) ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error) {
	var list []*entity.SearchHistory
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("timestamp desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteSearchHistory 从数据库删除指定用户ID的所有搜索历史记录。
func (r *searchRepository) DeleteSearchHistory(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.SearchHistory{}).Error
}

// --- 热门搜索 (HotKeyword methods) ---

// GetHotKeywords 从搜索日志中聚合计算热门搜索词列表。
// 注意：这是一个简化的实现，直接从SearchLog进行SQL聚合。在生产环境中，出于性能考虑，
// 通常会使用缓存（例如Redis）或者异步任务预计算热门搜索词。
func (r *searchRepository) GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error) {
	var results []*entity.HotKeyword
	err := r.db.WithContext(ctx).Model(&entity.SearchLog{}).
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

// Search 执行搜索操作。
// 备注：这是一个模拟实现，因为完整的Elasticsearch等搜索引擎集成超出了当前重构范围。
// 实际生产中，此方法会调用外部搜索引擎API。
func (r *searchRepository) Search(ctx context.Context, filter *entity.SearchFilter) (*entity.SearchResult, error) {
	// 模拟实现，总是返回空结果。
	return &entity.SearchResult{
		Total: 0,
		Items: []interface{}{},
	}, nil
}

// Suggest 提供搜索建议。
// 备注：这是一个模拟实现，实际生产中会使用搜索引擎的Suggest功能或预计算的建议数据。
func (r *searchRepository) Suggest(ctx context.Context, keyword string, limit int) ([]*entity.Suggestion, error) {
	// 模拟实现，从SearchLog中查找以指定关键词开头的不同关键词作为建议。
	var suggestions []*entity.Suggestion
	err := r.db.WithContext(ctx).Model(&entity.SearchLog{}).
		Select("DISTINCT keyword").           // 选择不重复的关键词。
		Where("keyword LIKE ?", keyword+"%"). // 查找以指定关键词开头的记录。
		Limit(limit).                         // 应用数量限制。
		Scan(&suggestions).Error              // 将结果扫描到Suggestion结构体中。
	// 这里将keyword直接映射到Suggestion的Keyword字段，Type和Score需要补充。
	// TODO: 如果需要填充Score和Type，需要修改查询或后续处理逻辑。

	if err != nil {
		return nil, err
	}
	return suggestions, nil
}
