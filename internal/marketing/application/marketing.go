package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// Marketing 是营销应用服务的门面。
type Marketing struct {
	Command *MarketingCommandService
	Query   *MarketingQueryService
}

// NewMarketing 创建营销服务门面实例。
func NewMarketing(command *MarketingCommandService, query *MarketingQueryService) *Marketing {
	return &Marketing{
		Command: command,
		Query:   query,
	}
}

// CreateCampaign 创建营销活动。
func (s *Marketing) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]any) (*domain.Campaign, error) {
	return s.Command.CreateCampaign(ctx, name, campaignType, description, startTime, endTime, budget, rules)
}

// GetCampaign 根据ID获取营销活动。
func (s *Marketing) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	return s.Query.GetCampaign(ctx, id)
}

// UpdateCampaignStatus 更新活动状态。
func (s *Marketing) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	return s.Command.UpdateCampaignStatus(ctx, id, status)
}

// ListCampaigns 分页列出营销活动。
func (s *Marketing) ListCampaigns(ctx context.Context, status domain.CampaignStatus, page, pageSize int) ([]*domain.Campaign, int64, error) {
	return s.Query.ListCampaigns(ctx, status, page, pageSize)
}

// RecordParticipation 记录用户参与。
func (s *Marketing) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	return s.Command.RecordParticipation(ctx, campaignID, userID, orderID, discount)
}

// CreateBanner 创建广告位。
func (s *Marketing) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*domain.Banner, error) {
	return s.Command.CreateBanner(ctx, title, imageURL, linkURL, position, priority, startTime, endTime)
}

// ListBanners 列出广告位。
func (s *Marketing) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	return s.Query.ListBanners(ctx, position, activeOnly)
}

// GetBanner 获取指定广告位。
func (s *Marketing) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	return s.Query.GetBanner(ctx, id)
}

// ListParticipations 列出活动参与记录。
func (s *Marketing) ListParticipations(ctx context.Context, campaignID uint64, page, pageSize int) ([]*domain.CampaignParticipation, int64, error) {
	return s.Query.ListParticipations(ctx, campaignID, page, pageSize)
}

// DeleteBanner 删除广告位。
func (s *Marketing) DeleteBanner(ctx context.Context, id uint64) error {
	return s.Command.DeleteBanner(ctx, id)
}

// ClickBanner 记录点击。
func (s *Marketing) ClickBanner(ctx context.Context, id uint64) error {
	return s.Command.ClickBanner(ctx, id)
}

// DistributeTargetedCoupons 定向优惠券分发。
func (s *Marketing) DistributeTargetedCoupons(ctx context.Context, couponID uint64, targetTags []string) error {
	return s.Command.DistributeTargetedCoupons(ctx, couponID, targetTags)
}
