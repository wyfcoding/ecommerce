package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/permission/domain"
)

// PermissionService 作为权限操作的门面。
type PermissionService struct {
	manager *PermissionManager
	query   *PermissionQuery
}

// NewPermissionService 创建权限服务门面实例。
func NewPermissionService(manager *PermissionManager, query *PermissionQuery) *PermissionService {
	return &PermissionService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateRole 创建一个新的角色并分配权限。
func (s *PermissionService) CreateRole(ctx context.Context, name, description string, permissionIDs []uint64) (*domain.Role, error) {
	return s.manager.CreateRole(ctx, name, description, permissionIDs)
}

// DeleteRole 删除指定的角色。
func (s *PermissionService) DeleteRole(ctx context.Context, id uint64) error {
	return s.manager.DeleteRole(ctx, id)
}

// CreatePermission 创建一个新的权限项（权限点）。
func (s *PermissionService) CreatePermission(ctx context.Context, code, description string) (*domain.Permission, error) {
	return s.manager.CreatePermission(ctx, code, description)
}

// AssignRole 为用户分配一个角色。
func (s *PermissionService) AssignRole(ctx context.Context, userID, roleID uint64) error {
	return s.manager.AssignRole(ctx, userID, roleID)
}

// RevokeRole 撤销用户的某个角色。
func (s *PermissionService) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return s.manager.RevokeRole(ctx, userID, roleID)
}

// --- 读操作（委托给 Query）---

// GetRole 获取指定ID的角色详情及其权限列表。
func (s *PermissionService) GetRole(ctx context.Context, id uint64) (*domain.Role, error) {
	return s.query.GetRole(ctx, id)
}

// ListRoles 分页获取角色列表。
func (s *PermissionService) ListRoles(ctx context.Context, page, pageSize int) ([]*domain.Role, int64, error) {
	return s.query.ListRoles(ctx, page, pageSize)
}

// ListPermissions 分页获取所有权限项列表。
func (s *PermissionService) ListPermissions(ctx context.Context, page, pageSize int) ([]*domain.Permission, int64, error) {
	return s.query.ListPermissions(ctx, page, pageSize)
}

// GetUserRoles 获取指定用户所拥有的所有角色。
func (s *PermissionService) GetUserRoles(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	return s.query.GetUserRoles(ctx, userID)
}

// CheckPermission 验证用户是否拥有指定的权限。
func (s *PermissionService) CheckPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	return s.query.CheckPermission(ctx, userID, permissionCode)
}
