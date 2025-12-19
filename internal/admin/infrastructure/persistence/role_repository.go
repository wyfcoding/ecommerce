package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

// Role CRUD

func (r *roleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetRoleByCode(ctx context.Context, code string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	var roles []*domain.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) DeleteRole(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error
}

// Permission CRUD

func (r *roleRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *roleRepository) GetPermissionByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var perm domain.Permission
	if err := r.db.WithContext(ctx).First(&perm, id).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *roleRepository) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	var perms []*domain.Permission
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role domain.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}

		var perms []domain.Permission
		if err := tx.Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
			return err
		}

		return tx.Model(&role).Association("Permissions").Replace(perms)
	})
}
