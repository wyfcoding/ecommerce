package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/repository"

	"log/slog"
)

type MarketingService struct {
	repo   repository.MarketingRepository
	logger *slog.Logger
}

func NewMarketingService(repo repository.MarketingRepository, logger *slog.Logger) *MarketingService {
	return &MarketingService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCampaign 创建营销活动
func (s *MarketingService) CreateCampaign(ctx context.Context, name string, campaignType entity.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]interface{}) (*entity.Campaign, error) {
	campaign := entity.NewCampaign(name, campaignType, description, startTime, endTime, budget, rules)
	if err := s.repo.SaveCampaign(ctx, campaign); err != nil {
		s.logger.ErrorContext(ctx, "failed to create campaign", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "campaign created successfully", "campaign_id", campaign.ID, "name", name)
	return campaign, nil
}

// GetCampaign 获取营销活动
func (s *MarketingService) GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error) {
	return s.repo.GetCampaign(ctx, id)
}

// UpdateCampaignStatus 更新活动状态
func (s *MarketingService) UpdateCampaignStatus(ctx context.Context, id uint64, status entity.CampaignStatus) error {
	campaign, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return err
	}

	switch status {
	case entity.CampaignStatusOngoing:
		campaign.Start()
	case entity.CampaignStatusEnded:
		campaign.End()
	case entity.CampaignStatusCanceled:
		campaign.Cancel()
	}

	return s.repo.SaveCampaign(ctx, campaign)
}

// ListCampaigns 获取活动列表
func (s *MarketingService) ListCampaigns(ctx context.Context, status entity.CampaignStatus, page, pageSize int) ([]*entity.Campaign, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListCampaigns(ctx, status, offset, pageSize)
}

// RecordParticipation 记录参与
func (s *MarketingService) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	campaign, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		return err
	}

	if !campaign.IsActive() {
		return entity.ErrCampaignEnded
	}

	if campaign.RemainingBudget() < discount {
		return entity.ErrCampaignEnded // Or insufficient budget error
	}

	participation := entity.NewCampaignParticipation(campaignID, userID, orderID, discount)
	if err := s.repo.SaveParticipation(ctx, participation); err != nil {
		return err
	}

	campaign.AddSpent(discount)
	campaign.IncrementReachedUsers()
	return s.repo.SaveCampaign(ctx, campaign)
}

// CreateBanner 创建Banner
func (s *MarketingService) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*entity.Banner, error) {
	banner := entity.NewBanner(title, imageURL, linkURL, position, priority, startTime, endTime)
	if err := s.repo.SaveBanner(ctx, banner); err != nil {
		s.logger.ErrorContext(ctx, "failed to create banner", "title", title, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "banner created successfully", "banner_id", banner.ID, "title", title)
	return banner, nil
}

// ListBanners 获取Banner列表
func (s *MarketingService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error) {
	return s.repo.ListBanners(ctx, position, activeOnly)
}

// DeleteBanner 删除Banner
func (s *MarketingService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.repo.DeleteBanner(ctx, id)
}

// ClickBanner 点击Banner
func (s *MarketingService) ClickBanner(ctx context.Context, id uint64) error {
	banner, err := s.repo.GetBanner(ctx, id)
	if err != nil {
		return err
	}
	banner.IncrementClick()
	return s.repo.SaveBanner(ctx, banner)
}
