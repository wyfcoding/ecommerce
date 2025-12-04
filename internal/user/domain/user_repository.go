package domain

import "context"

// UserRepository 是用户模块的仓储接口。
// 它定义了对 User 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type UserRepository interface {
	// Save 将用户实体保存到数据存储中。
	// 如果用户已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// user: 待保存的用户实体。
	Save(ctx context.Context, user *User) error
	// FindByID 根据ID获取用户实体。
	FindByID(ctx context.Context, id uint) (*User, error)
	// FindByUsername 根据用户名获取用户实体。
	FindByUsername(ctx context.Context, username string) (*User, error)
	// FindByEmail 根据邮箱获取用户实体。
	FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByPhone 根据手机号获取用户实体。
	FindByPhone(ctx context.Context, phone string) (*User, error)
	// Update 更新用户实体。
	Update(ctx context.Context, user *User) error
	// Delete 根据ID删除用户实体。
	Delete(ctx context.Context, id uint) error
	// List 列出所有用户实体，支持分页。
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

// AddressRepository 是用户地址模块的仓储接口。
// 它定义了对 Address 实体进行数据持久化操作的契约。
type AddressRepository interface {
	// Save 将地址实体保存到数据存储中。
	Save(ctx context.Context, address *Address) error
	// FindByID 根据ID获取地址实体。
	FindByID(ctx context.Context, id uint) (*Address, error)
	// FindDefaultByUserID 获取指定用户ID的默认地址实体。
	FindDefaultByUserID(ctx context.Context, userID uint) (*Address, error)
	// FindByUserID 获取指定用户ID的所有地址实体。
	FindByUserID(ctx context.Context, userID uint) ([]*Address, error)
	// Update 更新地址实体。
	Update(ctx context.Context, address *Address) error
	// Delete 根据ID删除地址实体。
	Delete(ctx context.Context, id uint) error
	// SetDefault 设置指定用户ID的默认地址，同时取消该用户其他地址的默认状态。
	SetDefault(ctx context.Context, userID, addressID uint) error
}
