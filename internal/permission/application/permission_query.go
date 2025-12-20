package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain"
)

// PermissionQuery handles read operations for permissions and roles.
type PermissionQuery struct {
	repo domain.PermissionRepository
}

// NewPermissionQuery creates a new PermissionQuery instance.
func NewPermissionQuery(repo domain.PermissionRepository) *PermissionQuery {
	return &PermissionQuery{
		repo: repo,
	}
}

// GetRole 获取指定ID的角色详情。
func (q *PermissionQuery) GetRole(ctx context.Context, id uint64) (*domain.Role, error) {
	return q.repo.GetRole(ctx, id)
}

// ListRoles 获取角色列表。
func (q *PermissionQuery) ListRoles(ctx context.Context, page, pageSize int) ([]*domain.Role, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRoles(ctx, offset, pageSize)
}

// ListPermissions 获取权限列表。
func (q *PermissionQuery) ListPermissions(ctx context.Context, page, pageSize int) ([]*domain.Permission, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListPermissions(ctx, offset, pageSize)
}

// GetUserRoles 获取用户拥有的角色列表。
func (q *PermissionQuery) GetUserRoles(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	return q.repo.GetUserRoles(ctx, userID)
}

// CheckPermission 检查用户是否拥有特定权限。
func (q *PermissionQuery) CheckPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	roles, err := q.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		for _, perm := range role.Permissions {
			if perm.Code == permissionCode {
				return true, nil
			}
		}
	}
	return false, nil
}
