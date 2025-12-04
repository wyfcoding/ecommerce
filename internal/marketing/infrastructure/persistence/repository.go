package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"     // 导入营销模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/repository" // 导入营销模块的领域仓储接口。
	"time"                                                                // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type marketingRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewMarketingRepository 创建并返回一个新的 marketingRepository 实例。
// db: GORM数据库连接实例。
func NewMarketingRepository(db *gorm.DB) repository.MarketingRepository {
	return &marketingRepository{db: db}
}

// --- 营销活动 (Campaign methods) ---

// SaveCampaign 将营销活动实体保存到数据库。
// 如果活动已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *marketingRepository) SaveCampaign(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Save(campaign).Error
}

// GetCampaign 根据ID从数据库获取营销活动记录。
// 如果记录未找到，则返回 entity.ErrCampaignNotFound 错误。
func (r *marketingRepository) GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.WithContext(ctx).First(&campaign, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCampaignNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &campaign, nil
}

// ListCampaigns 从数据库列出所有营销活动记录，支持通过状态过滤和分页。
func (r *marketingRepository) ListCampaigns(ctx context.Context, status entity.CampaignStatus, offset, limit int) ([]*entity.Campaign, int64, error) {
	var list []*entity.Campaign
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Campaign{})
	// 根据状态过滤活动。
	// 注意：当前查询假定 status 参数总是有效且需要过滤，不为0的状态值将被用于过滤。
	db = db.Where("status = ?", status)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 参与记录 (Participation methods) ---

// SaveParticipation 将活动参与记录实体保存到数据库。
func (r *marketingRepository) SaveParticipation(ctx context.Context, participation *entity.CampaignParticipation) error {
	return r.db.WithContext(ctx).Save(participation).Error
}

// ListParticipations 从数据库列出指定营销活动ID的所有参与记录，支持分页。
func (r *marketingRepository) ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*entity.CampaignParticipation, int64, error) {
	var list []*entity.CampaignParticipation
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CampaignParticipation{}).Where("campaign_id = ?", campaignID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 广告横幅 (Banner methods) ---

// SaveBanner 将广告横幅实体保存到数据库。
func (r *marketingRepository) SaveBanner(ctx context.Context, banner *entity.Banner) error {
	return r.db.WithContext(ctx).Save(banner).Error
}

// GetBanner 根据ID从数据库获取广告横幅记录。
func (r *marketingRepository) GetBanner(ctx context.Context, id uint64) (*entity.Banner, error) {
	var banner entity.Banner
	if err := r.db.WithContext(ctx).First(&banner, id).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

// ListBanners 从数据库列出所有广告横幅记录，支持通过位置和活跃状态过滤。
func (r *marketingRepository) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error) {
	var list []*entity.Banner
	db := r.db.WithContext(ctx).Model(&entity.Banner{})

	if position != "" { // 如果提供了位置，则按位置过滤。
		db = db.Where("position = ?", position)
	}

	if activeOnly { // 如果activeOnly为true，则只列出当前活跃的Banner。
		now := time.Now()
		db = db.Where("enabled = ? AND start_time <= ? AND end_time >= ?", true, now, now)
	}

	// 应用排序。
	if err := db.Order("priority desc, created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteBanner 根据ID从数据库删除广告横幅记录。
func (r *marketingRepository) DeleteBanner(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Banner{}, id).Error
}
