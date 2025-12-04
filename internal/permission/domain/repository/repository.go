package repository

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity" // 导入权限领域的实体定义。
)

// PermissionRepository 是权限模块的仓储接口。
// 它定义了对角色、权限和用户角色关联实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type PermissionRepository interface {
	// --- 角色管理 (Role methods) ---

	// SaveRole 将角色实体保存到数据存储中。
	// 如果角色已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// role: 待保存的角色实体。
	SaveRole(ctx context.Context, role *entity.Role) error
	// GetRole 根据ID获取角色实体。
	GetRole(ctx context.Context, id uint64) (*entity.Role, error)
	// ListRoles 列出所有角色实体，支持分页。
	ListRoles(ctx context.Context, offset, limit int) ([]*entity.Role, int64, error)
	// DeleteRole 根据ID删除角色实体。
	DeleteRole(ctx context.Context, id uint64) error

	// --- 权限管理 (Permission methods) ---

	// SavePermission 将权限实体保存到数据存储中。
	SavePermission(ctx context.Context, permission *entity.Permission) error
	// ListPermissions 列出所有权限实体，支持分页。
	ListPermissions(ctx context.Context, offset, limit int) ([]*entity.Permission, int64, error)
	// GetPermissionsByIDs 根据一组ID获取权限实体列表。
	GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*entity.Permission, error)

	// --- 用户角色关联 (UserRole methods) ---

	// AssignRole 为用户分配角色。
	AssignRole(ctx context.Context, userID, roleID uint64) error
	// RevokeRole 撤销用户已分配的角色。
	RevokeRole(ctx context.Context, userID, roleID uint64) error
	// GetUserRoles 获取指定用户ID的所有角色实体。
	GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error)
}
