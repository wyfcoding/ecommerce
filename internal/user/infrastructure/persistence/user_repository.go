package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Addresses").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Updates(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.WithContext(ctx).Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建地址仓储实例
func NewAddressRepository(db *gorm.DB) domain.AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Save(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

func (r *addressRepository) FindByID(ctx context.Context, id uint) (*domain.Address, error) {
	var address domain.Address
	if err := r.db.WithContext(ctx).First(&address, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

func (r *addressRepository) FindDefaultByUserID(ctx context.Context, userID uint) (*domain.Address, error) {
	var address domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

func (r *addressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var addresses []*domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *addressRepository) Update(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Save(address).Error // Save updates if ID exists
}

func (r *addressRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Address{}, id).Error
}

func (r *addressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Reset all to false
		if err := tx.Model(&domain.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		if err := tx.Model(&domain.Address{}).Where("id = ?", addressID).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}
