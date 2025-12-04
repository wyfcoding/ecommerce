package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity"     // 导入权限领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/permission/domain/repository" // 导入权限领域的仓储接口。
	"gorm.io/gorm"                                                         // 导入GORM ORM框架。
)

// PermissionRepository 结构体是 PermissionRepository 接口的MySQL实现。
type PermissionRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewPermissionRepository 创建并返回一个新的 PermissionRepository 实例。
func NewPermissionRepository(db *gorm.DB) repository.PermissionRepository {
	return &PermissionRepository{db: db}
}

// --- 角色管理 (Role methods) ---

// SaveRole 将角色实体保存到数据库。
// 如果角色已存在（通过ID），则更新其信息；如果不存在，则创建。
// 也会保存角色的 Permissions 关联。
func (r *PermissionRepository) SaveRole(ctx context.Context, role *entity.Role) error {
	// GORM的Save方法会处理主实体和其Has Many或Many2Many关联。
	// 对于Many2Many关系，Save会确保中间表的正确性。
	return r.db.WithContext(ctx).Save(role).Error
}

// GetRole 根据ID从数据库获取角色记录，并预加载其关联的权限。
func (r *PermissionRepository) GetRole(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	// Preload "Permissions" 确保在获取角色时，同时加载所有关联的权限。
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &role, nil
}

// ListRoles 从数据库列出所有角色记录，支持分页，并预加载其关联的权限。
func (r *PermissionRepository) ListRoles(ctx context.Context, offset, limit int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64
	db := r.db.WithContext(ctx).Model(&entity.Role{})
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}
	// Preload "Permissions" 确保在获取角色列表时，同时加载所有关联的权限。
	if err := db.Preload("Permissions").Offset(offset).Limit(limit).Find(&roles).Error; err != nil { // 应用分页和查找。
		return nil, 0, err
	}
	return roles, total, nil
}

// DeleteRole 根据ID从数据库删除角色记录。
// 注意：删除角色不会自动删除关联的权限，但会删除中间表 `role_permissions` 中的关联记录。
func (r *PermissionRepository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

// --- 权限管理 (Permission methods) ---

// SavePermission 将权限实体保存到数据库。
func (r *PermissionRepository) SavePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// ListPermissions 从数据库列出所有权限记录，支持分页。
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

// GetPermissionsByIDs 根据一组ID从数据库获取权限记录。
func (r *PermissionRepository) GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	if err := r.db.WithContext(ctx).Find(&permissions, ids).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// --- 用户角色关联 (UserRole methods) ---

// AssignRole 为用户分配角色。
// 在 `user_roles` 表中创建一条新的用户角色关联记录。
func (r *PermissionRepository) AssignRole(ctx context.Context, userID, roleID uint64) error {
	userRole := &entity.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	// GORM的Create方法会插入新记录。
	// 由于 UserRole 实体上定义了 uniqueIndex，重复分配会返回错误。
	return r.db.WithContext(ctx).Create(userRole).Error
}

// RevokeRole 撤销用户已分配的角色。
// 从 `user_roles` 表中删除对应的关联记录。
func (r *PermissionRepository) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&entity.UserRole{}).Error
}

// GetUserRoles 获取指定用户拥有的所有角色记录，并预加载每个角色关联的权限。
func (r *PermissionRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	var userRoles []entity.UserRole
	// Preload "Role.Permissions" 会加载 UserRole -> Role -> Permissions 的关系。
	// 确保能获取到每个角色及其所有关联的权限。
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}
	roles := make([]*entity.Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = &ur.Role // 从 UserRole 中提取关联的 Role 实体。
	}
	return roles, nil
}
