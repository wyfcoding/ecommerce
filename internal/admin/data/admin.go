package data

import (
	"context"
	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/data/model"
	"gorm.io/gorm"
)

type adminRepo struct {
	data *Data
}

// NewAdminRepo creates a new AdminRepo.
func NewAdminRepo(data *Data) biz.AdminRepo {
	return &adminRepo{data: data}
}

// toBizAdminUser converts a data.AdminUser to a biz.AdminUser.
func (r *adminRepo) toBizAdminUser(po *model.AdminUser) *biz.AdminUser {
	if po == nil {
		return nil
	}
	return &biz.AdminUser{
		ID:       uint32(po.ID),
		Username: po.Username,
		Password: po.Password,
		Name:     po.Name,
		Status:   po.Status,
	}
}

// CreateAdminUser creates a new admin user.
func (r *adminRepo) CreateAdminUser(ctx context.Context, user *biz.AdminUser) (*biz.AdminUser, error) {
	po := &model.AdminUser{
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	user.ID = uint32(po.ID) // Assign the generated ID back to biz.AdminUser
	return r.toBizAdminUser(po), nil
}

// GetAdminUserByUsername gets an admin user by username.
func (r *adminRepo) GetAdminUserByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var po model.AdminUser
	if err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return r.toBizAdminUser(&po), nil
}
