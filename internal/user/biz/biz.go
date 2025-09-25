// package biz 定义了业务逻辑层(Business Logic Layer)。
package biz

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// User 是用户的业务领域模型。
type User struct {
	ID        uint64
	Username  string
	Password  string
	Nickname  string
	Avatar    string
	Gender    int32
	Birthday  time.Time
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepo 定义了用户数据仓库需要实现的接口。
type UserRepo interface {
	// CreateUser 在数据仓库中创建一个新用户。
	CreateUser(ctx context.Context, u *User) (*User, error)

	// GetUserByUsername 根据用户名从数据仓库中获取用户信息。
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByUserID 根据用户的业务ID获取用户信息。
	GetUserByUserID(ctx context.Context, userID uint64) (*User, error)

	// UpdateUser 更新数据仓库中的用户信息。
	UpdateUser(ctx context.Context, u *User) (*User, error)
}

// Address 是收货地址的业务领域模型。
type Address struct {
	ID              uint64
	UserID          uint64
	Name            *string
	Phone           *string
	Province        *string
	City            *string
	District        *string
	DetailedAddress *string
	IsDefault       *bool
}

// AddressRepo 定义了地址数据仓库需要实现的接口。
type AddressRepo interface {
	CreateAddress(ctx context.Context, addr *Address) (*Address, error)
	UpdateAddress(ctx context.Context, addr *Address) (*Address, error)
	DeleteAddress(ctx context.Context, userID, addrID uint64) error
	GetAddress(ctx context.Context, userID, addrID uint64) (*Address, error)
	ListAddresses(ctx context.Context, userID uint64) ([]*Address, error)
	SetDefaultAddress(ctx context.Context, userID, addrID uint64) error
}

// UserUsecase 是用户的业务用例。
type UserUsecase struct {
	repo UserRepo
	// TODO: Add password hasher interface
}

// NewUserUsecase 创建一个新的 UserUsecase。
func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// RegisterUser 注册一个新用户。
func (uc *UserUsecase) RegisterUser(ctx context.Context, username, password string) (*User, error) {
	// 1. 检查用户名是否已存在
	existingUser, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil && err != gorm.ErrRecordNotFound { // Assuming gorm.ErrRecordNotFound for "not found"
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码哈希 (这里简化处理，实际应用中应使用 bcrypt 等安全哈希算法)
	hashedPassword := password // TODO: Implement actual password hashing

	// 3. 创建用户
	user := &User{
		Username: username,
		Password: hashedPassword,
		// 其他字段可以根据需要设置默认值或从请求中获取
	}
	createdUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
