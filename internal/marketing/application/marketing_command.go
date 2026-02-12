package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	couponv1 "github.com/wyfcoding/ecommerce/go-api/coupon/v1"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
	"github.com/wyfcoding/pkg/messagequeue"
)

// MarketingCommandService 处理营销的写操作。
type MarketingCommandService struct {
	repo           domain.MarketingRepository
	publisher      messagequeue.EventPublisher
	logger         *slog.Logger
	userFilter     *algorithm.BloomFilter[algorithm.ByteHash]
	segmentService *UserSegmentService
}

// NewMarketingCommandService 构造函数。
func NewMarketingCommandService(
	repo domain.MarketingRepository,
	publisher messagequeue.EventPublisher,
	couponCli couponv1.CouponServiceClient,
	logger *slog.Logger,
) (*MarketingCommandService, error) {
	filter, err := algorithm.NewBloomFilter[algorithm.ByteHash](1000000, 0.01)
	if err != nil {
		return nil, err
	}

	return &MarketingCommandService{
		repo:           repo,
		publisher:      publisher,
		logger:         logger,
		userFilter:     filter,
		segmentService: NewUserSegmentService(repo, couponCli, logger),
	}, nil
}

// CreateCampaign 创建一个新的营销活动。
func (s *MarketingCommandService) CreateCampaign(ctx context.Context, name string, campaignType domain.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]any) (*domain.Campaign, error) {
	campaign := domain.NewCampaign(name, campaignType, description, startTime, endTime, budget, rules)

	err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveCampaignInTx(ctx, tx, campaign); err != nil {
			return err
		}

		event := &domain.CampaignCreatedEvent{
			CampaignID: campaign.ID,
			Name:       campaign.Name,
			Type:       campaign.CampaignType,
			StartTime:  campaign.StartTime,
			EndTime:    campaign.EndTime,
			Timestamp:  time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.CampaignCreatedEventType, fmt.Sprintf("%d", campaign.ID), event)
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create campaign", "name", name, "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "campaign created", "campaign_id", campaign.ID, "name", name)
	return campaign, nil
}

// UpdateCampaignStatus 修改营销活动状态。
func (s *MarketingCommandService) UpdateCampaignStatus(ctx context.Context, id uint64, status domain.CampaignStatus) error {
	campaign, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := campaign.Status
	return s.repo.WithTx(ctx, func(tx any) error {
		switch status {
		case domain.CampaignStatusOngoing:
			campaign.Start()
		case domain.CampaignStatusEnded:
			campaign.End()
		case domain.CampaignStatusCanceled:
			campaign.Cancel()
		}

		if err := s.repo.SaveCampaignInTx(ctx, tx, campaign); err != nil {
			return err
		}

		event := &domain.CampaignStatusUpdatedEvent{
			CampaignID: campaign.ID,
			OldStatus:  oldStatus,
			NewStatus:  status,
			Timestamp:  time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.CampaignStatusUpdatedEventType, fmt.Sprintf("%d", id), event)
	})
}

// RecordParticipation 记录用户参与营销活动的行为。
func (s *MarketingCommandService) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	campaign, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		return err
	}

	if !campaign.IsActive() {
		return domain.ErrCampaignEnded
	}

	filterKey := algorithm.ByteHash(fmt.Appendf(nil, "%d:%d", campaignID, userID))
	if s.userFilter.Contains(filterKey) {
		s.logger.DebugContext(ctx, "potential double participation detected", "user_id", userID, "campaign_id", campaignID)
	}

	if campaign.RemainingBudget() < discount {
		return domain.ErrCampaignEnded
	}

	participation := domain.NewCampaignParticipation(campaignID, userID, orderID, discount)

	return s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveParticipationInTx(ctx, tx, participation); err != nil {
			return err
		}

		campaign.AddSpent(discount)
		campaign.IncrementReachedUsers()
		if err := s.repo.SaveCampaignInTx(ctx, tx, campaign); err != nil {
			return err
		}

		// 发布布隆过滤器标记（此处为内存操作，但逻辑上跟事务走）
		s.userFilter.Add(filterKey)

		event := &domain.ParticipationRecordedEvent{
			CampaignID: campaignID,
			UserID:     userID,
			OrderID:    orderID,
			Discount:   discount,
			Timestamp:  time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.ParticipationRecordedEventType, fmt.Sprintf("%d:%d", campaignID, userID), event)
	})
}

// CreateBanner 创建广告位。
func (s *MarketingCommandService) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*domain.Banner, error) {
	banner := domain.NewBanner(title, imageURL, linkURL, position, priority, startTime, endTime)

	err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveBannerInTx(ctx, tx, banner); err != nil {
			return err
		}

		event := &domain.BannerCreatedEvent{
			BannerID:  banner.ID,
			Title:     banner.Title,
			Position:  banner.Position,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BannerCreatedEventType, fmt.Sprintf("%d", banner.ID), event)
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create banner", "title", title, "error", err)
		return nil, err
	}

	return banner, nil
}

// DeleteBanner 删除广告位。
func (s *MarketingCommandService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.DeleteBannerInTx(ctx, tx, id); err != nil {
			return err
		}
		event := &domain.BannerDeletedEvent{
			BannerID:  id,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BannerDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// ClickBanner 记录广告位点击。
func (s *MarketingCommandService) ClickBanner(ctx context.Context, id uint64) error {
	banner, err := s.repo.GetBanner(ctx, id)
	if err != nil {
		return err
	}
	banner.IncrementClick()
	return s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveBannerInTx(ctx, tx, banner); err != nil {
			return err
		}
		event := &domain.BannerClickedEvent{
			BannerID:  id,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BannerClickedEventType, fmt.Sprintf("%d", id), event)
	})
}

// DistributeTargetedCoupons 定向优惠券分发。
func (s *MarketingCommandService) DistributeTargetedCoupons(ctx context.Context, couponID uint64, targetTags []string) error {
	return s.segmentService.DistributeCouponsToSegment(ctx, couponID, targetTags)
}
