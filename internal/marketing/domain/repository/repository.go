package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity" // 导入营销领域的实体定义。
)

// MarketingRepository 是营销模块的仓储接口。
// 它定义了对营销活动、活动参与记录和广告横幅实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type MarketingRepository interface {
	// --- 营销活动 (Campaign methods) ---

	// SaveCampaign 将营销活动实体保存到数据存储中。
	// 如果活动已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// campaign: 待保存的营销活动实体。
	SaveCampaign(ctx context.Context, campaign *entity.Campaign) error
	// GetCampaign 根据ID获取营销活动实体。
	GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error)
	// ListCampaigns 列出所有营销活动实体，支持通过状态过滤和分页。
	ListCampaigns(ctx context.Context, status entity.CampaignStatus, offset, limit int) ([]*entity.Campaign, int64, error)

	// --- 参与记录 (Participation methods) ---

	// SaveParticipation 将活动参与记录实体保存到数据存储中。
	SaveParticipation(ctx context.Context, participation *entity.CampaignParticipation) error
	// ListParticipations 列出指定营销活动ID的所有参与记录，支持分页。
	ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*entity.CampaignParticipation, int64, error)

	// --- 广告横幅 (Banner methods) ---

	// SaveBanner 将广告横幅实体保存到数据存储中。
	SaveBanner(ctx context.Context, banner *entity.Banner) error
	// GetBanner 根据ID获取广告横幅实体。
	GetBanner(ctx context.Context, id uint64) (*entity.Banner, error)
	// ListBanners 列出所有广告横幅实体，支持通过位置和活跃状态过滤。
	ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error)
	// DeleteBanner 根据ID删除广告横幅实体。
	DeleteBanner(ctx context.Context, id uint64) error
}
