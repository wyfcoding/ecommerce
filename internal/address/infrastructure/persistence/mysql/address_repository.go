// Package mysql 提供 AddressRepository 的 MySQL 实现。
package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/address/domain"
	"gorm.io/gorm"
)

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建一个 MySQL 存储的 AddressRepository 实现
func NewAddressRepository(db *gorm.DB) domain.AddressRepository {
	// 自动迁移表结构
	_ = db.AutoMigrate(&domain.Address{})
	return &addressRepository{db: db}
}

func (r *addressRepository) Save(ctx context.Context, addr *domain.Address) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *addressRepository) Update(ctx context.Context, addr *domain.Address) error {
	return r.db.WithContext(ctx).Save(addr).Error
}

// Delete 会因为模型嵌套了 gorm.Model 而执行软删除
func (r *addressRepository) Delete(ctx context.Context, id string, userID int64) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Address{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepository) FindByID(ctx context.Context, id string) (*domain.Address, error) {
	var addr domain.Address
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&addr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrAddressNotFound
		}
		return nil, err
	}
	return &addr, nil
}

func (r *addressRepository) FindByUserID(ctx context.Context, userID int64) ([]*domain.Address, error) {
	var addrs []*domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addrs).Error; err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *addressRepository) ClearDefaultByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Address{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

func (r *addressRepository) SetDefault(ctx context.Context, id string, userID int64) error {
	res := r.db.WithContext(ctx).
		Model(&domain.Address{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Address{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
