package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain/entity"     // 导入Admin模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/admin/domain/repository" // 导入Admin模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// adminRepository 是 AdminRepository 接口的GORM实现。
// 它负责将Admin模块的领域实体映射到数据库，并执行持久化操作。
type adminRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAdminRepository 创建并返回一个新的 adminRepository 实例。
// db: GORM数据库连接实例。
func NewAdminRepository(db *gorm.DB) repository.AdminRepository {
	return &adminRepository{db: db}
}

// --- Admin methods 管理员方法 ---

// CreateAdmin 在数据库中创建一个新的管理员记录。
func (r *adminRepository) CreateAdmin(ctx context.Context, admin *entity.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// GetAdminByID 根据ID从数据库获取管理员记录，并预加载其关联的角色和权限。
func (r *adminRepository) GetAdminByID(ctx context.Context, id uint64) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Preload("Roles").Preload("Permissions").First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAdminByUsername 根据用户名从数据库获取管理员记录，并预加载其关联的角色和权限。
func (r *adminRepository) GetAdminByUsername(ctx context.Context, username string) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Preload("Roles").Preload("Permissions").Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAdminByEmail 根据邮箱从数据库获取管理员记录。
func (r *adminRepository) GetAdminByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	var admin entity.Admin
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateAdmin 更新数据库中的管理员记录。
func (r *adminRepository) UpdateAdmin(ctx context.Context, admin *entity.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

// DeleteAdmin 根据ID从数据库删除管理员记录（软删除，因为使用了gorm.Model）。
func (r *adminRepository) DeleteAdmin(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Admin{}, id).Error
}

// ListAdmins 从数据库列出所有管理员记录，支持分页，并预加载关联的角色。
func (r *adminRepository) ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error) {
	var admins []*entity.Admin
	var total int64

	offset := (page - 1) * pageSize
	// 统计总记录数。
	if err := r.db.WithContext(ctx).Model(&entity.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据，并预加载关联的角色。
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Preload("Roles").Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// --- Role methods 角色方法 ---

// CreateRole 在数据库中创建一个新的角色记录。
func (r *adminRepository) CreateRole(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetRoleByID 根据ID从数据库获取角色记录，并预加载其关联的权限。
func (r *adminRepository) GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByCode 根据角色编码从数据库获取角色记录。
func (r *adminRepository) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// UpdateRole 更新数据库中的角色记录。
func (r *adminRepository) UpdateRole(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole 根据ID从数据库删除角色记录。
func (r *adminRepository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

// ListRoles 从数据库列出所有角色记录，支持分页。
func (r *adminRepository) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	offset := (page - 1) * pageSize
	// 统计总记录数。
	if err := r.db.WithContext(ctx).Model(&entity.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据。
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// --- Permission methods 权限方法 ---

// CreatePermission 在数据库中创建一个新的权限记录。
func (r *adminRepository) CreatePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetPermissionByID 根据ID从数据库获取权限记录。
func (r *adminRepository) GetPermissionByID(ctx context.Context, id uint64) (*entity.Permission, error) {
	var permission entity.Permission
	if err := r.db.WithContext(ctx).First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByCode 根据权限编码从数据库获取权限记录。
func (r *adminRepository) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var permission entity.Permission
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// UpdatePermission 更新数据库中的权限记录。
func (r *adminRepository) UpdatePermission(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// DeletePermission 根据ID从数据库删除权限记录。
func (r *adminRepository) DeletePermission(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, id).Error
}

// ListPermissions 从数据库列出所有权限记录，按排序字段升序排列。
func (r *adminRepository) ListPermissions(ctx context.Context) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	if err := r.db.WithContext(ctx).Order("sort asc").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionsByRoleID 根据角色ID获取该角色关联的所有权限记录。
// 它通过预加载Role实体的Permissions字段来实现。
func (r *adminRepository) GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]*entity.Permission, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// --- Association methods 关联方法 ---

// AssignRoleToAdmin 为指定的管理员分配一个角色。
// 通过GORM的Many2Many关联操作实现。
func (r *adminRepository) AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error {
	// 使用Model方法指定操作的主实体，然后使用Association方法处理多对多关系。
	return r.db.WithContext(ctx).Model(&entity.Admin{Model: gorm.Model{ID: uint(adminID)}}).Association("Roles").Append(&entity.Role{Model: gorm.Model{ID: uint(roleID)}})
}

// RemoveRoleFromAdmin 从指定的管理员移除一个角色。
func (r *adminRepository) RemoveRoleFromAdmin(ctx context.Context, adminID, roleID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Admin{Model: gorm.Model{ID: uint(adminID)}}).Association("Roles").Delete(&entity.Role{Model: gorm.Model{ID: uint(roleID)}})
}

// AssignPermissionToRole 为指定的角色分配一个权限。
func (r *adminRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Role{Model: gorm.Model{ID: uint(roleID)}}).Association("Permissions").Append(&entity.Permission{Model: gorm.Model{ID: uint(permissionID)}})
}

// RemovePermissionFromRole 从指定的角色移除一个权限。
func (r *adminRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Role{Model: gorm.Model{ID: uint(roleID)}}).Association("Permissions").Delete(&entity.Permission{Model: gorm.Model{ID: uint(permissionID)}})
}

// --- Log methods 日志方法 ---

// CreateLoginLog 在数据库中创建一条新的登录日志记录。
func (r *adminRepository) CreateLoginLog(ctx context.Context, log *entity.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListLoginLogs 列出指定管理员的登录日志，支持分页，按创建时间降序排列。
func (r *adminRepository) ListLoginLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.LoginLog, int64, error) {
	var logs []*entity.LoginLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.WithContext(ctx).Model(&entity.LoginLog{})
	if adminID > 0 { // 如果提供了adminID，则按adminID过滤。
		query = query.Where("admin_id = ?", adminID)
	}

	// 统计总记录数。
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据。
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CreateOperationLog 在数据库中创建一条新的操作日志记录。
func (r *adminRepository) CreateOperationLog(ctx context.Context, log *entity.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListOperationLogs 列出指定管理员的操作日志，支持分页，按创建时间降序排列。
func (r *adminRepository) ListOperationLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	var logs []*entity.OperationLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.WithContext(ctx).Model(&entity.OperationLog{})
	if adminID > 0 { // 如果提供了adminID，则按adminID过滤。
		query = query.Where("admin_id = ?", adminID)
	}

	// 统计总记录数。
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据。
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// --- SystemSetting methods 系统配置方法 ---

// GetSystemSetting 根据键获取系统配置。
func (r *adminRepository) GetSystemSetting(ctx context.Context, key string) (*entity.SystemSetting, error) {
	var setting entity.SystemSetting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// SaveSystemSetting 保存系统配置。
func (r *adminRepository) SaveSystemSetting(ctx context.Context, setting *entity.SystemSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}
