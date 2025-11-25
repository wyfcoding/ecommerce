package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/admin/domain/entity"
)

// AdminRepository 管理员仓储接口
type AdminRepository interface {
	// Admin methods
	CreateAdmin(ctx context.Context, admin *entity.Admin) error
	GetAdminByID(ctx context.Context, id uint64) (*entity.Admin, error)
	GetAdminByUsername(ctx context.Context, username string) (*entity.Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (*entity.Admin, error)
	UpdateAdmin(ctx context.Context, admin *entity.Admin) error
	DeleteAdmin(ctx context.Context, id uint64) error
	ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error)

	// Role methods
	CreateRole(ctx context.Context, role *entity.Role) error
	GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error)
	GetRoleByCode(ctx context.Context, code string) (*entity.Role, error)
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id uint64) error
	ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error)

	// Permission methods
	CreatePermission(ctx context.Context, permission *entity.Permission) error
	GetPermissionByID(ctx context.Context, id uint64) (*entity.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	UpdatePermission(ctx context.Context, permission *entity.Permission) error
	DeletePermission(ctx context.Context, id uint64) error
	ListPermissions(ctx context.Context) ([]*entity.Permission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]*entity.Permission, error)

	// Association methods
	AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error
	RemoveRoleFromAdmin(ctx context.Context, adminID, roleID uint64) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error

	// Log methods
	CreateLoginLog(ctx context.Context, log *entity.LoginLog) error
	ListLoginLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.LoginLog, int64, error)
	CreateOperationLog(ctx context.Context, log *entity.OperationLog) error
	ListOperationLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*entity.OperationLog, int64, error)
}
