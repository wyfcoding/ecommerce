// 生成摘要：
// - 从 channel 服务合并到 multichannel 域。
// - 渠道管理与多渠道管理完全同域，消除冗余。
// - 关键实体：SalesChannel（渠道聚合根）。
package domain

import (
	"gorm.io/gorm"
)

// ChannelType 渠道类型枚举。
type ChannelType string

const (
	// TypeApp 原生 App 渠道。
	TypeApp ChannelType = "APP"
	// TypeMiniApp 微信/支付宝小程序渠道。
	TypeMiniApp ChannelType = "MINI_APP"
	// TypeH5 移动端 H5 网页渠道。
	TypeH5 ChannelType = "H5"
	// TypeOffline 线下渠道。
	TypeOffline ChannelType = "OFFLINE"
	// TypePartner 第三方合作渠道 API。
	TypePartner ChannelType = "PARTNER_API"
)

// SalesChannel 渠道聚合根。
// 统一管控 App、微信小程序、H5、第三方分发平台的差异化定价、上架状态与订单来源归属。
type SalesChannel struct {
	gorm.Model
	// ChannelCode 渠道编码，全局唯一。
	ChannelCode string `gorm:"type:varchar(32);uniqueIndex" json:"channel_code"`
	// Name 渠道名称。
	Name string `gorm:"type:varchar(64)" json:"name"`
	// Type 渠道类型。
	Type ChannelType `gorm:"type:varchar(16)" json:"type"`
	// FeeRate 该渠道订单的基础抽佣率。
	FeeRate float64 `gorm:"type:decimal(5,4);comment:该渠道订单的基础抽佣率" json:"fee_rate"`
	// IsActive 渠道是否启用。
	IsActive bool `gorm:"default:true" json:"is_active"`
	// AppIDMapping 绑定的第三方 AppID。
	AppIDMapping string `gorm:"type:varchar(128);comment:绑定的第三方AppID" json:"app_id_mapping"`
	// TokenConfig 渠道鉴权配置 JSON。
	TokenConfig string `gorm:"type:json;comment:渠道鉴权配置" json:"token_config"`
}
