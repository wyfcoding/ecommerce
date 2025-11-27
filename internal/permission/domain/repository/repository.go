package repository

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity"
)

type PermissionRepository interface {
	// Role
	SaveRole(ctx context.Context, role *entity.Role) error
	GetRole(ctx context.Context, id uint64) (*entity.Role, error)
	ListRoles(ctx context.Context, offset, limit int) ([]*entity.Role, int64, error)
	DeleteRole(ctx context.Context, id uint64) error

	// Permission
	SavePermission(ctx context.Context, permission *entity.Permission) error
	ListPermissions(ctx context.Context, offset, limit int) ([]*entity.Permission, int64, error)
	GetPermissionsByIDs(ctx context.Context, ids []uint64) ([]*entity.Permission, error)

	// User Role
	AssignRole(ctx context.Context, userID, roleID uint64) error
	RevokeRole(ctx context.Context, userID, roleID uint64) error
	GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error)
}
