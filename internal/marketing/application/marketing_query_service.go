package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingQueryService 负责营销的查询操作。
type MarketingQueryService struct {
	repo domain.MarketingRepository
}

// NewMarketingQueryService 构造函数。
func NewMarketingQueryService(repo domain.MarketingRepository) *MarketingQueryService {
	return &MarketingQueryService{repo: repo}
}

func (s *MarketingQueryService) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	return s.repo.GetCampaign(ctx, id)
}

func (s *MarketingQueryService) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListCampaigns(ctx, status, offset, pageSize)
}

func (s *MarketingQueryService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return s.repo.ListBanners(ctx, position, activeOnly)
}

func (s *MarketingQueryService) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	return s.repo.GetBanner(ctx, id)
}

func (s *MarketingQueryService) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListParticipations(ctx, campaignID, offset, pageSize)
}
