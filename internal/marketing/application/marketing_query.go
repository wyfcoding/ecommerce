package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingQuery handles read operations for marketing.
type MarketingQuery struct {
	repo domain.MarketingRepository
}

// NewMarketingQuery creates a new MarketingQuery instance.
func NewMarketingQuery(repo domain.MarketingRepository) *MarketingQuery {
	return &MarketingQuery{
		repo: repo,
	}
}

// GetCampaign 获取指定ID的营销活动详情。
func (q *MarketingQuery) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	return q.repo.GetCampaign(ctx, id)
}

// ListCampaigns 获取营销活动列表。
func (q *MarketingQuery) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListCampaigns(ctx, status, offset, pageSize)
}

// ListParticipations 获取参与记录列表。
func (q *MarketingQuery) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListParticipations(ctx, campaignID, offset, pageSize)
}

// GetBanner 获取Banner详情。
func (q *MarketingQuery) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	return q.repo.GetBanner(ctx, id)
}

// ListBanners 获取Banner列表。
func (q *MarketingQuery) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return q.repo.ListBanners(ctx, position, activeOnly)
}
