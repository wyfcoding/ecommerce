package domain

import (
	"context"
)

// PermissionRepository 是权限模块的仓储接口。
type PermissionRepository interface {
	// Role
	SaveRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, id uint64) (*Role, error)
	ListRoles(ctx context.Context, offset, limit int) ([]*Role, int64, error)
	DeleteRole(ctx context.Context, id uint64) error

	// Permission
	SavePermission(ctx context.Context, permission *Permission) error
	ListPermissions(ctx context.Context, offset, limit int) ([]*Permission, int64, error)
	GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*Permission, error)

	// UserRole
	AssignRole(ctx context.Context, userID, roleID uint64) error
	RevokeRole(ctx context.Context, userID, roleID uint64) error
	GetUserRoles(ctx context.Context, userID uint64) ([]*Role, error)
}
