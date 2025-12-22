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

// NewMarketingService 创建营销服务门面实例。
func NewMarketingService(manager *MarketingManager, query *MarketingQuery) *MarketingService {
	return &MarketingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateCampaign 创建一个新的营销活动。
func (s *MarketingService) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]any) (*domain.Campaign, error) {
	return s.manager.CreateCampaign(ctx, name, campaignType, description, startTime, endTime, budget, rules)
}

// UpdateCampaignStatus 更新营销活动的状态。
func (s *MarketingService) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	return s.manager.UpdateCampaignStatus(ctx, id, status)
}

// RecordParticipation 记录用户参与营销活动的情况。
func (s *MarketingService) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	return s.manager.RecordParticipation(ctx, campaignID, userID, orderID, discount)
}

// CreateBanner 创建一个广告位/Banner。
func (s *MarketingService) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*domain.Banner, error) {
	return s.manager.CreateBanner(ctx, title, imageURL, linkURL, position, priority, startTime, endTime)
}

// DeleteBanner 删除指定的广告位/Banner。
func (s *MarketingService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.manager.DeleteBanner(ctx, id)
}

// ClickBanner 记录广告位/Banner 的点击事件。
func (s *MarketingService) ClickBanner(ctx context.Context, id uint64) error {
	return s.manager.ClickBanner(ctx, id)
}

// --- 读操作（委托给 Query）---

// GetCampaign 获取指定ID的营销活动详情。
func (s *MarketingService) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	return s.query.GetCampaign(ctx, id)
}

// ListCampaigns 获取营销活动列表。
func (s *MarketingService) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	return s.query.ListCampaigns(ctx, status, page, pageSize)
}

// ListParticipations 获取指定营销活动的参与记录。
func (s *MarketingService) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	return s.query.ListParticipations(ctx, campaignID, page, pageSize)
}

// GetBanner 获取指定ID的广告位/Banner详情。
func (s *MarketingService) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	return s.query.GetBanner(ctx, id)
}

// ListBanners 获取指定位置的广告位/Banner列表。
func (s *MarketingService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return s.query.ListBanners(ctx, position, activeOnly)
}
