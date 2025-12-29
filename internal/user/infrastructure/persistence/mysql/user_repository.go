package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"gorm.io/gorm"
)

// UserRepository 实现 domain.UserRepository 接口
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Save 保存用户 (Create or Update)
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	// GORM Save 会根据 ID 是否存在及非零值判断是 Insert 还是 Update
	// 但通常 Insert 用 Create 更明确
	if user.ID == 0 {
		return r.db.WithContext(ctx).Create(user).Error
	}
	return r.db.WithContext(ctx).Save(user).Error
}

// FindByID 根据 ID 查找用户
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Addresses").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据手机号查找用户
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

// List 列出用户
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AddressRepository 实现 domain.AddressRepository 接口
type AddressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建 AddressRepository 实例
func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// Save 保存地址
func (r *AddressRepository) Save(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

// FindByID 根据 ID 查找地址
func (r *AddressRepository) FindByID(ctx context.Context, id uint) (*domain.Address, error) {
	var address domain.Address
	if err := r.db.WithContext(ctx).First(&address, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// FindDefaultByUserID 查找用户默认地址
func (r *AddressRepository) FindDefaultByUserID(ctx context.Context, userID uint) (*domain.Address, error) {
	var address domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// FindByUserID 查找用户所有地址
func (r *AddressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var addresses []*domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

// Update 更新地址
func (r *AddressRepository) Update(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

// Delete 删除地址
func (r *AddressRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Address{}, id).Error
}

// SetDefault 设置默认地址
func (r *AddressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Reset all addresses for user to not default
		if err := tx.Model(&domain.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}

		// 2. Set target address to default
		if err := tx.Model(&domain.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}
