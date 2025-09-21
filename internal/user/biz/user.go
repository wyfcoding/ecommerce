package biz

import (
	"context"
	"errors"
	"fmt" // 导入 fmt 包
	"time"

	"ecommerce/pkg/jwt"
	"ecommerce/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrPasswordIncorrect = errors.New("密码不正确")
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// UserUsecase 封装了用户相关的业务逻辑。
type UserUsecase struct {
	repo      UserRepo
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewUserUsecase 是 UserUsecase 的构造函数。
func NewUserUsecase(repo UserRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpire: jwtExpire,
	}
}

// Register 负责处理用户注册的业务流程。
func (uc *UserUsecase) Register(ctx context.Context, username, password string) (*User, error) {
	_, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		UserID:   snowflake.GenID(),
		Username: username,
		Password: string(hashedPassword),
	}

	return uc.repo.CreateUser(ctx, user)
}

// Login 负责处理用户登录的业务流程。
func (uc *UserUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrPasswordIncorrect
	}

	claims := jwt.CustomClaims{
		UserID:   user.UserID,
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

// GetUserByID 负责根据用户ID获取用户信息的业务逻辑。
func (uc *UserUsecase) GetUserByID(ctx context.Context, userID uint64) (*User, error) {
	user, err := uc.repo.GetUserByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateUserInfo 负责更新用户个人信息的业务逻辑。
func (uc *UserUsecase) UpdateUserInfo(ctx context.Context, u *User) (*User, error) {
	if u.UserID == 0 {
		return nil, fmt.Errorf("user id is required for update")
	}

	// 示例：添加昵称长度校验
	if u.Nickname != nil && *u.Nickname == "" {
		return nil, errors.New("昵称不能为空")
	}

	// 此处可添加更复杂的业务校验，如昵称黑名单、头像URL格式验证等。
	// 当前实现直接委托给数据仓库层。

	updatedUser, err := uc.repo.UpdateUser(ctx, u)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return updatedUser, nil
}

// GetJwtSecret 返回 JWT 密钥。
func (uc *UserUsecase) GetJwtSecret() string {
	return uc.jwtSecret
}

