package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/permission/domain/repository"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) repository.PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) SaveRole(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *PermissionRepository) GetRole(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *PermissionRepository) ListRoles(ctx context.Context, offset, limit int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64
	db := r.db.WithContext(ctx).Model(&entity.Role{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Permissions").Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *PermissionRepository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *PermissionRepository) SavePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *PermissionRepository) ListPermissions(ctx context.Context, offset, limit int) ([]*entity.Permission, int64, error) {
	var permissions []*entity.Permission
	var total int64
	db := r.db.WithContext(ctx).Model(&entity.Permission{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&permissions).Error; err != nil {
		return nil, 0, err
	}
	return permissions, total, nil
}

func (r *PermissionRepository) GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	if err := r.db.WithContext(ctx).Find(&permissions, ids).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) AssignRole(ctx context.Context, userID, roleID uint64) error {
	userRole := &entity.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *PermissionRepository) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&entity.UserRole{}).Error
}

func (r *PermissionRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	var userRoles []entity.UserRole
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}
	roles := make([]*entity.Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = &ur.Role
	}
	return roles, nil
}
