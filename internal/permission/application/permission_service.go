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

// NewPermissionService creates a new PermissionService facade.
func NewPermissionService(manager *PermissionManager, query *PermissionQuery) *PermissionService {
	return &PermissionService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *PermissionService) CreateRole(ctx context.Context, name, description string, permissionIDs []uint64) (*domain.Role, error) {
	return s.manager.CreateRole(ctx, name, description, permissionIDs)
}

func (s *PermissionService) DeleteRole(ctx context.Context, id uint64) error {
	return s.manager.DeleteRole(ctx, id)
}

func (s *PermissionService) CreatePermission(ctx context.Context, code, description string) (*domain.Permission, error) {
	return s.manager.CreatePermission(ctx, code, description)
}

func (s *PermissionService) AssignRole(ctx context.Context, userID, roleID uint64) error {
	return s.manager.AssignRole(ctx, userID, roleID)
}

func (s *PermissionService) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return s.manager.RevokeRole(ctx, userID, roleID)
}

// --- 读操作（委托给 Query）---

func (s *PermissionService) GetRole(ctx context.Context, id uint64) (*domain.Role, error) {
	return s.query.GetRole(ctx, id)
}

func (s *PermissionService) ListRoles(ctx context.Context, page, pageSize int) ([]*domain.Role, int64, error) {
	return s.query.ListRoles(ctx, page, pageSize)
}

func (s *PermissionService) ListPermissions(ctx context.Context, page, pageSize int) ([]*domain.Permission, int64, error) {
	return s.query.ListPermissions(ctx, page, pageSize)
}

func (s *PermissionService) GetUserRoles(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	return s.query.GetUserRoles(ctx, userID)
}

func (s *PermissionService) CheckPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	return s.query.CheckPermission(ctx, userID, permissionCode)
}
