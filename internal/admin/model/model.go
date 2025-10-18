package model

import "time"

// AdminUser 是管理员用户的业务领域模型。
type AdminUser struct {
	ID        uint64
	Username  string
	Password  string // 存储哈希后的密码
	Email     string
	Nickname  string
	IsActive  bool
	RoleIDs   []uint64 // 关联的角色ID列表
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Role 是管理员角色的业务领域模型。
type Role struct {
	ID          uint64
	Name        string
	Description string
	Permissions []string // 关联的权限名称列表 (e.g., "product:read")
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission 是权限的业务领域模型。
type Permission struct {
	ID          uint64
	Name        string // 权限名称 (e.g., "product:read", "user:manage")
	Description string
}

// AuditLog 是审计日志的业务领域模型。
type AuditLog struct {
	ID            uint64
	AdminUserID   uint64
	AdminUsername string
	Action        string // 操作类型 (e.g., "CREATE_PRODUCT", "UPDATE_USER")
	EntityType    string // 被操作实体类型 (e.g., "Product", "User")
	EntityID      uint64 // 被操作实体ID
	Details       string // 操作详情 (JSON string)
	IPAddress     string // 操作IP地址
	CreatedAt     time.Time
}

// SystemSetting 是系统配置项的业务领域模型。
type SystemSetting struct {
	Key         string
	Value       string
	Description string
	UpdatedAt   time.Time
}
