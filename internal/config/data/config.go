package data

import (
	"context"
	"ecommerce/internal/config/biz"
	"ecommerce/internal/config/data/model"
	"time"

	"gorm.io/gorm"
)

type configRepo struct {
	data *Data
}

// NewConfigRepo creates a new ConfigRepo.
func NewConfigRepo(data *Data) biz.ConfigRepo {
	return &configRepo{data: data}
}

// GetConfig retrieves a configuration entry by its key.
func (r *configRepo) GetConfig(ctx context.Context, key string) (*biz.ConfigEntry, error) {
	var po model.ConfigEntry
	if err := r.data.db.WithContext(ctx).Where("key = ?", key).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Config not found
		}
		return nil, err
	}
	return &biz.ConfigEntry{
		ID:          po.ID,
		Key:         po.Key,
		Value:       po.Value,
		Description: po.Description,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}, nil
}

// SetConfig creates or updates a configuration entry.
func (r *configRepo) SetConfig(ctx context.Context, entry *biz.ConfigEntry) (*biz.ConfigEntry, error) {
	po := &model.ConfigEntry{
		Key:         entry.Key,
		Value:       entry.Value,
		Description: entry.Description,
	}
	// Try to find existing record
	var existing model.ConfigEntry
	if err := r.data.db.WithContext(ctx).Where("key = ?", entry.Key).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Not found, create new
			if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
				return nil, err
			}
			entry.ID = po.ID
			return entry, nil
		}
		return nil, err
	}

	// Found, update existing
	if err := r.data.db.WithContext(ctx).Model(&existing).Updates(po).Error; err != nil {
		return nil, err
	}
	entry.ID = existing.ID
	return entry, nil
}

// ListConfigs lists all configuration entries.
func (r *configRepo) ListConfigs(ctx context.Context) ([]*biz.ConfigEntry, error) {
	var entries []*model.ConfigEntry
	if err := r.data.db.WithContext(ctx).Find(&entries).Error; err != nil {
		return nil, err
	}
	bizEntries := make([]*biz.ConfigEntry, len(entries))
	for i, entry := range entries {
		bizEntries[i] = &biz.ConfigEntry{
			ID:          entry.ID,
			Key:         entry.Key,
			Value:       entry.Value,
			Description: entry.Description,
			CreatedAt:   entry.CreatedAt,
			UpdatedAt:   entry.UpdatedAt,
		}
	}
	return bizEntries, nil
}
