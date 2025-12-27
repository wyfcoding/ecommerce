package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/security"
)

// AdminAuth 处理认证和权限
type AdminAuth struct {
	userRepo domain.AdminRepository
	roleRepo domain.RoleRepository
	logger   *slog.Logger
}

// NewAdminAuth 定义了 NewAdminAuth 相关的服务逻辑。
func NewAdminAuth(userRepo domain.AdminRepository, roleRepo domain.RoleRepository, logger *slog.Logger) *AdminAuth {
	return &AdminAuth{
		userRepo: userRepo,
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// Login 验证密码并返回用户信息（后续可集成JWT）
func (s *AdminAuth) Login(ctx context.Context, username, password string) (*domain.AdminUser, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Status != domain.UserStatusActive {
		return nil, errors.New("user is disabled")
	}

	// 验证密码 (使用 pkg/security 封装的 bcrypt 验证)
	if !security.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	// 更新登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	return user, nil
}

// CheckPermission 检查用户是否拥有特定权限
func (s *AdminAuth) CheckPermission(ctx context.Context, userID uint, requiredPerm string) (bool, error) {
	perms, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == requiredPerm {
			return true, nil
		}
		// 支持超级管理员通配符
		if p == "*:*" {
			return true, nil
		}
	}
	return false, nil
}

// CreateUser 创建新管理员
func (s *AdminAuth) CreateUser(ctx context.Context, user *domain.AdminUser, password string) error {
	// 对密码进行哈希 (使用 pkg/security 封装的 bcrypt 哈希)
	hashed, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	user.PasswordHash = hashed
	return s.userRepo.Create(ctx, user)
}
