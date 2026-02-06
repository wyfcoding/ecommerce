package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingQueryService 负责营销的查询操作。
type MarketingQueryService struct {
	repo               domain.MarketingRepository
	campaignReadRepo   domain.CampaignReadRepository
	bannerReadRepo     domain.BannerReadRepository
	campaignSearchRepo domain.CampaignSearchRepository
}

// NewMarketingQueryService 构造函数。
func NewMarketingQueryService(
	repo domain.MarketingRepository,
	campaignReadRepo domain.CampaignReadRepository,
	bannerReadRepo domain.BannerReadRepository,
	campaignSearchRepo domain.CampaignSearchRepository,
) *MarketingQueryService {
	return &MarketingQueryService{
		repo:               repo,
		campaignReadRepo:   campaignReadRepo,
		bannerReadRepo:     bannerReadRepo,
		campaignSearchRepo: campaignSearchRepo,
	}
}

func (s *MarketingQueryService) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	if s.campaignReadRepo != nil {
		if cached, err := s.campaignReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	campaign, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if campaign != nil && s.campaignReadRepo != nil {
		_ = s.campaignReadRepo.Save(ctx, campaign)
	}
	return campaign, nil
}

func (s *MarketingQueryService) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	offset := (page - 1) * pageSize
	var statusPtr *domain.CampaignStatus
	if status != 0 {
		statusPtr = &status
	}
	if s.campaignSearchRepo != nil {
		list, total, err := s.campaignSearchRepo.Search(ctx, statusPtr, "", offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return s.repo.ListCampaigns(ctx, status, offset, pageSize)
}

func (s *MarketingQueryService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return s.repo.ListBanners(ctx, position, activeOnly)
}

func (s *MarketingQueryService) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	if s.bannerReadRepo != nil {
		if cached, err := s.bannerReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	banner, err := s.repo.GetBanner(ctx, id)
	if err != nil {
		return nil, err
	}
	if banner != nil && s.bannerReadRepo != nil {
		_ = s.bannerReadRepo.Save(ctx, banner)
	}
	return banner, nil
}

func (s *MarketingQueryService) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListParticipations(ctx, campaignID, offset, pageSize)
}
