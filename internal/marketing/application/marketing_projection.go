// 生成摘要：新增营销读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingProjectionService 负责将营销事件投影到读模型。
type MarketingProjectionService struct {
	repo               domain.MarketingRepository
	campaignReadRepo   domain.CampaignReadRepository
	bannerReadRepo     domain.BannerReadRepository
	campaignSearchRepo domain.CampaignSearchRepository
	logger             *slog.Logger
}

// NewMarketingProjectionService 创建营销投影服务。
func NewMarketingProjectionService(
	repo domain.MarketingRepository,
	campaignReadRepo domain.CampaignReadRepository,
	bannerReadRepo domain.BannerReadRepository,
	campaignSearchRepo domain.CampaignSearchRepository,
	logger *slog.Logger,
) *MarketingProjectionService {
	return &MarketingProjectionService{
		repo:               repo,
		campaignReadRepo:   campaignReadRepo,
		bannerReadRepo:     bannerReadRepo,
		campaignSearchRepo: campaignSearchRepo,
		logger:             logger,
	}
}

func (s *MarketingProjectionService) OnCampaignCreated(ctx context.Context, event *domain.CampaignCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCampaign(ctx, event.CampaignID)
}

func (s *MarketingProjectionService) OnCampaignStatusUpdated(ctx context.Context, event *domain.CampaignStatusUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCampaign(ctx, event.CampaignID)
}

func (s *MarketingProjectionService) OnParticipationRecorded(ctx context.Context, event *domain.ParticipationRecordedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCampaign(ctx, event.CampaignID)
}

func (s *MarketingProjectionService) OnBannerCreated(ctx context.Context, event *domain.BannerCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshBanner(ctx, event.BannerID)
}

func (s *MarketingProjectionService) OnBannerClicked(ctx context.Context, event *domain.BannerClickedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshBanner(ctx, event.BannerID)
}

func (s *MarketingProjectionService) OnBannerDeleted(ctx context.Context, event *domain.BannerDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.bannerReadRepo != nil {
		_ = s.bannerReadRepo.Delete(ctx, event.BannerID)
	}
	return nil
}

func (s *MarketingProjectionService) refreshCampaign(ctx context.Context, campaignID uint64) error {
	if campaignID == 0 {
		return nil
	}
	campaign, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load campaign for projection", "campaign_id", campaignID, "error", err)
		return err
	}
	if campaign == nil {
		if s.campaignReadRepo != nil {
			_ = s.campaignReadRepo.Delete(ctx, campaignID)
		}
		if s.campaignSearchRepo != nil {
			_ = s.campaignSearchRepo.Delete(ctx, campaignID)
		}
		return nil
	}
	if s.campaignReadRepo != nil {
		if err := s.campaignReadRepo.Save(ctx, campaign); err != nil {
			s.logger.ErrorContext(ctx, "failed to save campaign read model", "campaign_id", campaignID, "error", err)
			return err
		}
	}
	if s.campaignSearchRepo != nil {
		if err := s.campaignSearchRepo.Index(ctx, campaign); err != nil {
			s.logger.ErrorContext(ctx, "failed to index campaign", "campaign_id", campaignID, "error", err)
			return err
		}
	}
	return nil
}

func (s *MarketingProjectionService) refreshBanner(ctx context.Context, bannerID uint64) error {
	if bannerID == 0 {
		return nil
	}
	banner, err := s.repo.GetBanner(ctx, bannerID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load banner for projection", "banner_id", bannerID, "error", err)
		return err
	}
	if banner == nil {
		if s.bannerReadRepo != nil {
			_ = s.bannerReadRepo.Delete(ctx, bannerID)
		}
		return nil
	}
	if s.bannerReadRepo != nil {
		if err := s.bannerReadRepo.Save(ctx, banner); err != nil {
			s.logger.ErrorContext(ctx, "failed to save banner read model", "banner_id", bannerID, "error", err)
			return err
		}
	}
	return nil
}
