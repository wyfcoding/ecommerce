// 生成摘要：定义营销活动读模型仓储接口（Redis）。
package domain

import "context"

// CampaignReadRepository 读取活动详情的缓存接口。
type CampaignReadRepository interface {
	Save(ctx context.Context, campaign *Campaign) error
	GetByID(ctx context.Context, id uint64) (*Campaign, error)
	Delete(ctx context.Context, id uint64) error
}
