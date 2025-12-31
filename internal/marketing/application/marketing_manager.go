package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// MarketingManager 处理营销的写操作。
type MarketingManager struct {
	repo       domain.MarketingRepository
	logger     *slog.Logger
	userFilter *algorithm.BloomFilter
}

// NewMarketingManager creates a new MarketingManager instance.
func NewMarketingManager(repo domain.MarketingRepository, logger *slog.Logger) *MarketingManager {
	return &MarketingManager{
		repo:       repo,
		logger:     logger,
		userFilter: algorithm.NewBloomFilter(1000000, 0.01), // 100万用户容量，1%误判率
	}
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

// UpdateCampaignStatus 更新指定ID的营销活动状态。
func (m *MarketingManager) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	campaign, err := m.repo.GetCampaign(ctx, id)
	if err != nil {
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

	return m.repo.SaveCampaign(ctx, campaign)
}

// RecordParticipation 记录用户参与营销活动。
func (m *MarketingManager) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	campaign, err := m.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		return err
	}

	if !campaign.IsActive() {
		return domain.ErrCampaignEnded
	}

	// 使用布隆过滤器快速检查用户是否已参与 (针对 "每人限一次" 的规则优化)
	// Key: campaignID:userID
	filterKey := []byte(fmt.Sprintf("%d:%d", campaignID, userID))
	if m.userFilter.Contains(filterKey) {
		// 布隆过滤器说存在，可能是误判，也可能真存在。
		// 这里我们做一个快速拦截，或者也可以继续去DB确认。
		// 为了演示算法价值，假设这是一个高并发场景，我们倾向于相信布隆过滤器来挡掉绝大多数重复请求。
		// 如果需要绝对精确，这里应该 fallback 到 DB 查询。
		// 此处仅做日志记录，不强制阻断，依靠 DB 唯一索引兜底 (如果 DB 有的话)。
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
		return err
	}

	campaign.AddSpent(discount)
	// campaign.IncrementReachedUsers() // Moved to bloom filter logic above for 'unique' count
	return m.repo.SaveCampaign(ctx, campaign)
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

// DeleteBanner 删除指定ID的Banner。
func (m *MarketingManager) DeleteBanner(ctx context.Context, id uint64) error {
	return m.repo.DeleteBanner(ctx, id)
}

// ClickBanner 记录Banner点击事件。
func (m *MarketingManager) ClickBanner(ctx context.Context, id uint64) error {
	banner, err := m.repo.GetBanner(ctx, id)
	if err != nil {
		return err
	}
	banner.IncrementClick()
	return m.repo.SaveBanner(ctx, banner)
}
