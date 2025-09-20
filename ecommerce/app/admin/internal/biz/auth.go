package biz

import (
	"context"
	"errors"
	"time"

	"ecommerce/ecommerce/pkg/jwt" // 复用我们之前的JWT包

	"golang.org/x/crypto/bcrypt"
)

// AdminUser 是管理员用户的领域模型
type AdminUser struct {
	ID           uint
	Username     string
	PasswordHash string
	Name         string
	Status       int8
}

// AuthRepo 定义了认证授权所需的数据仓库接口
type AuthRepo interface {
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	CheckPermission(ctx context.Context, userID uint, permissionSlug string) (bool, error)
}

// AuthUsecase 负责认证授权的业务逻辑
type AuthUsecase struct {
	repo       AuthRepo
	jwtSecret  []byte
	jwtExpires time.Duration
}

// NewAuthUsecase 创建一个新的 AuthUsecase
func NewAuthUsecase(repo AuthRepo, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{
		repo:       repo,
		jwtSecret:  []byte(jwtSecret),
		jwtExpires: time.Hour * 8, // 后台登录有效期8小时
	}
}

// Login 处理管理员登录
func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (string, error) {
	// 1. 查找用户
	user, err := uc.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("username or password incorrect")
	}

	// 2. 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("username or password incorrect")
	}

	// 3. 校验用户状态
	if user.Status != 1 {
		return "", errors.New("user account is disabled")
	}

	// 4. 生成 JWT Token
	// 注意：后台JWT的载荷(claims)应该只包含后台需要的信息，如 admin_user_id
	token, _, err := jwt.GenerateToken(uint64(user.ID), user.Username, uc.jwtSecret, uc.jwtExpires)
	if err != nil {
		return "", err
	}

	return token, nil
}

// CheckPermission 检查用户是否有权限
func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID uint, permissionSlug string) (bool, error) {
	// 可以在此处增加超级管理员的逻辑，例如 userID 为 1 的用户拥有所有权限
	if userID == 1 {
		return true, nil
	}
	return uc.repo.CheckPermission(ctx, userID, permissionSlug)
}
