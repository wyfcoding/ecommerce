package biz

import (
	"context"
	"errors"
)

// Transaction 定义了事务管理器接口。
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AdminUser 是管理员用户的业务领域模型。
type AdminUser struct {
	ID       uint32
	Username string
	Password string
	Name     string
	Status   int32
}

// AdminRepo 定义了管理员数据仓库需要实现的接口。
type AdminRepo interface {
	CreateAdminUser(ctx context.Context, user *AdminUser) (*AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
}

// AdminUsecase 是管理员的业务用例。
type AdminUsecase struct {
	repo AdminRepo
	// TODO: Add password hasher interface
}

// NewAdminUsecase 创建一个新的 AdminUsecase。
func NewAdminUsecase(repo AdminRepo) *AdminUsecase {
	return &AdminUsecase{repo: repo}
}

// CreateAdminUser 注册一个新管理员用户。
func (uc *AdminUsecase) CreateAdminUser(ctx context.Context, username, password, name string) (*AdminUser, error) {
	// 1. 检查用户名是否已存在
	existingUser, err := uc.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码哈希 (这里简化处理，实际应用中应使用 bcrypt 等安全哈希算法)
	hashedPassword := password // TODO: Implement actual password hashing

	// 3. 创建管理员用户
	user := &AdminUser{
		Username: username,
		Password: hashedPassword,
		Name:     name,
		Status:   1, // Default status to active
	}
	createdUser, err := uc.repo.CreateAdminUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
