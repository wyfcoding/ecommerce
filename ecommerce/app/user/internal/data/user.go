package data

import (
	"context"

	"gorm.io/gorm"
)

// UserRepo 定义了用户数据仓库的接口
type UserRepo interface {
	CreateUser(ctx context.Context, ub *UserBasic, ua *UserAuth) error
	GetUserAuthByIdentifier(ctx context.Context, authType string, identifier string) (*UserAuth, error)
}

// userRepo 是 UserRepo 的实现
type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建一个新的 UserRepo
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

// CreateUser 在一个事务中创建用户基础信息和认证信息
func (r *userRepo) CreateUser(ctx context.Context, ub *UserBasic, ua *UserAuth) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建用户基础信息
		if err := tx.Create(ub).Error; err != nil {
			return err
		}
		// 创建用户认证信息
		if err := tx.Create(ua).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetUserAuthByIdentifier 根据唯一标识查询用户认证信息
func (r *userRepo) GetUserAuthByIdentifier(ctx context.Context, authType string, identifier string) (*UserAuth, error) {
	var userAuth UserAuth
	err := r.db.WithContext(ctx).
		Where("auth_type = ? AND identifier = ?", authType, identifier).
		First(&userAuth).Error
	if err != nil {
		return nil, err // gorm.ErrRecordNotFound 也会被直接返回
	}
	return &userAuth, nil
}
