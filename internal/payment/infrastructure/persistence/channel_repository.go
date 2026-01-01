package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/databases/sharding"
	"gorm.io/gorm"
)

// ChannelRepositoryImpl 渠道配置仓储实现
type ChannelRepositoryImpl struct {
	sharding *sharding.Manager
	tx       *gorm.DB
}

// NewChannelRepository 构造函数
func NewChannelRepository(sharding *sharding.Manager) domain.ChannelRepository {
	return &ChannelRepositoryImpl{sharding: sharding}
}

func (r *ChannelRepositoryImpl) getDB(_ context.Context) *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	// 假设渠道配置存储在全局库或第一个分片
	return r.sharding.GetDB(0)
}

// FindByCode 根据编码查找渠道
func (r *ChannelRepositoryImpl) FindByCode(ctx context.Context, code string) (*domain.ChannelConfig, error) {
	db := r.getDB(ctx)
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
func (r *ChannelRepositoryImpl) ListEnabledByType(ctx context.Context, channelType domain.ChannelType) ([]*domain.ChannelConfig, error) {
	db := r.getDB(ctx)
	var list []*domain.ChannelConfig
	err := db.WithContext(ctx).
		Where("type = ? AND enabled = ?", channelType, true).
		Order("priority DESC").
		Find(&list).Error
	return list, err
}

// Save 保存配置
func (r *ChannelRepositoryImpl) Save(ctx context.Context, channel *domain.ChannelConfig) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Save(channel).Error
}

func (r *ChannelRepositoryImpl) Transaction(ctx context.Context, fn func(tx any) error) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *ChannelRepositoryImpl) WithTx(tx any) domain.ChannelRepository {
	return &ChannelRepositoryImpl{
		sharding: r.sharding,
		tx:       tx.(*gorm.DB),
	}
}
