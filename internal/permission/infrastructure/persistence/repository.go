package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain"
	"gorm.io/gorm"
)

// PermissionRepository 结构体是 PermissionRepository 接口的MySQL实现。
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建并返回一个新的 PermissionRepository 实例。
func NewPermissionRepository(db *gorm.DB) domain.PermissionRepository {
	return &PermissionRepository{db: db}
}

// --- 角色管理 (Role methods) ---

func (r *PermissionRepository) SaveRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *PermissionRepository) GetRole(ctx context.Context, id uint64) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *PermissionRepository) ListRoles(ctx context.Context, offset, limit int) ([]*domain.Role, int64, error) {
	var roles []*domain.Role
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Role{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Permissions").Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *PermissionRepository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error
}

// --- 权限管理 (Permission methods) ---

func (r *PermissionRepository) SavePermission(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *PermissionRepository) ListPermissions(ctx context.Context, offset, limit int) ([]*domain.Permission, int64, error) {
	var permissions []*domain.Permission
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Permission{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&permissions).Error; err != nil {
		return nil, 0, err
	}
	return permissions, total, nil
}

func (r *PermissionRepository) GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	if err := r.db.WithContext(ctx).Find(&permissions, ids).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// --- 用户角色关联 (UserRole methods) ---

func (r *PermissionRepository) AssignRole(ctx context.Context, userID, roleID uint64) error {
	userRole := &domain.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *PermissionRepository) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&domain.UserRole{}).Error
}

func (r *PermissionRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	var userRoles []domain.UserRole
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = &ur.Role
	}
	return roles, nil
}
