package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"

	"log/slog"
)

// MarketingManager 处理营销的写操作。
type MarketingManager struct {
	repo   domain.MarketingRepository
	logger *slog.Logger
}

// NewMarketingManager creates a new MarketingManager instance.
func NewMarketingManager(repo domain.MarketingRepository, logger *slog.Logger) *MarketingManager {
	return &MarketingManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateCampaign 创建一个新的营销活动。
func (m *MarketingManager) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]interface{}) (*domain.Campaign, error) {
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

	if campaign.RemainingBudget() < discount {
		return domain.ErrCampaignEnded
	}

	participation := domain.NewCampaignParticipation(campaignID, userID, orderID, discount)
	if err := m.repo.SaveParticipation(ctx, participation); err != nil {
		return err
	}

	campaign.AddSpent(discount)
	campaign.IncrementReachedUsers()
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
