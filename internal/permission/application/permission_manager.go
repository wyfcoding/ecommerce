package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/permission/domain"
)

// PermissionManager 处理权限和角色的写操作。
type PermissionManager struct {
	repo   domain.PermissionRepository
	logger *slog.Logger
}

// NewPermissionManager creates a new PermissionManager instance.
func NewPermissionManager(repo domain.PermissionRepository, logger *slog.Logger) *PermissionManager {
	return &PermissionManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateRole 创建一个新角色。
func (m *PermissionManager) CreateRole(ctx context.Context, name, description string, permissionIDs []uint64) (*domain.Role, error) {
	permissions, err := m.repo.GetPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get permissions by IDs", "permission_ids", permissionIDs, "error", err)
		return nil, err
	}

	role := &domain.Role{
		Name:        name,
		Description: description,
		Permissions: permissions,
	}

	if err := m.repo.SaveRole(ctx, role); err != nil {
		m.logger.ErrorContext(ctx, "failed to save role", "role_name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "role created successfully", "role_id", role.ID, "role_name", name)
	return role, nil
}

// DeleteRole 物理删除指定角色。
func (m *PermissionManager) DeleteRole(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteRole(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete role", "role_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "role deleted", "role_id", id)
	return nil
}

// CreatePermission 创建一个新的细粒度权限码。
func (m *PermissionManager) CreatePermission(ctx context.Context, code, description string) (*domain.Permission, error) {
	permission := &domain.Permission{
		Code:        code,
		Description: description,
	}
	if err := m.repo.SavePermission(ctx, permission); err != nil {
		m.logger.ErrorContext(ctx, "failed to save permission", "permission_code", code, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "permission created successfully", "permission_id", permission.ID, "permission_code", code)
	return permission, nil
}

// AssignRole 将特定角色授予指定用户。
func (m *PermissionManager) AssignRole(ctx context.Context, userID, roleID uint64) error {
	if err := m.repo.AssignRole(ctx, userID, roleID); err != nil {
		m.logger.ErrorContext(ctx, "failed to assign role to user", "user_id", userID, "role_id", roleID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "role assigned to user", "user_id", userID, "role_id", roleID)
	return nil
}

// RevokeRole 撤销指定用户的特定角色权限。
func (m *PermissionManager) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	if err := m.repo.RevokeRole(ctx, userID, roleID); err != nil {
		m.logger.ErrorContext(ctx, "failed to revoke role from user", "user_id", userID, "role_id", roleID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "role revoked from user", "user_id", userID, "role_id", roleID)
	return nil
}
