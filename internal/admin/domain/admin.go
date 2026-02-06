package domain

import "time"

// AdminUser 代表后台管理员用户
// 拥有角色，通过角色获得权限
type AdminUser struct {
	ID           uint       // 主键ID
	CreatedAt    time.Time  // 创建时间
	UpdatedAt    time.Time  // 更新时间
	Username     string     // 用户名
	PasswordHash string     // 密码哈希
	Email        string     // 邮箱
	FullName     string     // 全名
	Status       UserStatus // 状态 1:启用 2:禁用
	LastLoginAt  *time.Time // 最后登录时间

	// 多对多关联角色
	Roles []Role
}

// UserStatus 结构体定义。
type UserStatus int

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
)

// Role 代表角色
// 角色是一组权限的集合
type Role struct {
	ID          uint      // 主键ID
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
	Name        string    // 角色名称
	Code        string    // 角色编码(如 SUPER_ADMIN)
	Description string    // 描述

	// 多对多关联权限
	Permissions []Permission
}

// Permission 代表具体的权限点
// 通常对应某个资源的某个操作，如 order:view, product:edit
type Permission struct {
	ID          uint      // 主键ID
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
	Name        string    // 权限名称
	Code        string    // 权限编码(resource:action)
	Description string    // 描述
	Resource    string    // 资源类型
	Action      string    // 操作类型
	Type        string    // 权限类型(menu/api/button)
	ParentID    uint      // 父权限ID
}

// ApprovalRequest 审批申请
// 针对高风险操作（如强制退款、系统配置变更）需要经过审批流程
type ApprovalRequest struct {
	ID            uint           // 主键ID
	CreatedAt     time.Time      // 创建时间
	UpdatedAt     time.Time      // 更新时间
	RequesterID   uint           // 申请人ID
	ActionType    string         // 申请动作类型
	Description   string         // 申请描述/理由
	Payload       string         // 操作数据快照(JSON)
	Status        ApprovalStatus // 审批状态
	CurrentStep   int            // 当前审批步骤
	TotalSteps    int            // 总步骤数
	ApproverRole  string         // 当前需要的审批角色Code
	FinalizedAt   *time.Time     // 流程结束时间
	FailureReason string         // 执行失败原因
	RetryCount    int            // 重试次数

	// 审批记录
	Logs []ApprovalLog
}

// ApprovalStatus 结构体定义。
type ApprovalStatus int

const (
	ApprovalStatusPending  ApprovalStatus = 1 // 待审批
	ApprovalStatusApproved ApprovalStatus = 2 // 已通过
	ApprovalStatusRejected ApprovalStatus = 3 // 已拒绝
	ApprovalStatusCanceled ApprovalStatus = 4 // 已取消
	ApprovalStatusFailed   ApprovalStatus = 5 // 执行失败
)

// ApprovalLog 单次审批操作记录
type ApprovalLog struct {
	ID           uint           // 主键ID
	CreatedAt    time.Time      // 创建时间
	UpdatedAt    time.Time      // 更新时间
	RequestID    uint           // 关联申请ID
	ApproverID   uint           // 审批人ID
	ApproverName string         // 审批人姓名
	Action       ApprovalAction // 动作 1:通过 2:拒绝
	Comment      string         // 审批意见
}

// ApprovalAction 结构体定义。
type ApprovalAction int

const (
	ApprovalActionApprove ApprovalAction = 1
	ApprovalActionReject  ApprovalAction = 2
)

// AuditLog 审计日志
// 记录所有管理端的操作行为，不可变
type AuditLog struct {
	ID        uint      // 主键ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
	UserID    uint      // 操作人ID
	Username  string    // 操作人用户名(冗余)
	Action    string    // 操作动作
	Resource  string    // 资源类型
	TargetID  string    // 目标资源ID
	ClientIP  string    // 客户端IP
	Payload   string    // 请求参数(JSON)
	Result    string    // 操作结果/错误信息
	Status    int       // 结果状态 1:成功 0:失败
	UserAgent string    // UserAgent
}

// SystemSetting 系统配置
type SystemSetting struct {
	ID          uint      // 主键ID
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
	Key         string    // 配置键
	Value       string    // 配置值
	Description string    // 描述
}
