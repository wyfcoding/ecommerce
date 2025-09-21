package data

import (
	"context"
	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/data/model"
	"gorm.io/gorm"
)

type authRepo struct {
	*Data
}

// NewAuthRepo 是 authRepo 的构造函数。
func NewAuthRepo(data *Data) biz.AuthRepo {
	return &authRepo{Data: data}
}

// GetAdminUserByUsername 根据用户名从数据仓库中获取管理员用户信息。
func (r *authRepo) GetAdminUserByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.AdminUser{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}, nil
}

// GetAdminUserByID 根据ID从数据仓库中获取管理员用户信息。
func (r *authRepo) GetAdminUserByID(ctx context.Context, id uint32) (*biz.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.AdminUser{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}, nil
}
