package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	"gorm.io/gorm"
)

type CrossBorderRepository struct {
	db *gorm.DB
}

func NewCrossBorderRepository(db *gorm.DB) domain.CrossBorderRepository {
	return &CrossBorderRepository{db: db}
}

func (r *CrossBorderRepository) SaveDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) error {
	return r.db.WithContext(ctx).Save(decl).Error
}

func (r *CrossBorderRepository) GetDeclaration(ctx context.Context, id string) (*domain.CustomsDeclaration, error) {
	var decl domain.CustomsDeclaration
	if err := r.db.WithContext(ctx).Preload("Items").Where("declaration_id = ?", id).First(&decl).Error; err != nil {
		return nil, err
	}
	return &decl, nil
}

func (r *CrossBorderRepository) GetHSCode(ctx context.Context, code string) (*domain.HSCode, error) {
	var hs domain.HSCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&hs).Error; err != nil {
		return nil, err
	}
	return &hs, nil
}

func (r *CrossBorderRepository) SaveHSCode(ctx context.Context, hs *domain.HSCode) error {
	return r.db.WithContext(ctx).Save(hs).Error
}
