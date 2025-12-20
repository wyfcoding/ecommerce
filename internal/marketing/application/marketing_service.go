package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingService 作为营销操作的门面。
type MarketingService struct {
	manager *MarketingManager
	query   *MarketingQuery
}

// NewMarketingService creates a new MarketingService facade.
func NewMarketingService(manager *MarketingManager, query *MarketingQuery) *MarketingService {
	return &MarketingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *MarketingService) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]interface{}) (*domain.Campaign, error) {
	return s.manager.CreateCampaign(ctx, name, campaignType, description, startTime, endTime, budget, rules)
}

func (s *MarketingService) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	return s.manager.UpdateCampaignStatus(ctx, id, status)
}

func (s *MarketingService) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	return s.manager.RecordParticipation(ctx, campaignID, userID, orderID, discount)
}

func (s *MarketingService) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*domain.Banner, error) {
	return s.manager.CreateBanner(ctx, title, imageURL, linkURL, position, priority, startTime, endTime)
}

func (s *MarketingService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.manager.DeleteBanner(ctx, id)
}

func (s *MarketingService) ClickBanner(ctx context.Context, id uint64) error {
	return s.manager.ClickBanner(ctx, id)
}

// --- 读操作（委托给 Query）---

func (s *MarketingService) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	return s.query.GetCampaign(ctx, id)
}

func (s *MarketingService) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	return s.query.ListCampaigns(ctx, status, page, pageSize)
}

func (s *MarketingService) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	return s.query.ListParticipations(ctx, campaignID, page, pageSize)
}

func (s *MarketingService) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	return s.query.GetBanner(ctx, id)
}

func (s *MarketingService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return s.query.ListBanners(ctx, position, activeOnly)
}
