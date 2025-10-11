package repository

import (
	"context"

	"ecommerce/internal/user/model" // Import the new model package
)

// UserRepo 定义了用户数据仓库需要实现的接口。
type UserRepo interface {
	// CreateUser 在数据仓库中创建一个新用户。
	CreateUser(ctx context.Context, u *model.User) (*model.User, error)

	// GetUserByUsername 根据用户名从数据仓库中获取用户信息。
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)

	// GetUserByUserID 根据用户的业务ID获取用户信息。
	GetUserByUserID(ctx context.Context, userID uint64) (*model.User, error)

	// UpdateUser 更新数据仓库中的用户信息。
	UpdateUser(ctx context.Context, u *model.User) (*model.User, error)
}

// AddressRepo 定义了地址数据仓库需要实现的接口。
type AddressRepo interface {
	CreateAddress(ctx context.Context, addr *model.Address) (*model.Address, error)
	UpdateAddress(ctx context.Context, addr *model.Address) (*model.Address, error)
	DeleteAddress(ctx context.Context, userID, addrID uint64) error
	GetAddress(ctx context.Context, userID, addrID uint64) (*model.Address, error)
	ListAddresses(ctx context.Context, userID uint64) ([]*model.Address, error)
	SetDefaultAddress(ctx context.Context, userID, addrID uint64) error
}
