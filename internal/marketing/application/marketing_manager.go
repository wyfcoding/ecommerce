package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	couponv1 "github.com/wyfcoding/ecommerce/goapi/coupon/v1"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
)

// MarketingManager 处理营销的写操作。
type MarketingManager struct {
	repo           domain.MarketingRepository
	logger         *slog.Logger
	userFilter     *algorithm.BloomFilter
	segmentService *UserSegmentService // 接入基于 Roaring Bitmap 的圈选服务
}

// NewMarketingManager creates a new MarketingManager instance.
func NewMarketingManager(repo domain.MarketingRepository, couponCli couponv1.CouponServiceClient, logger *slog.Logger) (*MarketingManager, error) {
	filter, err := algorithm.NewBloomFilter(1000000, 0.01)
	if err != nil {
		return nil, err
	}

	return &MarketingManager{
		repo:           repo,
		logger:         logger,
		userFilter:     filter,
		segmentService: NewUserSegmentService(repo, couponCli, logger), // 初始化圈选服务并注入优惠券客户端
	}, nil
}

// DistributeTargetedCoupons 定向优惠券分发：顶级架构实战
func (m *MarketingManager) DistributeTargetedCoupons(ctx context.Context, couponID uint64, targetTags []string) error {
	m.logger.InfoContext(ctx, "starting real targeted coupon distribution", "coupon_id", couponID, "tags", targetTags)

	// 调用增强后的分发服务
	return m.segmentService.DistributeCouponsToSegment(ctx, couponID, targetTags)
}

// CreateCampaign 创建一个新的营销活动。
func (m *MarketingManager) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]any) (*domain.Campaign, error) {
	campaign := domain.NewCampaign(name, campaignType, description, startTime, endTime, budget, rules)
	if err := m.repo.SaveCampaign(ctx, campaign); err != nil {
		m.logger.ErrorContext(ctx, "failed to create campaign", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "campaign created successfully", "campaign_id", campaign.ID, "name", name)
	return campaign, nil
}

// UpdateCampaignStatus 修改指定活动的生命周期状态（如启动、结束或取消）。
func (m *MarketingManager) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	campaign, err := m.repo.GetCampaign(ctx, id)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get campaign for status update", "campaign_id", id, "error", err)
		return err
	}

	switch status {
	case domain.CampaignStatusOngoing:
		campaign.Start()
	case domain.CampaignStatusEnded:
		campaign.End()
	case domain.CampaignStatusCanceled:
		campaign.Cancel()
	}

	if err := m.repo.SaveCampaign(ctx, campaign); err != nil {
		m.logger.ErrorContext(ctx, "failed to save campaign status", "campaign_id", id, "new_status", status, "error", err)
		return err
	}

	m.logger.InfoContext(ctx, "campaign status updated", "campaign_id", id, "new_status", status)
	return nil
}

// RecordParticipation 记录用户参与营销活动的行为，包含布隆过滤器预检、预算扣减及参与记录落库。
func (m *MarketingManager) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	campaign, err := m.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get campaign for participation", "campaign_id", campaignID, "error", err)
		return err
	}

	if !campaign.IsActive() {
		return domain.ErrCampaignEnded
	}

	// 性能优化：使用布隆过滤器在 O(1) 时间内初步阻断重复参与请求
	// Key: campaignID:userID
	filterKey := fmt.Appendf(nil, "%d:%d", campaignID, userID)
	if m.userFilter.Contains(filterKey) {
		m.logger.DebugContext(ctx, "user might have already participated (bloom filter hit)", "user_id", userID, "campaign_id", campaignID)
	} else {
		// 布隆过滤器说不存在，那就一定不存在
		m.userFilter.Add(filterKey)
		campaign.IncrementReachedUsers() // 仅当是新用户时才增加触达人数
	}

	if campaign.RemainingBudget() < discount {
		return domain.ErrCampaignEnded
	}

	participation := domain.NewCampaignParticipation(campaignID, userID, orderID, discount)
	if err := m.repo.SaveParticipation(ctx, participation); err != nil {
		m.logger.ErrorContext(ctx, "failed to save campaign participation", "user_id", userID, "campaign_id", campaignID, "error", err)
		return err
	}

	campaign.AddSpent(discount)
	if err := m.repo.SaveCampaign(ctx, campaign); err != nil {
		m.logger.ErrorContext(ctx, "failed to update campaign budget after participation", "campaign_id", campaignID, "error", err)
		return err
	}

	m.logger.InfoContext(ctx, "user participation recorded", "user_id", userID, "campaign_id", campaignID, "discount", discount)
	return nil
}

// CreateBanner 创建一个Banner。
func (m *MarketingManager) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*domain.Banner, error) {
	banner := domain.NewBanner(title, imageURL, linkURL, position, priority, startTime, endTime)
	if err := m.repo.SaveBanner(ctx, banner); err != nil {
		m.logger.ErrorContext(ctx, "failed to create banner", "title", title, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "banner created successfully", "banner_id", banner.ID, "title", title)
	return banner, nil
}

// DeleteBanner 物理删除指定的横幅广告位。
func (m *MarketingManager) DeleteBanner(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteBanner(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete banner", "banner_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "banner deleted", "banner_id", id)
	return nil
}

// ClickBanner 原子递增横幅的点击计数器。
func (m *MarketingManager) ClickBanner(ctx context.Context, id uint64) error {
	banner, err := m.repo.GetBanner(ctx, id)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get banner for click", "banner_id", id, "error", err)
		return err
	}
	banner.IncrementClick()
	if err := m.repo.SaveBanner(ctx, banner); err != nil {
		m.logger.ErrorContext(ctx, "failed to save banner click count", "banner_id", id, "error", err)
		return err
	}
	m.logger.DebugContext(ctx, "banner click recorded", "banner_id", id)
	return nil
}
