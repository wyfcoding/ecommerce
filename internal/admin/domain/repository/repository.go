package repository

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain/entity" // 导入领域实体定义。
)

// AdminRepository 是管理员模块的仓储接口。
// 它定义了对管理员、角色、权限以及相关日志进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AdminRepository interface {
	// --- Admin methods ---

	// CreateAdmin 在数据存储中创建一个新的管理员实体。
	// ctx: 上下文。
	// admin: 待创建的管理员实体。
	CreateAdmin(ctx context.Context, admin *entity.Admin) error
	// GetAdminByID 根据ID获取管理员实体。
	GetAdminByID(ctx context.Context, id uint64) (*entity.Admin, error)
	// GetAdminByUsername 根据用户名获取管理员实体。
	GetAdminByUsername(ctx context.Context, username string) (*entity.Admin, error)
	// GetAdminByEmail 根据邮箱获取管理员实体。
	GetAdminByEmail(ctx context.Context, email string) (*entity.Admin, error)
	// UpdateAdmin 更新管理员实体的信息。
	UpdateAdmin(ctx context.Context, admin *entity.Admin) error
	// DeleteAdmin 根据ID删除管理员实体。
	DeleteAdmin(ctx context.Context, id uint64) error
	// ListAdmins 列出所有管理员实体，支持分页。
	ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error)

	// --- Role methods ---

	// CreateRole 在数据存储中创建一个新的角色实体。
	CreateRole(ctx context.Context, role *entity.Role) error
	// GetRoleByID 根据ID获取角色实体。
	GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error)
	// GetRoleByCode 根据角色编码获取角色实体。
	GetRoleByCode(ctx context.Context, code string) (*entity.Role, error)
	// UpdateRole 更新角色实体的信息。
	UpdateRole(ctx context.Context, role *entity.Role) error
	// DeleteRole 根据ID删除角色实体。
	DeleteRole(ctx context.Context, id uint64) error
	// ListRoles 列出所有角色实体，支持分页。
	ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)

	// --- Permission methods ---

	// CreatePermission 在数据存储中创建一个新的权限实体。
	CreatePermission(ctx context.Context, permission *entity.Permission) error
	// GetPermissionByID 根据ID获取权限实体。
	GetPermissionByID(ctx context.Context, id uint64) (*entity.Permission, error)
	// GetPermissionByCode 根据权限编码获取权限实体。
	GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	// UpdatePermission 更新权限实体的信息。
	UpdatePermission(ctx context.Context, permission *entity.Permission) error
	// DeletePermission 根据ID删除权限实体。
	DeletePermission(ctx context.Context, id uint64) error
	// ListPermissions 列出所有权限实体。
	ListPermissions(ctx context.Context) ([]*entity.Permission, error)
	// GetPermissionsByRoleID 根据角色ID获取该角色拥有的所有权限实体。
	GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]*entity.Permission, error)

	// --- Association methods ---
	// 管理员与角色、角色与权限之间的关联关系。

	// AssignRoleToAdmin 为指定的管理员分配一个角色。
	AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error
	// RemoveRoleFromAdmin 从指定的管理员移除一个角色。
	RemoveRoleFromAdmin(ctx context.Context, adminID, roleID uint64) error
	// AssignPermissionToRole 为指定的角色分配一个权限。
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error
	// RemovePermissionFromRole 从指定的角色移除一个权限。
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error

	// --- Log methods ---
	// 管理员登录日志和操作日志的持久化。

	// CreateLoginLog 创建一条新的登录日志记录。
	CreateLoginLog(ctx context.Context, log *entity.LoginLog) error
	// ListLoginLogs 列出指定管理员的登录日志，支持分页。
	ListLoginLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.LoginLog, int64, error)
	// CreateOperationLog 创建一条新的操作日志记录。
	CreateOperationLog(ctx context.Context, log *entity.OperationLog) error
	// ListOperationLogs 列出指定管理员的操作日志，支持分页。
	ListOperationLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.OperationLog, int64, error)

	// --- SystemSetting methods ---

	// GetSystemSetting 获取系统设置。
	GetSystemSetting(ctx context.Context, key string) (*entity.SystemSetting, error)
	// SaveSystemSetting 保存系统设置。
	SaveSystemSetting(ctx context.Context, setting *entity.SystemSetting) error
}
