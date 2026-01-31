package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/database/sharding"
	"gorm.io/gorm"
)

// channelRepositoryImpl 渠道配置仓储实现
type channelRepositoryImpl struct {
	sharding *sharding.Manager
	tx       *gorm.DB
}

// NewChannelRepository 构造函数
func NewChannelRepository(sharding *sharding.Manager) domain.ChannelRepository {
	return &channelRepositoryImpl{sharding: sharding}
}

func (r *channelRepositoryImpl) getDB() *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	// 渠道配置通常存储在第一个分片 (shard 0)
	return r.sharding.GetDB(0)
}

// FindByCode 根据编码查找渠道
func (r *channelRepositoryImpl) FindByCode(ctx context.Context, code string) (*domain.ChannelConfig, error) {
	db := r.getDB()
	var config domain.ChannelConfig
	if err := db.WithContext(ctx).Where("code = ?", code).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// ListEnabledByType 获取指定类型的所有启用渠道
func (r *channelRepositoryImpl) ListEnabledByType(ctx context.Context, channelType domain.ChannelType) ([]*domain.ChannelConfig, error) {
	db := r.getDB()
	var list []*domain.ChannelConfig
	err := db.WithContext(ctx).
		Where("type = ? AND enabled = ?", channelType, true).
		Order("priority DESC").
		Find(&list).Error
	return list, err
}

// Save 保存配置
func (r *channelRepositoryImpl) Save(ctx context.Context, channel *domain.ChannelConfig) error {
	db := r.getDB()
	return db.WithContext(ctx).Save(channel).Error
}

func (r *channelRepositoryImpl) Transaction(ctx context.Context, fn func(tx any) error) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *channelRepositoryImpl) WithTx(tx any) domain.ChannelRepository {
	if db, ok := tx.(*gorm.DB); ok {
		return &channelRepositoryImpl{
			sharding: r.sharding,
			tx:       db,
		}
	}
	return r
}
