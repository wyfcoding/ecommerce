package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/model"
	"ecommerce/internal/user/repository" // Import the new repository package
	"ecommerce/pkg/hash"
	"ecommerce/pkg/jwt"
	"ecommerce/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Error definitions
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrPasswordIncorrect = errors.New("incorrect password")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserService is the gRPC service implementation.
type UserService struct {
	v1.UnimplementedUserServer

	userUsecase    *UserUsecase // Renamed from biz.UserUsecase
	addressUsecase *AddressUsecase // Assuming AddressUsecase will also be moved here

	// JWT configuration
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewUserService is the constructor for UserService.
func NewUserService(userUC *UserUsecase, addressUC *AddressUsecase, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *UserService {
	return &UserService{
		userUsecase:    userUC,
		addressUsecase: addressUC,
		jwtSecret:      jwtSecret,
		jwtIssuer:      jwtIssuer,
		jwtExpire:      jwtExpire,
	}
}

// UserUsecase 封装了用户相关的业务逻辑。
type UserUsecase struct {
	repo      repository.UserRepo // Use the new repository interface
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewUserUsecase 是 UserUsecase 的构造函数。
func NewUserUsecase(repo repository.UserRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpire: jwtExpire,
	}
}

// Register 负责处理用户注册的业务流程。
func (uc *UserUsecase) Register(ctx context.Context, username, password string) (*model.User, error) {
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

	user := &model.User{
		ID:       snowflake.GenID(), // Assuming snowflake.GenID() is available
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

	claims := jwt.CustomClaims{ // Assuming jwt.CustomClaims is defined
		UserID:   user.ID,
		Username: user.Username,
	}
	claims.ExpiresAt = time.Now().Add(uc.jwtExpire).Unix()
	claims.Issuer = uc.jwtIssuer

	token, err := jwt.GenerateToken(user.ID, user.Username, uc.jwtSecret, uc.jwtIssuer, uc.jwtExpire, jwt.SigningMethodHS256)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserByID 负责根据用户ID获取用户信息的业务逻辑。
func (uc *UserUsecase) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
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
func (uc *UserUsecase) UpdateUserInfo(ctx context.Context, u *model.User) (*model.User, error) {
	if u.ID == 0 {
		return nil, fmt.Errorf("user id is required for update")
	}

	// 示例：添加昵称长度校验
	if u.Nickname != "" { // Changed from *u.Nickname == ""
		return nil, errors.New("nickname cannot be empty")
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

// RegisterUser 注册一个新用户。(from biz.go)
func (uc *UserUsecase) RegisterUser(ctx context.Context, username, password string) (*model.User, error) {
	// 1. 检查用户名是否已存在
	_, err := uc.repo.GetUserByUsername(ctx, username)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, errors.New("username already exists")
		}
		return nil, err
	}

	// 2. 密码哈希
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. 创建用户
	user := &model.User{
		Username: username,
		Password: hashedPassword,
	}
	createdUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// VerifyPassword 验证用户密码.(from biz.go)
func (uc *UserUsecase) VerifyPassword(ctx context.Context, username, password string) (*model.User, error) {
	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found or other error")
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// AddressUsecase 封装了地址相关的业务逻辑。
type AddressUsecase struct {
	repo repository.AddressRepo // Use the new repository interface
}

// NewAddressUsecase 是 AddressUsecase 的构造函数。
func NewAddressUsecase(repo repository.AddressRepo) *AddressUsecase {
	return &AddressUsecase{repo: repo}
}

// validatePhoneNumber 校验手机号格式。
func (uc *AddressUsecase) validatePhoneNumber(phone string) bool {
	// 一个简单的手机号校验规则 (11位数字，以1开头)
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(phone)
}

// CreateAddress 负责创建收货地址的业务逻辑。
func (uc *AddressUsecase) CreateAddress(ctx context.Context, addr *model.Address) (*model.Address, error) {
	// 业务校验：例如手机号格式
	if addr.Phone == nil || !uc.validatePhoneNumber(*addr.Phone) {
		return nil, errors.New("invalid phone number format")
	}
	// 其他校验，如地址库校验、字段非空等可在此处添加
	if addr.Name == nil || *addr.Name == "" {
		return nil, errors.New("recipient name cannot be empty")
	}

	return uc.repo.CreateAddress(ctx, addr)
}

// UpdateAddress 负责更新收货地址的业务逻辑。
func (uc *AddressUsecase) UpdateAddress(ctx context.Context, addr *model.Address) (*model.Address, error) {
	if addr.Phone != nil && !uc.validatePhoneNumber(*addr.Phone) {
		return nil, fmt.Errorf("invalid phone number format")
	}
	return uc.repo.UpdateAddress(ctx, addr)
}

// DeleteAddress 负责删除收货地址的业务逻辑。
func (uc *AddressUsecase) DeleteAddress(ctx context.Context, userID, addrID uint64) error {
	return uc.repo.DeleteAddress(ctx, userID, addrID)
}

// GetAddress 负责获取单个地址的业务逻辑。
func (uc *AddressUsecase) GetAddress(ctx context.Context, userID, addrID uint64) (*model.Address, error) {
	return uc.repo.GetAddress(ctx, userID, addrID)
}

// ListAddresses 负责获取地址列表的业务逻辑。
func (uc *AddressUsecase) ListAddresses(ctx context.Context, userID uint64) ([]*model.Address, error) {
	return uc.repo.ListAddresses(ctx, userID)
}

// SetDefaultAddress 负责设置默认地址的业务逻辑。
func (uc *AddressUsecase) SetDefaultAddress(ctx context.Context, userID, addrID uint64) error {
	return uc.repo.SetDefaultAddress(ctx, userID, addrID)
}
