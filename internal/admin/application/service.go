package application

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/admin/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/jwt"

	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo   repository.AdminRepository
	logger *slog.Logger
}

func NewAdminService(repo repository.AdminRepository, logger *slog.Logger) *AdminService {
	return &AdminService{
		repo:   repo,
		logger: logger,
	}
}

// Admin methods

func (s *AdminService) RegisterAdmin(ctx context.Context, username, email, password, realName, phone string) (*entity.Admin, error) {
	// Check if username or email exists
	if _, err := s.repo.GetAdminByUsername(ctx, username); err == nil {
		return nil, entity.ErrUsernameExists
	}
	if _, err := s.repo.GetAdminByEmail(ctx, email); err == nil {
		return nil, entity.ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := entity.NewAdmin(username, email, string(hashedPassword), realName, phone)
	if err := s.repo.CreateAdmin(ctx, admin); err != nil {
		s.logger.ErrorContext(ctx, "failed to create admin", "username", username, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "admin created successfully", "admin_id", admin.ID, "username", username)

	return admin, nil
}

func (s *AdminService) Login(ctx context.Context, username, password, ip string) (string, error) {
	admin, err := s.repo.GetAdminByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !admin.IsActive() {
		return "", errors.New("account is inactive")
	}
	if admin.IsLocked() {
		return "", errors.New("account is locked")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		admin.RecordLoginFailure()
		s.repo.UpdateAdmin(ctx, admin)
		s.repo.CreateLoginLog(ctx, &entity.LoginLog{
			AdminID: uint64(admin.ID),
			IP:      ip,
			Success: false,
			Reason:  "invalid password",
		})
		return "", errors.New("invalid credentials")
	}

	admin.RecordLoginSuccess(ip)
	s.repo.UpdateAdmin(ctx, admin)
	s.repo.CreateLoginLog(ctx, &entity.LoginLog{
		AdminID: uint64(admin.ID),
		IP:      ip,
		Success: true,
	})

	// Generate JWT token
	// TODO: Load secret key from config
	token, err := jwt.GenerateToken(uint64(admin.ID), admin.Username, "your-secret-key", "admin-service", 24*time.Hour, nil)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AdminService) GetAdminProfile(ctx context.Context, id uint64) (*entity.Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

func (s *AdminService) ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error) {
	return s.repo.ListAdmins(ctx, page, pageSize)
}

// Role methods

func (s *AdminService) CreateRole(ctx context.Context, name, code, description string) (*entity.Role, error) {
	if _, err := s.repo.GetRoleByCode(ctx, code); err == nil {
		return nil, entity.ErrRoleCodeExists
	}

	role := &entity.Role{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      1,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		s.logger.ErrorContext(ctx, "failed to create role", "role_code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "role created successfully", "role_id", role.ID, "role_code", code)

	return role, nil
}

func (s *AdminService) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	return s.repo.ListRoles(ctx, page, pageSize)
}

func (s *AdminService) AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error {
	return s.repo.AssignRoleToAdmin(ctx, adminID, roleID)
}

// Permission methods

func (s *AdminService) CreatePermission(ctx context.Context, name, code, permType, path, method string, parentID uint64) (*entity.Permission, error) {
	if _, err := s.repo.GetPermissionByCode(ctx, code); err == nil {
		return nil, entity.ErrPermCodeExists
	}

	permission := &entity.Permission{
		Name:     name,
		Code:     code,
		Type:     permType,
		Path:     path,
		Method:   method,
		ParentID: parentID,
		Status:   1,
	}

	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		s.logger.ErrorContext(ctx, "failed to create permission", "perm_code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "permission created successfully", "perm_id", permission.ID, "perm_code", code)

	return permission, nil
}

func (s *AdminService) ListPermissions(ctx context.Context) ([]*entity.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *AdminService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	return s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
}

// CheckPermission 检查权限
func (s *AdminService) CheckPermission(ctx context.Context, adminID uint64, path, method string) (bool, error) {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return false, err
	}

	// 1. Check direct permissions (if implemented)
	// 2. Check role permissions
	for _, role := range admin.Roles {
		permissions, err := s.repo.GetPermissionsByRoleID(ctx, uint64(role.ID))
		if err != nil {
			continue
		}
		for _, perm := range permissions {
			if perm.Path == path && perm.Method == method {
				return true, nil
			}
			// Wildcard support could be added here
		}
	}

	return false, nil
}
