// package biz 定义了业务逻辑层(Business Logic Layer)。
package biz

import "context"

// User 是用户的业务领域模型。
type User struct {
	UserID   uint64
	Username string
	Password string
	Nickname *string // 使用指针以区分 "未提供" 与 "设置为空字符串"
	Avatar   *string
	Gender   *int32
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
