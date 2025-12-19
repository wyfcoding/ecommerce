package domain

import "context"

// AdminRepository 用户管理仓储
type AdminRepository interface {
	Create(ctx context.Context, user *AdminUser) error
	GetByID(ctx context.Context, id uint) (*AdminUser, error)
	GetByUsername(ctx context.Context, username string) (*AdminUser, error)
	Update(ctx context.Context, user *AdminUser) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int) ([]*AdminUser, int64, error)

	// 角色关联
	AssignRole(ctx context.Context, userID uint, roleIDs []uint) error
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error) // 获取用户所有权限Code
}

// RoleRepository 角色与权限仓储
type RoleRepository interface {
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByID(ctx context.Context, id uint) (*Role, error)
	GetRoleByCode(ctx context.Context, code string) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uint) error

	// 权限管理
	CreatePermission(ctx context.Context, perm *Permission) error
	GetPermissionByID(ctx context.Context, id uint) (*Permission, error)
	ListPermissions(ctx context.Context) ([]*Permission, error)
	AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
}

// AuditRepository 审计日志仓储
type AuditRepository interface {
	Save(ctx context.Context, log *AuditLog) error
	Find(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*AuditLog, int64, error)
}

// ApprovalRepository 审批流程仓储
type ApprovalRepository interface {
	CreateRequest(ctx context.Context, req *ApprovalRequest) error
	GetRequestByID(ctx context.Context, id uint) (*ApprovalRequest, error)
	UpdateRequest(ctx context.Context, req *ApprovalRequest) error
	ListPendingRequests(ctx context.Context, roleLimit string) ([]*ApprovalRequest, error) // 根据角色获取待办

	AddLog(ctx context.Context, log *ApprovalLog) error
}

// SettingRepository 系统配置仓储
type SettingRepository interface {
	GetByKey(ctx context.Context, key string) (*SystemSetting, error)
	Save(ctx context.Context, setting *SystemSetting) error
}
