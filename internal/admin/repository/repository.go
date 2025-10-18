package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/admin/model"
	"ecommerce/pkg/hash"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// --- 接口定义 ---

// AdminUserRepo 定义了管理员用户数据访问接口。
type AdminUserRepo interface {
	// CreateAdminUser 创建一个新的管理员用户。
	CreateAdminUser(ctx context.Context, user *model.AdminUser) (*model.AdminUser, error)
	// GetAdminUserByID 根据ID获取管理员用户。
	GetAdminUserByID(ctx context.Context, id uint64) (*model.AdminUser, error)
	// GetAdminUserByUsername 根据用户名获取管理员用户。
	GetAdminUserByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	// UpdateAdminUser 更新管理员用户信息。
	UpdateAdminUser(ctx context.Context, user *model.AdminUser) (*model.AdminUser, error)
	// DeleteAdminUser 删除管理员用户。
	DeleteAdminUser(ctx context.Context, id uint64) error
	// ListAdminUsers 获取管理员用户列表。
	ListAdminUsers(ctx context.Context, usernameKeyword, emailKeyword string, pageSize, pageToken int32) ([]*model.AdminUser, int32, error)
	// VerifyAdminPassword 验证管理员用户名和密码。
	VerifyAdminPassword(ctx context.Context, username, password string) (*model.AdminUser, error)
}

// RoleRepo 定义了角色数据访问接口。
type RoleRepo interface {
	// CreateRole 创建一个新的角色。
	CreateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	// GetRoleByID 根据ID获取角色。
	GetRoleByID(ctx context.Context, id uint64) (*model.Role, error)
	// UpdateRole 更新角色信息。
	UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	// DeleteRole 删除角色。
	DeleteRole(ctx context.Context, id uint64) error
	// ListRoles 获取角色列表。
	ListRoles(ctx context.Context, nameKeyword string, pageSize, pageToken int32) ([]*model.Role, int32, error)
	// AssignRoleToAdminUser 为管理员用户分配角色。
	AssignRoleToAdminUser(ctx context.Context, adminUserID, roleID uint64) error
	// RemoveRoleFromAdminUser 从管理员用户移除角色。
	RemoveRoleFromAdminUser(ctx context.Context, adminUserID, roleID uint64) error
	// GetAdminUserRoles 获取管理员用户的所有角色。
	GetAdminUserRoles(ctx context.Context, adminUserID uint64) ([]*model.Role, error)
}

// PermissionRepo 定义了权限数据访问接口。
type PermissionRepo interface {
	// CreatePermission 创建一个新的权限。
	CreatePermission(ctx context.Context, perm *model.Permission) (*model.Permission, error)
	// GetPermissionByID 根据ID获取权限。
	GetPermissionByID(ctx context.Context, id uint64) (*model.Permission, error)
	// GetPermissionByName 根据名称获取权限。
	GetPermissionByName(ctx context.Context, name string) (*model.Permission, error)
	// UpdatePermission 更新权限信息。
	UpdatePermission(ctx context.Context, perm *model.Permission) (*model.Permission, error)
	// DeletePermission 删除权限。
	DeletePermission(ctx context.Context, id uint64) error
	// ListPermissions 获取权限列表。
	ListPermissions(ctx context.Context, nameKeyword string, pageSize, pageToken int32) ([]*model.Permission, int32, error)
	// AssignPermissionToRole 为角色分配权限。
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error
	// RemovePermissionFromRole 从角色移除权限。
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error
	// GetRolePermissions 获取角色的所有权限。
	GetRolePermissions(ctx context.Context, roleID uint64) ([]*model.Permission, error)
	// CheckAdminUserPermission 检查管理员用户是否拥有某个权限。
	CheckAdminUserPermission(ctx context.Context, adminUserID uint64, permissionName string) (bool, error)
}

// AuditLogRepo 定义了审计日志数据访问接口。
type AuditLogRepo interface {
	// CreateAuditLog 创建一条审计日志。
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
	// ListAuditLogs 获取审计日志列表。
	ListAuditLogs(ctx context.Context, adminUserID uint64, actionKeyword, entityType string, startTime, endTime *time.Time, pageSize, pageToken int32) ([]*model.AuditLog, int32, error)
	// GetAuditLogByID 根据ID获取审计日志。
	GetAuditLogByID(ctx context.Context, id uint64) (*model.AuditLog, error)
}

// --- 数据库模型 ---

// DBAdminUser 对应数据库中的管理员用户表。
type DBAdminUser struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null;type:varchar(64);comment:管理员用户名"`
	Password     string `gorm:"not null;type:varchar(255);comment:密码哈希值"`
	Nickname     string `gorm:"type:varchar(64);comment:昵称"`
	Email        string `gorm:"uniqueIndex;type:varchar(100);comment:邮箱"`
	Phone        string `gorm:"uniqueIndex;type:varchar(20);comment:手机号"`
	IsSuperAdmin bool   `gorm:"default:false;comment:是否是超级管理员"`
}

// TableName 返回 DBAdminUser 对应的表名。
func (DBAdminUser) TableName() string {
	return "admin_users"
}

// DBRole 对应数据库中的角色表。
type DBRole struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;type:varchar(64);comment:角色名称"`
	Description string `gorm:"type:varchar(255);comment:角色描述"`
}

// TableName 返回 DBRole 对应的表名。
func (DBRole) TableName() string {
	return "roles"
}

// DBPermission 对应数据库中的权限表。
type DBPermission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;type:varchar(128);comment:权限名称, 例如 product:create"`
	Description string `gorm:"type:varchar(255);comment:权限描述"`
}

// TableName 返回 DBPermission 对应的表名。
func (DBPermission) TableName() string {
	return "permissions"
}

// DBAdminUserRole 对应数据库中的管理员用户与角色关联表。
type DBAdminUserRole struct {
	AdminUserID uint64 `gorm:"primaryKey;comment:管理员用户ID"`
	RoleID      uint64 `gorm:"primaryKey;comment:角色ID"`
	CreatedAt   time.Time
}

// TableName 返回 DBAdminUserRole 对应的表名。
func (DBAdminUserRole) TableName() string {
	return "admin_user_roles"
}

// DBRolePermission 对应数据库中的角色与权限关联表。
type DBRolePermission struct {
	RoleID       uint64 `gorm:"primaryKey;comment:角色ID"`
	PermissionID uint64 `gorm:"primaryKey;comment:权限ID"`
	CreatedAt    time.Time
}

// TableName 返回 DBRolePermission 对应的表名。
func (DBRolePermission) TableName() string {
	return "role_permissions"
}

// DBAuditLog 对应数据库中的审计日志表。
type DBAuditLog struct {
	gorm.Model
	AdminUserID   uint64    `gorm:"index;comment:操作管理员用户ID"`
	AdminUsername string    `gorm:"type:varchar(64);comment:操作管理员用户名"`
	Action        string    `gorm:"type:varchar(128);not null;comment:操作类型"`
	EntityType    string    `gorm:"type:varchar(64);comment:操作实体类型"`
	EntityID      uint64    `gorm:"comment:操作实体ID"`
	Details       string    `gorm:"type:text;comment:操作详情 (JSON格式)"`
	IPAddress     string    `gorm:"type:varchar(45);comment:操作IP地址"`
}

// TableName 返回 DBAuditLog 对应的表名。
func (DBAuditLog) TableName() string {
	return "audit_logs"
}

// --- 数据层核心 ---

// Data 封装了所有数据库操作的 GORM 客户端。
type Data struct {
	db *gorm.DB
}

// NewData 创建一个新的 Data 实例，并执行数据库迁移。
func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}
	zap.S().Info("running database migrations for admin service...")
	// 自动迁移所有相关的数据库表
	if err := db.AutoMigrate(
		&DBAdminUser{},
		&DBRole{},
		&DBPermission{},
		&DBAdminUserRole{},
		&DBRolePermission{},
		&DBAuditLog{},
	); err != nil {
		zap.S().Errorf("failed to migrate admin database: %v", err)
		return nil, nil, fmt.Errorf("failed to migrate admin database: %w", err)
	}

	cleanup := func() {
		zap.S().Info("closing admin data layer...")
		// 可以在这里添加数据库连接关闭逻辑，如果 GORM 提供了的话
	}

	return d, cleanup, nil
}

// --- AdminUserRepo 实现 ---

// adminUserRepository 是 AdminUserRepo 接口的 GORM 实现。
type adminUserRepository struct {
	*Data
}

// NewAdminUserRepo 创建一个新的 AdminUserRepo 实例。
func NewAdminUserRepo(data *Data) AdminUserRepo {
	return &adminUserRepository{data}
}

// CreateAdminUser 在数据库中创建一个新的管理员用户。
func (r *adminUserRepository) CreateAdminUser(ctx context.Context, user *model.AdminUser) (*model.AdminUser, error) {
	// 对密码进行哈希处理
	hashedPassword, err := hash.HashPassword(user.Password)
	if err != nil {
		zap.S().Errorf("failed to hash password for admin user %s: %v", user.Username, err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	dbUser := fromBizAdminUser(user)
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		zap.S().Errorf("failed to create admin user %s in db: %v", user.Username, err)
		return nil, err
	}
	zap.S().Infof("admin user created in db: %d", dbUser.ID)
	return toBizAdminUser(dbUser), nil
}

// GetAdminUserByID 根据ID从数据库中获取管理员用户。
func (r *adminUserRepository) GetAdminUserByID(ctx context.Context, id uint64) (*model.AdminUser, error) {
	var dbUser DBAdminUser
	if err := r.db.WithContext(ctx).First(&dbUser, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到记录
		}
		zap.S().Errorf("failed to get admin user by id %d from db: %v", id, err)
		return nil, err
	}
	return toBizAdminUser(&dbUser), nil
}

// GetAdminUserByUsername 根据用户名从数据库中获取管理员用户。
func (r *adminUserRepository) GetAdminUserByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var dbUser DBAdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到记录
		}
		zap.S().Errorf("failed to get admin user by username %s from db: %v", username, err)
		return nil, err
	}
	return toBizAdminUser(&dbUser), nil
}

// UpdateAdminUser 更新数据库中的管理员用户信息。
func (r *adminUserRepository) UpdateAdminUser(ctx context.Context, user *model.AdminUser) (*model.AdminUser, error) {
	// 如果密码字段不为空，则重新哈希密码
	if user.Password != "" {
		hashedPassword, err := hash.HashPassword(user.Password)
		if err != nil {
			zap.S().Errorf("failed to hash new password for admin user %d: %v", user.ID, err)
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		user.Password = hashedPassword
	}

	dbUser := fromBizAdminUser(user)
	// 使用 Select 仅更新非零值字段，或者明确指定要更新的字段
	if err := r.db.WithContext(ctx).Model(&DBAdminUser{}).Where("id = ?", user.ID).Updates(dbUser).Error; err != nil {
		zap.S().Errorf("failed to update admin user %d in db: %v", user.ID, err)
		return nil, err
	}
	zap.S().Infof("admin user updated in db: %d", user.ID)
	return r.GetAdminUserByID(ctx, user.ID) // 返回更新后的完整用户数据
}

// DeleteAdminUser 从数据库中删除管理员用户。
func (r *adminUserRepository) DeleteAdminUser(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&DBAdminUser{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete admin user %d from db: %v", id, err)
		return err
	}
	zap.S().Infof("admin user deleted from db: %d", id)
	return nil
}

// ListAdminUsers 从数据库中获取管理员用户列表。
func (r *adminUserRepository) ListAdminUsers(ctx context.Context, usernameKeyword, emailKeyword string, pageSize, pageToken int32) ([]*model.AdminUser, int32, error) {
	var dbUsers []*DBAdminUser
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBAdminUser{})

	if usernameKeyword != "" {
		query = query.Where("username LIKE ?", "%"+usernameKeyword+"%")
	}
	if emailKeyword != "" {
		query = query.Where("email LIKE ?", "%"+emailKeyword+"%")
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("failed to count admin users: %v", err)
		return nil, 0, err
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10 // 默认每页大小
	}
	if pageToken <= 0 {
		pageToken = 1 // 默认页码
	}
	offset := (pageToken - 1) * pageSize

	if err := query.Limit(int(pageSize)).Offset(int(offset)).Find(&dbUsers).Error; err != nil {
		zap.S().Errorf("failed to list admin users from db: %v", err)
		return nil, 0, err
	}

	bizUsers := make([]*model.AdminUser, len(dbUsers))
	for i, dbUser := range dbUsers {
		bizUsers[i] = toBizAdminUser(dbUser)
	}

	return bizUsers, int32(totalCount), nil
}

// VerifyAdminPassword 验证管理员用户的用户名和密码。
func (r *adminUserRepository) VerifyAdminPassword(ctx context.Context, username, password string) (*model.AdminUser, error) {
	user, err := r.GetAdminUserByUsername(ctx, username)
	if err != nil {
		// GetAdminUserByUsername 已经记录了错误
		return nil, err
	}
	if user == nil {
		return nil, errors.New("admin user not found")
	}

	// 检查密码哈希
	if !hash.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}
	zap.S().Infof("admin user %s password verified successfully", username)
	return user, nil
}

// --- RoleRepo 实现 ---

// roleRepository 是 RoleRepo 接口的 GORM 实现。
type roleRepository struct {
	*Data
}

// NewRoleRepo 创建一个新的 RoleRepo 实例。
func NewRoleRepo(data *Data) RoleRepo {
	return &roleRepository{data}
}

// CreateRole 在数据库中创建一个新的角色。
func (r *roleRepository) CreateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	dbRole := fromBizRole(role)
	if err := r.db.WithContext(ctx).Create(dbRole).Error; err != nil {
		zap.S().Errorf("failed to create role %s in db: %v", role.Name, err)
		return nil, err
	}
	zap.S().Infof("role created in db: %d", dbRole.ID)
	return toBizRole(dbRole), nil
}

// GetRoleByID 根据ID从数据库中获取角色。
func (r *roleRepository) GetRoleByID(ctx context.Context, id uint64) (*model.Role, error) {
	var dbRole DBRole
	if err := r.db.WithContext(ctx).First(&dbRole, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get role by id %d from db: %v", id, err)
		return nil, err
	}
	return toBizRole(&dbRole), nil
}

// UpdateRole 更新数据库中的角色信息。
func (r *roleRepository) UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	dbRole := fromBizRole(role)
	if err := r.db.WithContext(ctx).Model(&DBRole{}).Where("id = ?", role.ID).Updates(dbRole).Error; err != nil {
		zap.S().Errorf("failed to update role %d in db: %v", role.ID, err)
		return nil, err
	}
	zap.S().Infof("role updated in db: %d", role.ID)
	return r.GetRoleByID(ctx, role.ID) // 返回更新后的完整角色数据
}

// DeleteRole 从数据库中删除角色。
func (r *roleRepository) DeleteRole(ctx context.Context, id uint64) error {
	// 事务性删除：先删除关联表中的记录，再删除角色本身
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除 admin_user_roles 中与此角色相关的记录
		if err := tx.Where("role_id = ?", id).Delete(&DBAdminUserRole{}).Error; err != nil {
			zap.S().Errorf("failed to delete admin_user_roles for role %d: %v", id, err)
			return err
		}
		// 删除 role_permissions 中与此角色相关的记录
		if err := tx.Where("role_id = ?", id).Delete(&DBRolePermission{}).Error; err != nil {
			zap.S().Errorf("failed to delete role_permissions for role %d: %v", id, err)
			return err
		}
		// 删除角色本身
		if err := tx.Delete(&DBRole{}, id).Error; err != nil {
			zap.S().Errorf("failed to delete role %d from db: %v", id, err)
			return err
		}
		zap.S().Infof("role deleted from db: %d", id)
		return nil
	})
}

// ListRoles 从数据库中获取角色列表。
func (r *roleRepository) ListRoles(ctx context.Context, nameKeyword string, pageSize, pageToken int32) ([]*model.Role, int32, error) {
	var dbRoles []*DBRole
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBRole{})

	if nameKeyword != "" {
		query = query.Where("name LIKE ?", "%"+nameKeyword+"%")
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("failed to count roles: %v", err)
		return nil, 0, err
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageToken <= 0 {
		pageToken = 1
	}
	offset := (pageToken - 1) * pageSize

	if err := query.Limit(int(pageSize)).Offset(int(offset)).Find(&dbRoles).Error; err != nil {
		zap.S().Errorf("failed to list roles from db: %v", err)
		return nil, 0, err
	}

	bizRoles := make([]*model.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		bizRoles[i] = toBizRole(dbRole)
	}

	return bizRoles, int32(totalCount), nil
}

// AssignRoleToAdminUser 为管理员用户分配角色。
func (r *roleRepository) AssignRoleToAdminUser(ctx context.Context, adminUserID, roleID uint64) error {
	adminUserRole := &DBAdminUserRole{AdminUserID: adminUserID, RoleID: roleID}
	// 使用 FirstOrCreate 避免重复插入
	if err := r.db.WithContext(ctx).FirstOrCreate(adminUserRole, adminUserRole).Error; err != nil {
		zap.S().Errorf("failed to assign role %d to admin user %d: %v", roleID, adminUserID, err)
		return err
	}
	zap.S().Infof("role %d assigned to admin user %d", roleID, adminUserID)
	return nil
}

// RemoveRoleFromAdminUser 从管理员用户移除角色。
func (r *roleRepository) RemoveRoleFromAdminUser(ctx context.Context, adminUserID, roleID uint64) error {
	if err := r.db.WithContext(ctx).Where("admin_user_id = ? AND role_id = ?", adminUserID, roleID).Delete(&DBAdminUserRole{}).Error; err != nil {
		zap.S().Errorf("failed to remove role %d from admin user %d: %v", roleID, adminUserID, err)
		return err
	}
	zap.S().Infof("role %d removed from admin user %d", roleID, adminUserID)
	return nil
}

// GetAdminUserRoles 获取管理员用户的所有角色。
func (r *roleRepository) GetAdminUserRoles(ctx context.Context, adminUserID uint64) ([]*model.Role, error) {
	var dbRoles []*DBRole
	// 通过 admin_user_roles 关联表查询角色
	if err := r.db.WithContext(ctx).Table("roles").
		Joins("JOIN admin_user_roles ON roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_user_id = ?", adminUserID).Find(&dbRoles).Error; err != nil {
		zap.S().Errorf("failed to get roles for admin user %d: %v", adminUserID, err)
		return nil, err
	}

	bizRoles := make([]*model.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		bizRoles[i] = toBizRole(dbRole)
	}
	return bizRoles, nil
}

// --- PermissionRepo 实现 ---

// permissionRepository 是 PermissionRepo 接口的 GORM 实现。
type permissionRepository struct {
	*Data
}

// NewPermissionRepo 创建一个新的 PermissionRepo 实例。
func NewPermissionRepo(data *Data) PermissionRepo {
	return &permissionRepository{data}
}

// CreatePermission 在数据库中创建一个新的权限。
func (r *permissionRepository) CreatePermission(ctx context.Context, perm *model.Permission) (*model.Permission, error) {
	dbPerm := fromBizPermission(perm)
	if err := r.db.WithContext(ctx).Create(dbPerm).Error; err != nil {
		zap.S().Errorf("failed to create permission %s in db: %v", perm.Name, err)
		return nil, err
	}
	zap.S().Infof("permission created in db: %d", dbPerm.ID)
	return toBizPermission(dbPerm), nil
}

// GetPermissionByID 根据ID从数据库中获取权限。
func (r *permissionRepository) GetPermissionByID(ctx context.Context, id uint64) (*model.Permission, error) {
	var dbPerm DBPermission
	if err := r.db.WithContext(ctx).First(&dbPerm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get permission by id %d from db: %v", id, err)
		return nil, err
	}
	return toBizPermission(&dbPerm), nil
}

// GetPermissionByName 根据名称从数据库中获取权限。
func (r *permissionRepository) GetPermissionByName(ctx context.Context, name string) (*model.Permission, error) {
	var dbPerm DBPermission
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbPerm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get permission by name %s from db: %v", name, err)
		return nil, err
	}
	return toBizPermission(&dbPerm), nil
}

// UpdatePermission 更新数据库中的权限信息。
func (r *permissionRepository) UpdatePermission(ctx context.Context, perm *model.Permission) (*model.Permission, error) {
	dbPerm := fromBizPermission(perm)
	if err := r.db.WithContext(ctx).Model(&DBPermission{}).Where("id = ?", perm.ID).Updates(dbPerm).Error; err != nil {
		zap.S().Errorf("failed to update permission %d in db: %v", perm.ID, err)
		return nil, err
	}
	zap.S().Infof("permission updated in db: %d", perm.ID)
	return r.GetPermissionByID(ctx, perm.ID) // 返回更新后的完整权限数据
}

// DeletePermission 从数据库中删除权限。
func (r *permissionRepository) DeletePermission(ctx context.Context, id uint64) error {
	// 事务性删除：先删除关联表中的记录，再删除权限本身
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除 role_permissions 中与此权限相关的记录
		if err := tx.Where("permission_id = ?", id).Delete(&DBRolePermission{}).Error; err != nil {
			zap.S().Errorf("failed to delete role_permissions for permission %d: %v", id, err)
			return err
		}
		// 删除权限本身
		if err := tx.Delete(&DBPermission{}, id).Error; err != nil {
			zap.S().Errorf("failed to delete permission %d from db: %v", id, err)
			return err
		}
		zap.S().Infof("permission deleted from db: %d", id)
		return nil
	})
}

// ListPermissions 从数据库中获取权限列表。
func (r *permissionRepository) ListPermissions(ctx context.Context, nameKeyword string, pageSize, pageToken int32) ([]*model.Permission, int32, error) {
	var dbPerms []*DBPermission
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBPermission{})

	if nameKeyword != "" {
		query = query.Where("name LIKE ?", "%"+nameKeyword+"%")
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("failed to count permissions: %v", err)
		return nil, 0, err
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageToken <= 0 {
		pageToken = 1
	}
	offset := (pageToken - 1) * pageSize

	if err := query.Limit(int(pageSize)).Offset(int(offset)).Find(&dbPerms).Error; err != nil {
		zap.S().Errorf("failed to list permissions from db: %v", err)
		return nil, 0, err
	}

	bizPerms := make([]*model.Permission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		bizPerms[i] = toBizPermission(dbPerm)
	}

	return bizPerms, int32(totalCount), nil
}

// AssignPermissionToRole 为角色分配权限。
func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	rolePermission := &DBRolePermission{RoleID: roleID, PermissionID: permissionID}
	// 使用 FirstOrCreate 避免重复插入
	if err := r.db.WithContext(ctx).FirstOrCreate(rolePermission, rolePermission).Error; err != nil {
		zap.S().Errorf("failed to assign permission %d to role %d: %v", permissionID, roleID, err)
		return err
	}
	zap.S().Infof("permission %d assigned to role %d", permissionID, roleID)
	return nil
}

// RemovePermissionFromRole 从角色移除权限。
func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error {
	if err := r.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&DBRolePermission{}).Error; err != nil {
		zap.S().Errorf("failed to remove permission %d from role %d: %v", permissionID, roleID, err)
		return err
	}
	zap.S().Infof("permission %d removed from role %d", permissionID, roleID)
	return nil
}

// GetRolePermissions 获取角色的所有权限。
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uint64) ([]*model.Permission, error) {
	var dbPerms []*DBPermission
	// 通过 role_permissions 关联表查询权限
	if err := r.db.WithContext(ctx).Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).Find(&dbPerms).Error; err != nil {
		zap.S().Errorf("failed to get permissions for role %d: %v", roleID, err)
		return nil, err
	}

	bizPerms := make([]*model.Permission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		bizPerms[i] = toBizPermission(dbPerm)
	}
	return bizPerms, nil
}

// CheckAdminUserPermission 检查管理员用户是否拥有某个权限。
// 这是一个复杂的业务逻辑，需要查询管理员用户的角色，然后查询这些角色拥有的权限。
func (r *permissionRepository) CheckAdminUserPermission(ctx context.Context, adminUserID uint64, permissionName string) (bool, error) {
	var count int64
	// 查询管理员用户是否直接拥有该权限 (通过角色)
	// 联结 admin_user_roles, role_permissions, permissions 表
	if err := r.db.WithContext(ctx).Table("admin_user_roles").
		Joins("JOIN role_permissions ON admin_user_roles.role_id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("admin_user_roles.admin_user_id = ? AND permissions.name = ?", adminUserID, permissionName).
		Count(&count).Error; err != nil {
		zap.S().Errorf("failed to check permission %s for admin user %d: %v", permissionName, adminUserID, err)
		return false, err
	}

	return count > 0, nil
}

// --- AuditLogRepo 实现 ---

// auditLogRepository 是 AuditLogRepo 接口的 GORM 实现。
type auditLogRepository struct {
	*Data
}

// NewAuditLogRepo 创建一个新的 AuditLogRepo 实例。
func NewAuditLogRepo(data *Data) AuditLogRepo {
	return &auditLogRepository{data}
}

// CreateAuditLog 在数据库中创建一条审计日志。
func (r *auditLogRepository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	dbLog := fromBizAuditLog(log)
	if err := r.db.WithContext(ctx).Create(dbLog).Error; err != nil {
		zap.S().Errorf("failed to create audit log for admin user %d, action %s: %v", log.AdminUserID, log.Action, err)
		return err
	}
	zap.S().Infof("audit log created: %s by admin user %d", log.Action, log.AdminUserID)
	return nil
}

// ListAuditLogs 从数据库中获取审计日志列表。
func (r *auditLogRepository) ListAuditLogs(ctx context.Context, adminUserID uint64, actionKeyword, entityType string, startTime, endTime *time.Time, pageSize, pageToken int32) ([]*model.AuditLog, int32, error) {
	var dbLogs []*DBAuditLog
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBAuditLog{})

	if adminUserID != 0 {
		query = query.Where("admin_user_id = ?", adminUserID)
	}
	if actionKeyword != "" {
		query = query.Where("action LIKE ?", "%"+actionKeyword+"%")
	}
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("failed to count audit logs: %v", err)
		return nil, 0, err
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageToken <= 0 {
		pageToken = 1
	}
	offset := (pageToken - 1) * pageSize

	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&dbLogs).Error; err != nil {
		zap.S().Errorf("failed to list audit logs from db: %v", err)
		return nil, 0, err
	}

	bizLogs := make([]*model.AuditLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		bizLogs[i] = toBizAuditLog(dbLog)
	}

	return bizLogs, int32(totalCount), nil
}

// GetAuditLogByID 根据ID从数据库中获取审计日志。
func (r *auditLogRepository) GetAuditLogByID(ctx context.Context, id uint64) (*model.AuditLog, error) {
	var dbLog DBAuditLog
	if err := r.db.WithContext(ctx).First(&dbLog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get audit log by id %d from db: %v", id, err)
		return nil, err
	}
	return toBizAuditLog(&dbLog), nil
}

// --- 模型转换辅助函数 ---

// toBizAdminUser 将 DBAdminUser 数据库模型转换为 model.AdminUser 业务领域模型。
func toBizAdminUser(dbUser *DBAdminUser) *model.AdminUser {
	if dbUser == nil {
		return nil
	}
	return &model.AdminUser{
		ID:           uint64(dbUser.ID),
		Username:     dbUser.Username,
		Password:     dbUser.Password,
		Nickname:     dbUser.Nickname,
		Email:        dbUser.Email,
		Phone:        dbUser.Phone,
		IsSuperAdmin: dbUser.IsSuperAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}
}

// fromBizAdminUser 将 model.AdminUser 业务领域模型转换为 DBAdminUser 数据库模型。
func fromBizAdminUser(bizUser *model.AdminUser) *DBAdminUser {
	if bizUser == nil {
		return nil
	}
	return &DBAdminUser{
		Model:        gorm.Model{ID: uint(bizUser.ID), CreatedAt: bizUser.CreatedAt, UpdatedAt: bizUser.UpdatedAt},
		Username:     bizUser.Username,
		Password:     bizUser.Password,
		Nickname:     bizUser.Nickname,
		Email:        bizUser.Email,
		Phone:        bizUser.Phone,
		IsSuperAdmin: bizUser.IsSuperAdmin,
	}
}

// toBizRole 将 DBRole 数据库模型转换为 model.Role 业务领域模型。
func toBizRole(dbRole *DBRole) *model.Role {
	if dbRole == nil {
		return nil
	}
	return &model.Role{
		ID:          uint64(dbRole.ID),
		Name:        dbRole.Name,
		Description: dbRole.Description,
		CreatedAt:   dbRole.CreatedAt,
		UpdatedAt:   dbRole.UpdatedAt,
	}
}

// fromBizRole 将 model.Role 业务领域模型转换为 DBRole 数据库模型。
func fromBizRole(bizRole *model.Role) *DBRole {
	if bizRole == nil {
		return nil
	}
	return &DBRole{
		Model:       gorm.Model{ID: uint(bizRole.ID), CreatedAt: bizRole.CreatedAt, UpdatedAt: bizRole.UpdatedAt},
		Name:        bizRole.Name,
		Description: bizRole.Description,
	}
}

// toBizPermission 将 DBPermission 数据库模型转换为 model.Permission 业务领域模型。
func toBizPermission(dbPerm *DBPermission) *model.Permission {
	if dbPerm == nil {
		return nil
	}
	return &model.Permission{
		ID:          uint64(dbPerm.ID),
		Name:        dbPerm.Name,
		Description: dbPerm.Description,
		CreatedAt:   dbPerm.CreatedAt,
		UpdatedAt:   dbPerm.UpdatedAt,
	}
}

// fromBizPermission 将 model.Permission 业务领域模型转换为 DBPermission 数据库模型。
func fromBizPermission(bizPerm *model.Permission) *DBPermission {
	if bizPerm == nil {
		return nil
	}
	return &DBPermission{
		Model:       gorm.Model{ID: uint(bizPerm.ID), CreatedAt: bizPerm.CreatedAt, UpdatedAt: bizPerm.UpdatedAt},
		Name:        bizPerm.Name,
		Description: bizPerm.Description,
	}
}

// toBizAuditLog 将 DBAuditLog 数据库模型转换为 model.AuditLog 业务领域模型。
func toBizAuditLog(dbLog *DBAuditLog) *model.AuditLog {
	if dbLog == nil {
		return nil
	}
	return &model.AuditLog{
		ID:            uint64(dbLog.ID),
		AdminUserID:   dbLog.AdminUserID,
		AdminUsername: dbLog.AdminUsername,
		Action:        dbLog.Action,
		EntityType:    dbLog.EntityType,
		EntityID:      dbLog.EntityID,
		Details:       dbLog.Details,
		IPAddress:     dbLog.IPAddress,
		CreatedAt:     dbLog.CreatedAt,
	}
}

// fromBizAuditLog 将 model.AuditLog 业务领域模型转换为 DBAuditLog 数据库模型。
func fromBizAuditLog(bizLog *model.AuditLog) *DBAuditLog {
	if bizLog == nil {
		return nil
	}
	return &DBAuditLog{
		Model:         gorm.Model{ID: uint(bizLog.ID), CreatedAt: bizLog.CreatedAt},
		AdminUserID:   bizLog.AdminUserID,
		AdminUsername: bizLog.AdminUsername,
		Action:        bizLog.Action,
		EntityType:    bizLog.EntityType,
		EntityID:      bizLog.EntityID,
		Details:       bizLog.Details,
		IPAddress:     bizLog.IPAddress,
	}
}