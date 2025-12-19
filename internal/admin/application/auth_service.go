package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain" // Assuming there is a password helper
	"golang.org/x/crypto/bcrypt"
)

// AdminAuthService 处理认证和权限
type AdminAuthService struct {
	userRepo domain.AdminRepository
	roleRepo domain.RoleRepository
	logger   *slog.Logger
}

func NewAdminAuthService(userRepo domain.AdminRepository, roleRepo domain.RoleRepository, logger *slog.Logger) *AdminAuthService {
	return &AdminAuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// Login 验证密码并返回用户信息（后续可集成JWT）
func (s *AdminAuthService) Login(ctx context.Context, username, password string) (*domain.AdminUser, error) {
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

	// 验证密码 (假设存储的是 bcrypt hash)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	// 更新登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	return user, nil
}

// CheckPermission 检查用户是否拥有特定权限
func (s *AdminAuthService) CheckPermission(ctx context.Context, userID uint, requiredPerm string) (bool, error) {
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
func (s *AdminAuthService) CreateUser(ctx context.Context, user *domain.AdminUser, password string) error {
	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashed)
	return s.userRepo.Create(ctx, user)
}
