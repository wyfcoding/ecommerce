package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"
)

// MarketingRepository 营销仓储接口
type MarketingRepository interface {
	// 营销活动
	SaveCampaign(ctx context.Context, campaign *entity.Campaign) error
	GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error)
	ListCampaigns(ctx context.Context, status entity.CampaignStatus, offset, limit int) ([]*entity.Campaign, int64, error)

	// 参与记录
	SaveParticipation(ctx context.Context, participation *entity.CampaignParticipation) error
	ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*entity.CampaignParticipation, int64, error)

	// 广告横幅
	SaveBanner(ctx context.Context, banner *entity.Banner) error
	GetBanner(ctx context.Context, id uint64) (*entity.Banner, error)
	ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error)
	DeleteBanner(ctx context.Context, id uint64) error
}
