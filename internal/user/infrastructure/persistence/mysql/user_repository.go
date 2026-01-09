// Package mysql 提供了用户及地址仓储接口的 MySQL GORM 实现。
package mysql

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"gorm.io/gorm"
)

// UserRepository 实现了 domain.UserRepository 接口，提供基于 MySQL 的用户数据持久化能力。
type UserRepository struct {
	db *gorm.DB // GORM 数据库连接实例
}

// NewUserRepository 构造一个新的用户仓储实例。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Save 将用户实体状态同步至数据库（支持创建与全量更新）。
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	if user.ID == 0 {
		return r.db.WithContext(ctx).Create(user).Error
	}
	return r.db.WithContext(ctx).Save(user).Error
}

// FindByID 根据主键 ID 获取用户实体，并预加载关联的地址列表。
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

// FindByUsername 根据唯一用户名检索用户。
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

// FindByEmail 根据唯一邮箱地址检索用户。
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

// FindByPhone 根据手机号检索用户。
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

// Update 更新用户实体的所有可变字段。
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 根据 ID 物理删除用户记录。
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

// List 分页列出系统所有用户。
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

// AddressRepository 实现了 domain.AddressRepository 接口，管理收货地址的持久化。
type AddressRepository struct {
	db *gorm.DB // GORM 数据库连接实例
}

// NewAddressRepository 构造一个新的地址仓储实例。
func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// Save 持久化单条地址记录。
func (r *AddressRepository) Save(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

// FindByID 根据主键获取地址。
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

// FindDefaultByUserID 获取用户当前生效的默认地址。
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

// FindByUserID 获取用户名下的所有收货地址。
func (r *AddressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var addresses []*domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

// Update 更新地址实体的细节。
func (r *AddressRepository) Update(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

// Delete 物理删除特定的地址。
func (r *AddressRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Address{}, id).Error
}

// SetDefault 切换用户的默认地址。
// 架构设计：在数据库事务中执行“全量取消”与“单条设定”，确保原子性。
func (r *AddressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 将该用户下的所有地址重置为非默认
		if err := tx.Model(&domain.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}

		// 2. 将选定的目标地址设为默认
		if err := tx.Model(&domain.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to set default address in db", "user_id", userID, "address_id", addressID, "error", err)
		return err
	}

	slog.InfoContext(ctx, "default address updated in db", "user_id", userID, "address_id", addressID)
	return nil
}
