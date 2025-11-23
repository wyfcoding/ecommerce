package persistence

import (
	"context"
	"ecommerce/internal/search/domain/entity"
	"ecommerce/internal/search/domain/repository"
	"time"

	"gorm.io/gorm"
)

type searchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) repository.SearchRepository {
	return &searchRepository{db: db}
}

// 搜索日志
func (r *searchRepository) SaveSearchLog(ctx context.Context, log *entity.SearchLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// 搜索历史
func (r *searchRepository) SaveSearchHistory(ctx context.Context, history *entity.SearchHistory) error {
	// Upsert or just insert. For simplicity, insert.
	// In real app, might want to deduplicate or update timestamp.
	var existing entity.SearchHistory
	err := r.db.WithContext(ctx).Where("user_id = ? AND keyword = ?", history.UserID, history.Keyword).First(&existing).Error
	if err == nil {
		existing.Timestamp = time.Now()
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *searchRepository) ListSearchHistory(ctx context.Context, userID uint64, limit int) ([]*entity.SearchHistory, error) {
	var list []*entity.SearchHistory
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("timestamp desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *searchRepository) DeleteSearchHistory(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.SearchHistory{}).Error
}

// 热门搜索
func (r *searchRepository) GetHotKeywords(ctx context.Context, limit int) ([]*entity.HotKeyword, error) {
	// This is a simplified implementation using SQL aggregation on logs.
	// In production, this should be cached or computed asynchronously.
	var results []*entity.HotKeyword
	err := r.db.WithContext(ctx).Model(&entity.SearchLog{}).
		Select("keyword, count(*) as search_count").
		Group("keyword").
		Order("search_count desc").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// 核心搜索功能 (Mock for now, as we don't have ES setup in this refactor scope)
func (r *searchRepository) Search(ctx context.Context, filter *entity.SearchFilter) (*entity.SearchResult, error) {
	// Mock implementation
	return &entity.SearchResult{
		Total: 0,
		Items: []interface{}{},
	}, nil
}

func (r *searchRepository) Suggest(ctx context.Context, keyword string, limit int) ([]*entity.Suggestion, error) {
	// Mock implementation based on logs
	var suggestions []*entity.Suggestion
	err := r.db.WithContext(ctx).Model(&entity.SearchLog{}).
		Select("DISTINCT keyword").
		Where("keyword LIKE ?", keyword+"%").
		Limit(limit).
		Scan(&suggestions).Error

	if err != nil {
		return nil, err
	}
	return suggestions, nil
}
