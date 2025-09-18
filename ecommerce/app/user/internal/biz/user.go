package biz

import (
	"context"
	"ecommerce/ecommerce/app/user/internal/data"
	"ecommerce/ecommerce/pkg/snowflake"

	"ecommerce/ecommerce/pkg/jwt"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// JWTSecret 应该从配置中加载
	JWTSecret = []byte("your-very-secret-key")
	// JWTExpires 是 token 的有效期
	JWTExpires = time.Hour * 24 * 7
)

// UserUsecase 是用户业务逻辑的容器
type UserUsecase struct {
	repo data.UserRepo
}

// NewUserUsecase 创建一个新的 UserUsecase
func NewUserUsecase(repo data.UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// LoginByPassword 实现了登录的业务逻辑
func (uc *UserUsecase) LoginByPassword(ctx context.Context, username, password string) (string, int64, error) {
	// 1. 根据用户名从数据库中查找用户认证信息
	userAuth, err := uc.repo.GetUserAuthByIdentifier(ctx, "password", username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, errors.New("user not found or password incorrect")
		}
		return "", 0, err
	}

	// 2. 比较数据库中存储的哈希密码和用户提供的密码
	err = bcrypt.CompareHashAndPassword([]byte(userAuth.Credential), []byte(password))
	if err != nil {
		// 如果比较失败，也返回统一的错误信息，防止攻击者猜测用户名是否存在
		return "", 0, errors.New("user not found or password incorrect")
	}

	// 3. 密码验证成功，生成 JWT
	token, expiresAt, err := jwt.GenerateToken(userAuth.UserID, username, JWTSecret, JWTExpires)
	if err != nil {
		return "", 0, err
	}

	return token, expiresAt, nil
}

// CreateUserByPassword 实现了通过用户名密码创建用户的业务逻辑
func (uc *UserUsecase) CreateUserByPassword(ctx context.Context, username, password string) (int64, error) {
	// 1. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// 2. 生成分布式用户ID
	userID := snowflake.GenID()

	// 3. 准备数据模型
	userBasic := &data.UserBasic{
		UserID:   uint64(userID),
		Username: username,
		Nickname: username, // 默认昵称与用户名相同
	}
	userAuth := &data.UserAuth{
		UserID:     uint64(userID),
		AuthType:   "password",
		Identifier: username,
		Credential: string(hashedPassword),
	}

	// 4. 调用数据层写入数据
	err = uc.repo.CreateUser(ctx, userBasic, userAuth)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
