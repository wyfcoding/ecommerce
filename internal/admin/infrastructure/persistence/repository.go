package persistence

import (
	"context"
	"ecommerce/internal/admin/domain/entity"
	"ecommerce/internal/admin/domain/repository"

	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) repository.AdminRepository {
	return &adminRepository{db: db}
}

// Admin methods

func (r *adminRepository) CreateAdmin(ctx context.Context, admin *entity.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *adminRepository) GetAdminByID(ctx context.Context, id uint64) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Preload("Roles").Preload("Permissions").First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetAdminByUsername(ctx context.Context, username string) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Preload("Roles").Preload("Permissions").Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetAdminByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) UpdateAdmin(ctx context.Context, admin *entity.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *adminRepository) DeleteAdmin(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Admin{}, id).Error
}

func (r *adminRepository) ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error) {
	var admins []*entity.Admin
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&entity.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Preload("Roles").Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// Role methods

func (r *adminRepository) CreateRole(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *adminRepository) GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *adminRepository) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *adminRepository) UpdateRole(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *adminRepository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *adminRepository) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&entity.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// Permission methods

func (r *adminRepository) CreatePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *adminRepository) GetPermissionByID(ctx context.Context, id uint64) (*entity.Permission, error) {
	var permission entity.Permission
	if err := r.db.WithContext(ctx).First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *adminRepository) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var permission entity.Permission
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *adminRepository) UpdatePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *adminRepository) DeletePermission(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, id).Error
}

func (r *adminRepository) ListPermissions(ctx context.Context) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	if err := r.db.WithContext(ctx).Order("sort asc").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *adminRepository) GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]*entity.Permission, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// Association methods

func (r *adminRepository) AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Admin{Model: gorm.Model{ID: uint(adminID)}}).Association("Roles").Append(&entity.Role{Model: gorm.Model{ID: uint(roleID)}})
}

func (r *adminRepository) RemoveRoleFromAdmin(ctx context.Context, adminID, roleID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Admin{Model: gorm.Model{ID: uint(adminID)}}).Association("Roles").Delete(&entity.Role{Model: gorm.Model{ID: uint(roleID)}})
}

func (r *adminRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Role{Model: gorm.Model{ID: uint(roleID)}}).Association("Permissions").Append(&entity.Permission{Model: gorm.Model{ID: uint(permissionID)}})
}

func (r *adminRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Role{Model: gorm.Model{ID: uint(roleID)}}).Association("Permissions").Delete(&entity.Permission{Model: gorm.Model{ID: uint(permissionID)}})
}

// Log methods

func (r *adminRepository) CreateLoginLog(ctx context.Context, log *entity.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *adminRepository) ListLoginLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.LoginLog, int64, error) {
	var logs []*entity.LoginLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.WithContext(ctx).Model(&entity.LoginLog{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *adminRepository) CreateOperationLog(ctx context.Context, log *entity.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *adminRepository) ListOperationLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	var logs []*entity.OperationLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.WithContext(ctx).Model(&entity.OperationLog{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
