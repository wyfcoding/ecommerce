package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/user/domain"

	"gorm.io/gorm"
)

// AddressRepository implements domain.AddressRepository using GORM.
type AddressRepository struct {
	db *gorm.DB
}

// NewAddressRepository creates a new AddressRepository.
func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// Save saves a new address.
func (r *AddressRepository) Save(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

// FindByID finds an address by ID.
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

// FindDefaultByUserID finds the default address for a user.
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

// FindByUserID finds all addresses for a user.
func (r *AddressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var addresses []*domain.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

// Update updates an address.
func (r *AddressRepository) Update(ctx context.Context, address *domain.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

// Delete deletes an address by ID.
func (r *AddressRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Address{}, id).Error
}

// SetDefault sets the default address for a user.
func (r *AddressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reset all addresses for this user to non-default
		if err := tx.Model(&domain.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}

		// Set the specified address to default
		if err := tx.Model(&domain.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}
