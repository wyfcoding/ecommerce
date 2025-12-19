package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/databases/sharding"
	"gorm.io/gorm"
)

// ChannelRepositoryImpl 渠道配置仓储实现
type ChannelRepositoryImpl struct {
	sharding *sharding.Manager
}

// NewChannelRepository 构造函数
func NewChannelRepository(sharding *sharding.Manager) domain.ChannelRepository {
	return &ChannelRepositoryImpl{sharding: sharding}
}

func (r *ChannelRepositoryImpl) getDB(_ context.Context) *gorm.DB {
	// 假设渠道配置存储在全局库或第一个分片
	return r.sharding.GetDB(0)
}

// FindByCode 根据编码查找渠道
func (r *ChannelRepositoryImpl) FindByCode(ctx context.Context, code string) (*domain.ChannelConfig, error) {
	// 实际应查询数据库
	// return r.mockChannel(code), nil
	return r.mockChannel(code), nil
}

// ListEnabledByType 获取指定类型的所有启用渠道
func (r *ChannelRepositoryImpl) ListEnabledByType(ctx context.Context, channelType domain.ChannelType) ([]*domain.ChannelConfig, error) {
	// 模拟返回
	if channelType == domain.ChannelTypeAlipay {
		return []*domain.ChannelConfig{
			r.mockChannel("alipay_global_1"),
			r.mockChannel("alipay_local_1"),
		}, nil
	}
	return []*domain.ChannelConfig{}, nil
}

// Save 保存配置
func (r *ChannelRepositoryImpl) Save(ctx context.Context, channel *domain.ChannelConfig) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Save(channel).Error
}

// mockChannel 返回模拟数据
func (r *ChannelRepositoryImpl) mockChannel(code string) *domain.ChannelConfig {
	cfg := &domain.ChannelConfig{
		Code:    code,
		Enabled: true,
	}

	switch code {
	case "alipay_global_1":
		cfg.Type = domain.ChannelTypeAlipay
		cfg.Name = "Alipay Global"
		cfg.ConfigJSON = `{"app_id": "mock_app_id", "private_key": "mock_key"}`
		cfg.Priority = 100
	case "wechat_app":
		cfg.Type = domain.ChannelTypeWechat
		cfg.Name = "Wechat App"
		cfg.ConfigJSON = `{"mch_id": "mock_mch"}`
	default:
		cfg.Type = domain.ChannelTypeAlipay // 默认
		cfg.Name = "Default Channel"
	}

	return cfg
}
