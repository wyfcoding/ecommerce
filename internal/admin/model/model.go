package model

import "time"

// AdminUser 是管理员用户的业务领域模型。
// 包含了管理员用户的基本信息、认证凭据和状态。
type AdminUser struct {
	ID        uint64    // 管理员用户唯一标识符
	Username  string    // 登录用户名，唯一
	Password  string    // 存储哈希后的密码，不应直接暴露
	Email     string    // 邮箱地址，唯一
	Nickname  string    // 昵称
	IsActive  bool      // 账户是否激活
	Status    int32     // 用户状态 (例如：0-禁用, 1-正常)
	RoleIDs   []uint64  // 关联的角色ID列表，用于权限管理
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 最后更新时间
}

// Role 是管理员角色的业务领域模型。
// 定义了角色的名称、描述及其关联的权限。
type Role struct {
	ID          uint64    // 角色唯一标识符
	Name        string    // 角色名称 (例如："产品经理", "运营")，唯一
	Description string    // 角色描述
	Permissions []string  // 关联的权限名称列表 (e.g., "product:read", "order:manage")
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 最后更新时间
}

// Permission 是权限的业务领域模型。
// 定义了具体的权限点，例如对某个资源的读写操作。
type Permission struct {
	ID          uint64 // 权限唯一标识符
	Name        string // 权限名称 (e.g., "product:read", "user:manage")，唯一
	Description string // 权限描述
}

// AuditLog 是审计日志的业务领域模型。
// 记录了管理员用户在系统中执行的各项操作，用于追踪和审计。
type AuditLog struct {
	ID            uint64    // 审计日志唯一标识符
	AdminUserID   uint64    // 执行操作的管理员用户ID
	AdminUsername string    // 执行操作的管理员用户名
	Action        string    // 操作类型 (e.g., "CREATE_PRODUCT", "UPDATE_USER_STATUS")
	EntityType    string    // 被操作实体类型 (e.g., "Product", "User", "Order")
	EntityID      uint64    // 被操作实体ID
	Details       string    // 操作详情 (通常为 JSON 格式的变更内容或参数)
	IPAddress     string    // 操作发起方的 IP 地址
	CreatedAt     time.Time // 操作发生时间
}

// SystemSetting 是系统配置项的业务领域模型。
// 用于存储可由管理员动态配置的系统参数。
type SystemSetting struct {
	Key         string    // 配置项的键，唯一
	Value       string    // 配置项的值
	Description string    // 配置项的描述
	UpdatedAt   time.Time // 最后更新时间
}