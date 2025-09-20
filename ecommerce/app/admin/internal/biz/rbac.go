package biz

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

// RbacRepo 定义了 RBAC 管理所需的数据仓库接口
type RbacRepo interface {
	CreateRole(ctx context.Context, name, slug string) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	UpdateRolePermissions(ctx context.Context, roleID uint32, permissionIDs []uint32) error
	ListPermissions(ctx context.Context) ([]*Permission, error)
	CreateAdminUser(ctx context.Context, username, passwordHash, name string) (*AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]*AdminUser, error)
	UpdateUserRoles(ctx context.Context, userID uint32, roleIDs []uint32) error
}

// RbacUsecase 负责 RBAC 管理的业务逻辑
type RbacUsecase struct {
	repo RbacRepo
}

func NewRbacUsecase(repo RbacRepo) *RbacUsecase {
	return &RbacUsecase{repo: repo}
}

func (uc *RbacUsecase) CreateAdminUser(ctx context.Context, username, password, name string) (*AdminUser, error) {
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return uc.repo.CreateAdminUser(ctx, username, string(hashedPassword), name)
}

// 其他 Usecase 方法大多直接调用 repo，可按需增加校验逻辑
// ...
