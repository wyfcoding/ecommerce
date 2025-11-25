package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type marketingRepository struct {
	db *gorm.DB
}

func NewMarketingRepository(db *gorm.DB) repository.MarketingRepository {
	return &marketingRepository{db: db}
}

// 营销活动
func (r *marketingRepository) SaveCampaign(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Save(campaign).Error
}

func (r *marketingRepository) GetCampaign(ctx context.Context, id uint64) (*entity.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.WithContext(ctx).First(&campaign, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCampaignNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

func (r *marketingRepository) ListCampaigns(ctx context.Context, status entity.CampaignStatus, offset, limit int) ([]*entity.Campaign, int64, error) {
	var list []*entity.Campaign
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Campaign{})
	// If status is -1 (or some indicator), we might list all. Assuming valid enum >= 0
	// Here we just filter if passed. You might want a more flexible filter.
	// For now, let's assume we always filter by status if it's a specific query,
	// but the interface signature implies a specific status.
	// Let's assume the caller handles logic or we add a "All" status.
	// For simplicity, we filter by status.
	db = db.Where("status = ?", status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 参与记录
func (r *marketingRepository) SaveParticipation(ctx context.Context, participation *entity.CampaignParticipation) error {
	return r.db.WithContext(ctx).Save(participation).Error
}

func (r *marketingRepository) ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*entity.CampaignParticipation, int64, error) {
	var list []*entity.CampaignParticipation
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CampaignParticipation{}).Where("campaign_id = ?", campaignID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 广告横幅
func (r *marketingRepository) SaveBanner(ctx context.Context, banner *entity.Banner) error {
	return r.db.WithContext(ctx).Save(banner).Error
}

func (r *marketingRepository) GetBanner(ctx context.Context, id uint64) (*entity.Banner, error) {
	var banner entity.Banner
	if err := r.db.WithContext(ctx).First(&banner, id).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

func (r *marketingRepository) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*entity.Banner, error) {
	var list []*entity.Banner
	db := r.db.WithContext(ctx).Model(&entity.Banner{})

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
	return r.db.WithContext(ctx).Delete(&entity.Banner{}, id).Error
}
