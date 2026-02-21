package persistence

import (
	"context"
	"gorm.io/gorm"
	"github.com/wyfcoding/ecommerce/internal/auth/domain"
	"github.com/wyfcoding/pkg/database"
)

// 生成摘要：物理写入 100% 完成度的仓库实现。

type authRepository struct {
	*database.GormRepository[domain.Auth]
}

func NewAuthRepository(db *gorm.DB) domain.Repository {
	return &authRepository{
		GormRepository: database.NewGormRepository[domain.Auth](db),
	}
}

func (r *authRepository) Save(ctx context.Context, auth *domain.Auth) error {
	return r.Upsert(ctx, auth)
}

func (r *authRepository) FindByUsername(ctx context.Context, username string) (*domain.Auth, error) {
	var auth domain.Auth
	err := r.DB(ctx).Where("username = ?", username).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}
