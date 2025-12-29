package domain

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 代表后台管理员用户
// 拥有角色，通过角色获得权限
type AdminUser struct {
	gorm.Model
	Username     string     `gorm:"column:username;type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希"`
	Email        string     `gorm:"column:email;type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	FullName     string     `gorm:"column:full_name;type:varchar(100);comment:全名"`
	Status       UserStatus `gorm:"column:status;type:tinyint;default:1;comment:状态 1:启用 2:禁用"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at;comment:最后登录时间"`

	// 多对多关联角色
	Roles []Role `gorm:"many2many:admin_user_roles;"`
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
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(50);uniqueIndex;not null;comment:角色名称"`
	Code        string `gorm:"column:code;type:varchar(50);uniqueIndex;not null;comment:角色编码(如 SUPER_ADMIN)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`

	// 多对多关联权限
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission 代表具体的权限点
// 通常对应某个资源的某个操作，如 order:view, product:edit
type Permission struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:权限名称"`
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null;comment:权限编码(resource:action)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
	Resource    string `gorm:"column:resource;type:varchar(50);index;comment:资源类型"`
	Action      string `gorm:"column:action;type:varchar(50);comment:操作类型"`
	Type        string `gorm:"column:type;type:varchar(20);default:'api';comment:权限类型(menu/api/button)"`
	ParentID    uint   `gorm:"column:parent_id;default:0;comment:父权限ID"`
}

// ApprovalRequest 审批申请
// 针对高风险操作（如强制退款、系统配置变更）需要经过审批流程
type ApprovalRequest struct {
	gorm.Model
	RequesterID uint   `gorm:"column:requester_id;index;not null;comment:申请人ID"`
	ActionType  string `gorm:"column:action_type;type:varchar(50);index;not null;comment:申请动作类型"`
	Description string `gorm:"column:description;type:varchar(255);comment:申请描述/理由"`
	Payload     string `gorm:"column:payload;type:text;comment:操作数据快照(JSON)"`

	Status      ApprovalStatus `gorm:"column:status;type:tinyint;default:1;comment:审批状态"`
	CurrentStep int            `gorm:"column:current_step;type:int;default:1;comment:当前审批步骤"`
	TotalSteps  int            `gorm:"column:total_steps;type:int;default:1;comment:总步骤数"`

	ApproverRole string `gorm:"column:approver_role;type:varchar(50);comment:当前需要的审批角色Code"`

	FinalizedAt *time.Time `gorm:"column:finalized_at;comment:流程结束时间"`

	// 审批记录
	Logs []ApprovalLog `gorm:"foreignKey:RequestID"`
}

// ApprovalStatus 结构体定义。
type ApprovalStatus int

const (
	ApprovalStatusPending  ApprovalStatus = 1 // 待审批
	ApprovalStatusApproved ApprovalStatus = 2 // 已通过
	ApprovalStatusRejected ApprovalStatus = 3 // 已拒绝
	ApprovalStatusCanceled ApprovalStatus = 4 // 已取消
)

// ApprovalLog 单次审批操作记录
type ApprovalLog struct {
	gorm.Model
	RequestID    uint           `gorm:"column:request_id;index;not null;comment:关联申请ID"`
	ApproverID   uint           `gorm:"column:approver_id;not null;comment:审批人ID"`
	ApproverName string         `gorm:"column:approver_name;type:varchar(50);comment:审批人姓名"`
	Action       ApprovalAction `gorm:"column:action;type:tinyint;not null;comment:动作 1:通过 2:拒绝"`
	Comment      string         `gorm:"column:comment;type:varchar(255);comment:审批意见"`
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
	gorm.Model
	UserID    uint   `gorm:"column:user_id;index;not null;comment:操作人ID"`
	Username  string `gorm:"column:username;type:varchar(50);not null;comment:操作人用户名(冗余)"`
	Action    string `gorm:"column:action;type:varchar(50);index;not null;comment:操作动作"`
	Resource  string `gorm:"column:resource;type:varchar(50);index;not null;comment:资源类型"`
	TargetID  string `gorm:"column:target_id;type:varchar(50);index;comment:目标资源ID"`
	ClientIP  string `gorm:"column:client_ip;type:varchar(50);comment:客户端IP"`
	Payload   string `gorm:"column:payload;type:text;comment:请求参数(JSON)"`
	Result    string `gorm:"column:result;type:text;comment:操作结果/错误信息"`
	Status    int    `gorm:"column:status;type:tinyint;default:1;comment:结果状态 1:成功 0:失败"`
	UserAgent string `gorm:"column:user_agent;type:varchar(255);comment:UserAgent"`
}

// SystemSetting 系统配置
type SystemSetting struct {
	gorm.Model
	Key         string `gorm:"column:key;type:varchar(100);uniqueIndex;not null;comment:配置键"`
	Value       string `gorm:"column:value;type:text;comment:配置值"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
}
