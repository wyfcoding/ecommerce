package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"

	"gorm.io/gorm"
)

type marketingRepository struct {
	db *gorm.DB
}

// NewMarketingRepository 创建并返回一个新的 marketingRepository 实例。
func NewMarketingRepository(db *gorm.DB) domain.MarketingRepository {
	return &marketingRepository{db: db}
}

// --- 营销活动 (Campaign methods) ---

func (r *marketingRepository) SaveCampaign(ctx context.Context, campaign *domain.Campaign) error {
	return r.db.WithContext(ctx).Save(campaign).Error
}

func (r *marketingRepository) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	var campaign domain.Campaign
	if err := r.db.WithContext(ctx).First(&campaign, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCampaignNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

func (r *marketingRepository) ListCampaigns(ctx context.Context, status domain.CampaignStatus, offset, limit int) ([]*domain.Campaign, int64, error) {
	var list []*domain.Campaign
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Campaign{})
	db = db.Where("status = ?", status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 参与记录 (Participation methods) ---

func (r *marketingRepository) SaveParticipation(ctx context.Context, participation *domain.CampaignParticipation) error {
	return r.db.WithContext(ctx).Save(participation).Error
}

func (r *marketingRepository) ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*domain.CampaignParticipation, int64, error) {
	var list []*domain.CampaignParticipation
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.CampaignParticipation{}).Where("campaign_id = ?", campaignID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 广告横幅 (Banner methods) ---

func (r *marketingRepository) SaveBanner(ctx context.Context, banner *domain.Banner) error {
	return r.db.WithContext(ctx).Save(banner).Error
}

func (r *marketingRepository) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	var banner domain.Banner
	if err := r.db.WithContext(ctx).First(&banner, id).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

func (r *marketingRepository) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	var list []*domain.Banner
	db := r.db.WithContext(ctx).Model(&domain.Banner{})

	if position != "" {
		db = db.Where("position = ?", position)
	}

	if activeOnly {
		now := time.Now()
		db = db.Where("enabled = ? AND start_time <= ? AND end_time >= ?", true, now, now)
	}

	if err := db.Order("priority desc, created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *marketingRepository) DeleteBanner(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Banner{}, id).Error
}

// GetUserIDsByTag 从数据库获取具有特定标签的所有用户ID。
func (r *marketingRepository) GetUserIDsByTag(ctx context.Context, tagName string) ([]uint32, error) {
	var userIDs []uint32
	// 假设存在 user_tags 表，存储 user_id 和 tag_name 的映射
	err := r.db.WithContext(ctx).Table("user_tags").
		Where("tag_name = ?", tagName).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		return nil, err
	}
	return userIDs, nil
}
