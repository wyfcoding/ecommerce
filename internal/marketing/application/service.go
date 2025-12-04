package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"     // 导入营销领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/repository" // 导入营销领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// MarketingService 结构体定义了营销活动相关的应用服务。
// 它协调领域层和基础设施层，处理营销活动的创建、管理、用户参与记录以及Banner的管理等业务逻辑。
type MarketingService struct {
	repo   repository.MarketingRepository // 依赖MarketingRepository接口，用于数据持久化操作。
	logger *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewMarketingService 创建并返回一个新的 MarketingService 实例。
func NewMarketingService(repo repository.MarketingRepository, logger *slog.Logger) *MarketingService {
	return &MarketingService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCampaign 创建一个新的营销活动。
// ctx: 上下文。
// name: 活动名称。
// campaignType: 活动类型。
// description: 活动描述。
// startTime, endTime: 活动的开始和结束时间。
// budget: 活动预算。
// rules: 活动规则（例如，JSON格式）。
// 返回创建成功的Campaign实体和可能发生的错误。
func (s *MarketingService) CreateCampaign(ctx context.Context, name string, campaignType entity.CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]interface{}) (*entity.Campaign, error) {
	campaign := entity.NewCampaign(name, campaignType, description, startTime, endTime, budget, rules) // 创建Campaign实体。
	// 通过仓储接口保存营销活动。
	if err := s.repo.SaveCampaign(ctx, campaign); err != nil {
		s.logger.ErrorContext(ctx, "failed to create campaign", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "campaign created successfully", "campaign_id", campaign.ID, "name", name)
	return campaign, nil
}

// GetCampaign 获取指定ID的营销活动详情。
// ctx: 上下文。
// id: 营销活动ID。
// 返回Campaign实体和可能发生的错误。
func (s *MarketingService) GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error) {
	return s.repo.GetCampaign(ctx, id)
}

// UpdateCampaignStatus 更新指定ID的营销活动状态。
// ctx: 上下文。
// id: 营销活动ID。
// status: 新的状态。
// 返回可能发生的错误。
func (s *MarketingService) UpdateCampaignStatus(ctx context.Context, id uint64, status entity.CampaignStatus) error {
	// 获取营销活动实体。
	campaign, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return err
	}

	// 根据新的状态调用实体的方法进行状态转换。
	switch status {
	case entity.CampaignStatusOngoing:
		campaign.Start()
	case entity.CampaignStatusEnded:
		campaign.End()
	case entity.CampaignStatusCanceled:
		campaign.Cancel()
	}

	// 通过仓储接口保存更新后的营销活动。
	return s.repo.SaveCampaign(ctx, campaign)
}

// ListCampaigns 获取营销活动列表。
// ctx: 上下文。
// status: 筛选活动的活跃状态。
// page, pageSize: 分页参数。
// 返回营销活动列表、总数和可能发生的错误。
func (s *MarketingService) ListCampaigns(ctx context.Context, status entity.CampaignStatus, page, pageSize int) ([]*entity.Campaign, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListCampaigns(ctx, status, offset, pageSize)
}

// RecordParticipation 记录用户参与营销活动。
// ctx: 上下文。
// campaignID: 营销活动ID。
// userID: 参与用户ID。
// orderID: 关联订单ID。
// discount: 用户获得的优惠金额。
// 返回可能发生的错误。
func (s *MarketingService) RecordParticipation(ctx context.Context, campaignID, userID, orderID, discount uint64) error {
	// 获取营销活动实体。
	campaign, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		return err
	}

	// 检查活动是否仍在进行中。
	if !campaign.IsActive() {
		return entity.ErrCampaignEnded // 活动已结束。
	}

	// 检查活动预算是否充足。
	if campaign.RemainingBudget() < discount {
		return entity.ErrCampaignEnded // 预算不足（或返回更具体的错误）。
	}

	// 创建参与记录实体。
	participation := entity.NewCampaignParticipation(campaignID, userID, orderID, discount)
	// 通过仓储接口保存参与记录。
	if err := s.repo.SaveParticipation(ctx, participation); err != nil {
		return err
	}

	// 更新营销活动相关统计数据。
	campaign.AddSpent(discount)      // 增加已花费金额。
	campaign.IncrementReachedUsers() // 增加参与用户数。
	return s.repo.SaveCampaign(ctx, campaign)
}

// CreateBanner 创建一个Banner。
// ctx: 上下文。
// title: Banner标题。
// imageURL, linkURL: 图片URL和点击跳转链接。
// position: Banner展示位置。
// priority: 优先级。
// startTime, endTime: Banner的展示时间范围。
// 返回创建成功的Banner实体和可能发生的错误。
func (s *MarketingService) CreateBanner(ctx context.Context, title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) (*entity.Banner, error) {
	banner := entity.NewBanner(title, imageURL, linkURL, position, priority, startTime, endTime) // 创建Banner实体。
	// 通过仓储接口保存Banner。
	if err := s.repo.SaveBanner(ctx, banner); err != nil {
		s.logger.ErrorContext(ctx, "failed to create banner", "title", title, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "banner created successfully", "banner_id", banner.ID, "title", title)
	return banner, nil
}

// ListBanners 获取Banner列表。
// ctx: 上下文。
// position: 筛选Banner的展示位置。
// activeOnly: 布尔值，如果为true，则只列出当前活跃的Banner。
// 返回Banner列表和可能发生的错误。
func (s *MarketingService) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error) {
	return s.repo.ListBanners(ctx, position, activeOnly)
}

// DeleteBanner 删除指定ID的Banner。
// ctx: 上下文。
// id: Banner ID。
// 返回可能发生的错误。
func (s *MarketingService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.repo.DeleteBanner(ctx, id)
}

// ClickBanner 记录Banner点击事件。
// ctx: 上下文。
// id: Banner ID。
// 返回可能发生的错误。
func (s *MarketingService) ClickBanner(ctx context.Context, id uint64) error {
	// 获取Banner实体。
	banner, err := s.repo.GetBanner(ctx, id)
	if err != nil {
		return err
	}
	// 调用实体方法增加点击次数。
	banner.IncrementClick()
	// 保存更新后的Banner。
	return s.repo.SaveBanner(ctx, banner)
}
