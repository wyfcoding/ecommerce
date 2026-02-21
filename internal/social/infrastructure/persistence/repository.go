package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/social/domain"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// 生成摘要：实现 Social 服务的 GORM 仓储。
// 关键改动：继承泛型仓储，实现业务特定的查询逻辑。

type socialRepository struct {
	*database.GormRepository[domain.SocialRelation]
}

func NewSocialRepository(db *gorm.DB) domain.Repository {
	return &socialRepository{
		GormRepository: database.NewGormRepository[domain.SocialRelation](db),
	}
}

func (r *socialRepository) Save(ctx context.Context, rel *domain.SocialRelation) error {
	return r.Upsert(ctx, rel)
}

func (r *socialRepository) FindByUserID(ctx context.Context, userID string) (*domain.SocialRelation, error) {
	var rel domain.SocialRelation
	err := r.DB(ctx).Where("user_id = ?", userID).First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *socialRepository) FindChildrenByParentID(ctx context.Context, parentID string) ([]*domain.SocialRelation, error) {
	var list []*domain.SocialRelation
	err := r.DB(ctx).Where("parent_id = ?", parentID).Find(&list).Error
	return list, err
}
