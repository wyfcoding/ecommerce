package domain

import "context"

// UserRepository 用户仓储接口
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

// AddressRepository 地址仓储接口
type AddressRepository interface {
	Save(ctx context.Context, address *Address) error
	FindByID(ctx context.Context, id uint) (*Address, error)
	FindDefaultByUserID(ctx context.Context, userID uint) (*Address, error)
	FindByUserID(ctx context.Context, userID uint) ([]*Address, error)
	Update(ctx context.Context, address *Address) error
	Delete(ctx context.Context, id uint) error
	SetDefault(ctx context.Context, userID, addressID uint) error
}
