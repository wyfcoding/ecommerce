package data

import (
	"context"

	"ecommerce/ecommerce/app/admin/internal/biz"

	"gorm.io/gorm"
)

type rbacRepo struct {
	db *gorm.DB
}

func NewRbacRepo(db *gorm.DB) biz.RbacRepo {
	return &rbacRepo{db: db}
}

func (r *rbacRepo) CreateAdminUser(ctx context.Context, username, passwordHash, name string) (*biz.AdminUser, error) {
	user := AdminUser{Username: username, PasswordHash: passwordHash, Name: name, Status: 1}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}
	// 模型转换...
	return &bizUser, nil
}

func (r *rbacRepo) CreateRole(ctx context.Context, name, slug string) (*biz.Role, error) {
	// ...
}

func (r *rbacRepo) ListRoles(ctx context.Context) ([]*biz.Role, error) {
	// ...
}

func (r *rbacRepo) ListPermissions(ctx context.Context) ([]*biz.Permission, error) {
	// ...
}

// UpdateUserRoles 是一个事务操作
func (r *rbacRepo) UpdateUserRoles(ctx context.Context, userID uint32, roleIDs []uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除该用户的所有旧角色关系
		if err := tx.Where("user_id = ?", userID).Delete(&AdminUserRoleMap{}).Error; err != nil {
			return err
		}

		// 2. 如果要分配新角色，则批量插入新关系
		if len(roleIDs) > 0 {
			maps := make([]AdminUserRoleMap, len(roleIDs))
			for i, roleID := range roleIDs {
				maps[i] = AdminUserRoleMap{UserID: userID, RoleID: roleID}
			}
			if err := tx.Create(&maps).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateRolePermissions 同样是事务操作
func (r *rbacRepo) UpdateRolePermissions(ctx context.Context, roleID uint32, permissionIDs []uint32) error {
	// ... 逻辑与 UpdateUserRoles 类似 ...
}
