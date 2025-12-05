package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity"     // 导入权限领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/permission/domain/repository" // 导入权限领域的仓储接口。
)

// PermissionService 结构体定义了权限管理相关的应用服务。
// 它协调领域层和基础设施层，处理角色、权限的创建与管理，以及用户角色分配和权限检查等业务逻辑。
type PermissionService struct {
	repo   repository.PermissionRepository // 依赖PermissionRepository接口，用于数据持久化操作。
	logger *slog.Logger                    // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewPermissionService 创建并返回一个新的 PermissionService 实例。
func NewPermissionService(repo repository.PermissionRepository, logger *slog.Logger) *PermissionService {
	return &PermissionService{
		repo:   repo,
		logger: logger,
	}
}

// CreateRole 创建一个新角色。
// ctx: 上下文。
// name: 角色名称。
// description: 角色描述。
// permissionIDs: 角色关联的权限ID列表。
// 返回created successfully的Role实体和可能发生的错误。
func (s *PermissionService) CreateRole(ctx context.Context, name, description string, permissionIDs []uint64) (*entity.Role, error) {
	// 1. 根据权限ID获取权限实体列表。
	permissions, err := s.repo.GetPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get permissions by IDs", "permission_ids", permissionIDs, "error", err)
		return nil, err
	}

	// 2. 创建角色实体。
	role := &entity.Role{
		Name:        name,
		Description: description,
		Permissions: permissions, // 关联权限实体。
	}

	// 3. 通过仓储接口保存角色。
	if err := s.repo.SaveRole(ctx, role); err != nil {
		s.logger.ErrorContext(ctx, "failed to save role", "role_name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "role created successfully", "role_id", role.ID, "role_name", name)
	return role, nil
}

// GetRole 获取指定ID的角色详情。
// ctx: 上下文。
// id: 角色ID。
// 返回Role实体和可能发生的错误。
func (s *PermissionService) GetRole(ctx context.Context, id uint64) (*entity.Role, error) {
	return s.repo.GetRole(ctx, id)
}

// ListRoles 获取角色列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回角色列表、总数和可能发生的错误。
func (s *PermissionService) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRoles(ctx, offset, pageSize)
}

// DeleteRole 删除指定ID的角色。
// ctx: 上下文。
// id: 角色ID。
// 返回可能发生的错误。
func (s *PermissionService) DeleteRole(ctx context.Context, id uint64) error {
	return s.repo.DeleteRole(ctx, id)
}

// CreatePermission 创建一个新权限。
// ctx: 上下文。
// code: 权限代码（唯一标识符）。
// description: 权限描述。
// 返回created successfully的Permission实体和可能发生的错误。
func (s *PermissionService) CreatePermission(ctx context.Context, code, description string) (*entity.Permission, error) {
	permission := &entity.Permission{
		Code:        code,
		Description: description,
	}
	if err := s.repo.SavePermission(ctx, permission); err != nil {
		s.logger.ErrorContext(ctx, "failed to save permission", "permission_code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "permission created successfully", "permission_id", permission.ID, "permission_code", code)
	return permission, nil
}

// ListPermissions 获取权限列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回权限列表、总数和可能发生的错误。
func (s *PermissionService) ListPermissions(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPermissions(ctx, offset, pageSize)
}

// AssignRole 为用户分配角色。
// ctx: 上下文。
// userID: 用户ID。
// roleID: 角色ID。
// 返回可能发生的错误。
func (s *PermissionService) AssignRole(ctx context.Context, userID, roleID uint64) error {
	return s.repo.AssignRole(ctx, userID, roleID)
}

// RevokeRole 撤销用户角色。
// ctx: 上下文。
// userID: 用户ID。
// roleID: 角色ID。
// 返回可能发生的错误。
func (s *PermissionService) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return s.repo.RevokeRole(ctx, userID, roleID)
}

// GetUserRoles 获取用户拥有的角色列表。
// ctx: 上下文。
// userID: 用户ID。
// 返回角色列表和可能发生的错误。
func (s *PermissionService) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

// CheckPermission 检查用户是否拥有特定权限。
// ctx: 上下文。
// userID: 用户ID。
// permissionCode: 权限代码。
// 返回布尔值（是否拥有权限）和可能发生的错误。
func (s *PermissionService) CheckPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	// 1. 获取用户的所有角色。
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user roles", "user_id", userID, "error", err)
		return false, err
	}

	// 2. 遍历用户的每个角色，检查其关联的权限。
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if perm.Code == permissionCode {
				return true, nil // 如果找到匹配的权限，则返回true。
			}
		}
	}
	return false, nil // 未找到匹配权限，返回false。
}
