// 生成摘要：定义营销活动搜索仓储接口（Elasticsearch）。
package domain

import "context"

// CampaignSearchRepository 定义活动搜索接口。
type CampaignSearchRepository interface {
	Index(ctx context.Context, campaign *Campaign) error
	Delete(ctx context.Context, id uint64) error
	Search(ctx context.Context, status *CampaignStatus, keyword string, offset, limit int) ([]*Campaign, int64, error)
}
