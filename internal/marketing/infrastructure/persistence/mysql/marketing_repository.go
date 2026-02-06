package mysql

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

// NewMarketingRepository 创建并返回一个新的 MarketingRepository 实例。
func NewMarketingRepository(db *gorm.DB) domain.MarketingRepository {
	return &marketingRepository{db: db}
}

func (r *marketingRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *marketingRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *marketingRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *marketingRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- Campaign ---

func (r *marketingRepository) SaveCampaign(ctx context.Context, campaign *domain.Campaign) error {
	return r.saveCampaignWithTx(ctx, r.db, campaign)
}

func (r *marketingRepository) SaveCampaignInTx(ctx context.Context, tx any, campaign *domain.Campaign) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveCampaignWithTx(ctx, gormTx, campaign)
}

func (r *marketingRepository) GetCampaign(ctx context.Context, id uint64) (*domain.Campaign, error) {
	var model CampaignModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCampaignNotFound
		}
		return nil, err
	}
	return toCampaign(&model), nil
}

func (r *marketingRepository) ListCampaigns(ctx context.Context, status domain.CampaignStatus, offset, limit int) ([]*domain.Campaign, int64, error) {
	var list []*CampaignModel
	var total int64

	db := r.db.WithContext(ctx).Model(&CampaignModel{})
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Campaign, len(list))
	for i, model := range list {
		items[i] = toCampaign(model)
	}

	return items, total, nil
}

// --- Participation ---

func (r *marketingRepository) SaveParticipation(ctx context.Context, participation *domain.CampaignParticipation) error {
	return r.saveParticipationWithTx(ctx, r.db, participation)
}

func (r *marketingRepository) SaveParticipationInTx(ctx context.Context, tx any, participation *domain.CampaignParticipation) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveParticipationWithTx(ctx, gormTx, participation)
}

func (r *marketingRepository) ListParticipations(ctx context.Context, campaignID uint64, offset, limit int) ([]*domain.CampaignParticipation, int64, error) {
	var list []*CampaignParticipationModel
	var total int64

	db := r.db.WithContext(ctx).Model(&CampaignParticipationModel{})
	if campaignID != 0 {
		db = db.Where("campaign_id = ?", campaignID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.CampaignParticipation, len(list))
	for i, model := range list {
		items[i] = toParticipation(model)
	}

	return items, total, nil
}

// --- Banner ---

func (r *marketingRepository) SaveBanner(ctx context.Context, banner *domain.Banner) error {
	return r.saveBannerWithTx(ctx, r.db, banner)
}

func (r *marketingRepository) SaveBannerInTx(ctx context.Context, tx any, banner *domain.Banner) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveBannerWithTx(ctx, gormTx, banner)
}

func (r *marketingRepository) GetBanner(ctx context.Context, id uint64) (*domain.Banner, error) {
	var model BannerModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBannerNotFound
		}
		return nil, err
	}
	return toBanner(&model), nil
}

func (r *marketingRepository) ListBanners(ctx context.Context, position string, activeOnly bool) ([]*domain.Banner, error) {
	var list []*BannerModel
	db := r.db.WithContext(ctx).Model(&BannerModel{})

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

	items := make([]*domain.Banner, len(list))
	for i, model := range list {
		items[i] = toBanner(model)
	}
	return items, nil
}

func (r *marketingRepository) DeleteBanner(ctx context.Context, id uint64) error {
	return r.deleteBannerWithTx(ctx, r.db, id)
}

func (r *marketingRepository) DeleteBannerInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteBannerWithTx(ctx, gormTx, id)
}

// --- User Tags ---

func (r *marketingRepository) GetUserIDsByTag(ctx context.Context, tagName string) ([]uint32, error) {
	var userIDs []uint32
	err := r.db.WithContext(ctx).Table("user_tags").
		Where("tag_name = ?", tagName).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *marketingRepository) saveCampaignWithTx(ctx context.Context, tx *gorm.DB, campaign *domain.Campaign) error {
	if campaign == nil {
		return nil
	}
	model := toCampaignModel(campaign)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	campaign.ID = uint64(model.ID)
	campaign.CreatedAt = model.CreatedAt
	campaign.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *marketingRepository) saveParticipationWithTx(ctx context.Context, tx *gorm.DB, participation *domain.CampaignParticipation) error {
	if participation == nil {
		return nil
	}
	model := toParticipationModel(participation)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	participation.ID = uint64(model.ID)
	participation.CreatedAt = model.CreatedAt
	participation.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *marketingRepository) saveBannerWithTx(ctx context.Context, tx *gorm.DB, banner *domain.Banner) error {
	if banner == nil {
		return nil
	}
	model := toBannerModel(banner)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	banner.ID = uint64(model.ID)
	banner.CreatedAt = model.CreatedAt
	banner.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *marketingRepository) deleteBannerWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&BannerModel{}, id).Error
}
