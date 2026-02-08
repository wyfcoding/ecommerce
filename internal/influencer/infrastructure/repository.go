package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/influencer/domain"
	"gorm.io/gorm"
)

type GormInfluencerRepository struct {
	db *gorm.DB
}

func NewGormInfluencerRepository(db *gorm.DB) *GormInfluencerRepository {
	return &GormInfluencerRepository{db: db}
}

func (r *GormInfluencerRepository) SaveInfluencer(ctx context.Context, i *domain.Influencer) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *GormInfluencerRepository) GetInfluencer(ctx context.Context, id string) (*domain.Influencer, error) {
	var i domain.Influencer
	err := r.db.WithContext(ctx).Where("influencer_id = ?", id).First(&i).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &i, err
}

func (r *GormInfluencerRepository) GetInfluencerByUserID(ctx context.Context, userID string) (*domain.Influencer, error) {
	var i domain.Influencer
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&i).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &i, err
}

func (r *GormInfluencerRepository) SaveCampaign(ctx context.Context, c *domain.Campaign) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *GormInfluencerRepository) ListCampaigns(ctx context.Context, influencerID string) ([]*domain.Campaign, error) {
	var cs []*domain.Campaign
	err := r.db.WithContext(ctx).Where("influencer_id = ?", influencerID).Find(&cs).Error
	return cs, err
}
