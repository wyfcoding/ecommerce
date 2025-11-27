package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/permission/domain/repository"
)

type PermissionService struct {
	repo   repository.PermissionRepository
	logger *slog.Logger
}

func NewPermissionService(repo repository.PermissionRepository, logger *slog.Logger) *PermissionService {
	return &PermissionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *PermissionService) CreateRole(ctx context.Context, name, description string, permissionIDs []uint64) (*entity.Role, error) {
	permissions, err := s.repo.GetPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		return nil, err
	}

	role := &entity.Role{
		Name:        name,
		Description: description,
		Permissions: permissions,
	}

	if err := s.repo.SaveRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *PermissionService) GetRole(ctx context.Context, id uint64) (*entity.Role, error) {
	return s.repo.GetRole(ctx, id)
}

func (s *PermissionService) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRoles(ctx, offset, pageSize)
}

func (s *PermissionService) DeleteRole(ctx context.Context, id uint64) error {
	return s.repo.DeleteRole(ctx, id)
}

func (s *PermissionService) CreatePermission(ctx context.Context, code, description string) (*entity.Permission, error) {
	permission := &entity.Permission{
		Code:        code,
		Description: description,
	}
	if err := s.repo.SavePermission(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}

func (s *PermissionService) ListPermissions(ctx context.Context, page, pageSize int) ([]*entity.Permission, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPermissions(ctx, offset, pageSize)
}

func (s *PermissionService) AssignRole(ctx context.Context, userID, roleID uint64) error {
	return s.repo.AssignRole(ctx, userID, roleID)
}

func (s *PermissionService) RevokeRole(ctx context.Context, userID, roleID uint64) error {
	return s.repo.RevokeRole(ctx, userID, roleID)
}

func (s *PermissionService) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *PermissionService) CheckPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID)
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
