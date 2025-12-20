package domain

import (
	"context"
)

// MarketingRepository 是营销模块的仓储接口。
type MarketingRepository interface {
	// Campaign
	SaveCampaign(ctx context.Context, campaign *Campaign) error
	GetCampaign(ctx context.Context, id uint64) (*Campaign, error)
	ListCampaigns(ctx context.Context, status CampaignStatus, offset, limit int) ([]*Campaign, int64, error)

	// Participation
	SaveParticipation(ctx context.Context, participation *CampaignParticipation) error
	ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*CampaignParticipation, int64, error)

	// Banner
	SaveBanner(ctx context.Context, banner *Banner) error
	GetBanner(ctx context.Context, id uint64) (*Banner, error)
	ListBanners(ctx context.Context, position string, activeOnly bool) ([]*Banner, error)
	DeleteBanner(ctx context.Context, id uint64) error
}
