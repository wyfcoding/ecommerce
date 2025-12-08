package domain

import (
	"context"

	"gorm.io/gorm"
)

// ChannelType 渠道类型
type ChannelType string

const (
	ChannelTypeAlipay ChannelType = "alipay"
	ChannelTypeWechat ChannelType = "wechat"
	ChannelTypeStripe ChannelType = "stripe"
)

// ChannelConfig 支付渠道配置实体
// 存储渠道的密钥、证书路径、费率等信息
type ChannelConfig struct {
	gorm.Model
	Code        string      `gorm:"uniqueIndex;size:32;not null" json:"code"` // 渠道编码 (e.g., "alipay_global_1")
	Type        ChannelType `gorm:"size:32;not null" json:"type"`             // 渠道类型 (alipay, wechat)
	Name        string      `gorm:"size:64" json:"name"`                      // 渠道名称 (e.g., "Global Alipay Account")
	Priority    int         `gorm:"default:0" json:"priority"`                // 优先级 (数字越大越优先)
	Enabled     bool        `gorm:"default:true" json:"enabled"`              // 是否启用
	ConfigJSON  string      `gorm:"type:text" json:"config_json"`             // JSON配置 (AppID, PrivateKey, etc.)
	RatePercent float64     `gorm:"type:decimal(5,2)" json:"rate_percent"`    // 费率百分比 (用于智能路由)
	Description string      `gorm:"size:255" json:"description"`
}

// ChannelRepository 渠道配置仓储接口
type ChannelRepository interface {
	// FindByCode 根据编码查找渠道配置
	FindByCode(ctx context.Context, code string) (*ChannelConfig, error)

	// ListEnabledByType 获取某类型下所有启用的渠道（按优先级排序）
	ListEnabledByType(ctx context.Context, channelType ChannelType) ([]*ChannelConfig, error)

	// Save 保存渠道配置
	Save(ctx context.Context, channel *ChannelConfig) error
}
