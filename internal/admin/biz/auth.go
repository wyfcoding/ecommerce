package biz

import (
	"context"
	"errors"
	"time"

	"ecommerce/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrAdminUserNotFound      = errors.New("管理员用户不存在")
	ErrAdminPasswordIncorrect = errors.New("管理员密码不正确")
)

// AdminUser 是管理员用户的业务领域模型。
type AdminUser struct {
	ID       uint32
	Username string
	Password string
	Name     string
	Status   int32
}

// AuthRepo 定义了认证数据仓库的接口。
type AuthRepo interface {
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	GetAdminUserByID(ctx context.Context, id uint32) (*AdminUser, error)
}

// AuthUsecase 封装了认证相关的业务逻辑。
type AuthUsecase struct {
	repo AuthRepo
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewAuthUsecase 是 AuthUsecase 的构造函数。
func NewAuthUsecase(repo AuthRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *AuthUsecase {
	return &AuthUsecase{
		repo: repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpire: jwtExpire,
	}
}

// AdminLogin 负责管理员登录的业务逻辑。
func (uc *AuthUsecase) AdminLogin(ctx context.Context, username, password string) (string, error) {
	user, err := uc.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAdminUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrAdminPasswordIncorrect
	}

	// 检查用户状态
	if user.Status != 1 { // 假设 1 为正常状态
		return "", errors.New("管理员账户已被禁用或状态异常")
	}

	claims := jwt.CustomClaims{
		UserID:   uint64(user.ID), // AdminUser ID 是 uint32 类型，转换为 uint64 以用于 CustomClaims
		Username: user.Username,
	}
	claims.ExpiresAt = time.Now().Add(uc.jwtExpire).Unix()
	claims.Issuer = uc.jwtIssuer

	token, err := jwt.GenerateToken(claims, uc.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetJwtSecret 返回 JWT 密钥。
func (uc *AuthUsecase) GetJwtSecret() string {
	return uc.jwtSecret
}

// GetAdminUserByID 负责根据ID获取管理员用户信息的业务逻辑。
func (uc *AuthUsecase) GetAdminUserByID(ctx context.Context, id uint32) (*AdminUser, error) {
	user, err := uc.repo.GetAdminUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return user, nil
}
