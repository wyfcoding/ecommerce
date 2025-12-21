package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 定义了数据持久层接口。
func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(ctx context.Context, user *domain.AdminUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *adminRepository) GetByID(ctx context.Context, id uint) (*domain.AdminUser, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到
		}
		return nil, err
	}
	return &user, nil
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *adminRepository) Update(ctx context.Context, user *domain.AdminUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *adminRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.AdminUser{}, id).Error
}

func (r *adminRepository) List(ctx context.Context, page, pageSize int) ([]*domain.AdminUser, int64, error) {
	var users []*domain.AdminUser
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&domain.AdminUser{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *adminRepository) AssignRole(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user domain.AdminUser
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		var roles []domain.Role
		if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}

		// GORM Replace 会替换掉当前的关联
		return tx.Model(&user).Association("Roles").Replace(roles)
	})
}

func (r *adminRepository) GetUserRoles(ctx context.Context, userID uint) ([]domain.Role, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// GetUserPermissions 获取用户的所有权限Code（去重）
func (r *adminRepository) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var user domain.AdminUser
	// 深度预加载：User -> Roles -> Permissions
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	permMap := make(map[string]bool)
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.Code] = true
		}
	}

	perms := make([]string, 0, len(permMap))
	for code := range permMap {
		perms = append(perms, code)
	}
	return perms, nil
}
