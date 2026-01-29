package domain

import (
	"context"
)

// MarketingRepository 是营销模块的仓储接口。
type MarketingRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// Campaign
	SaveCampaign(ctx context.Context, campaign *Campaign) error
	SaveCampaignInTx(ctx context.Context, tx any, campaign *Campaign) error
	GetCampaign(ctx context.Context, id uint64) (*Campaign, error)
	ListCampaigns(ctx context.Context, status CampaignStatus, offset, limit int) ([]*Campaign, int64, error)

	// Participation
	SaveParticipation(ctx context.Context, participation *CampaignParticipation) error
	SaveParticipationInTx(ctx context.Context, tx any, participation *CampaignParticipation) error
	ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*CampaignParticipation, int64, error)

	// Banner
	SaveBanner(ctx context.Context, banner *Banner) error
	SaveBannerInTx(ctx context.Context, tx any, banner *Banner) error
	GetBanner(ctx context.Context, id uint64) (*Banner, error)
	ListBanners(ctx context.Context, position string, activeOnly bool) ([]*Banner, error)
	DeleteBanner(ctx context.Context, id uint64) error

	// User Tags
	GetUserIDsByTag(ctx context.Context, tagName string) ([]uint32, error)
}
