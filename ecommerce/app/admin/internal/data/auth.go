package data

import (
	"context"

	"ecommerce/ecommerce/app/admin/internal/biz"

	"gorm.io/gorm"
)

type authRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) biz.AuthRepo {
	return &authRepo{db: db}
}

func (r *authRepo) GetAdminUserByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var user AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.AdminUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		Status:       user.Status,
	}, nil
}

func (r *authRepo) CheckPermission(ctx context.Context, userID uint, permissionSlug string) (bool, error) {
	var count int64

	// 通过 JOIN 查询来检查用户、角色、权限之间的关联是否存在
	err := r.db.WithContext(ctx).Model(&AdminUserRoleMap{}).
		Joins("JOIN admin_role_permission_map ON admin_user_role_map.role_id = admin_role_permission_map.role_id").
		Joins("JOIN admin_permission ON admin_role_permission_map.permission_id = admin_permission.id").
		Where("admin_user_role_map.user_id = ?", userID).
		Where("admin_permission.slug = ?", permissionSlug).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
