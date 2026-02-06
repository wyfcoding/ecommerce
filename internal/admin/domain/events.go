package domain

import "time"

const (
	AdminUserCreatedEventType        = "admin.user.created"
	AdminUserUpdatedEventType        = "admin.user.updated"
	AdminUserDisabledEventType       = "admin.user.disabled"
	RoleAssignedEventType            = "admin.user.role.assigned"
	RoleCreatedEventType             = "admin.role.created"
	RoleUpdatedEventType             = "admin.role.updated"
	RoleDeletedEventType             = "admin.role.deleted"
	PermissionCreatedEventType       = "admin.permission.created"
	PermissionUpdatedEventType       = "admin.permission.updated"
	PermissionDeletedEventType       = "admin.permission.deleted"
	ApprovalRequestCreatedEventType  = "admin.approval.request.created"
	ApprovalRequestApprovedEventType = "admin.approval.request.approved"
	ApprovalRequestRejectedEventType = "admin.approval.request.rejected"
	SystemSettingUpdatedEventType    = "admin.setting.updated"
	AuditLogCreatedEventType         = "admin.audit.log.created"
)

// AdminUserCreatedEvent 管理员创建事件。
type AdminUserCreatedEvent struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

// AdminUserUpdatedEvent 管理员信息更新事件。
type AdminUserUpdatedEvent struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

// AdminUserDisabledEvent 管理员禁用事件。
type AdminUserDisabledEvent struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// RoleAssignedEvent 角色分配事件。
type RoleAssignedEvent struct {
	UserID    uint      `json:"user_id"`
	RoleIDs   []uint    `json:"role_ids"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

// RoleCreatedEvent 角色创建事件。
type RoleCreatedEvent struct {
	RoleID    uint      `json:"role_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

// RoleUpdatedEvent 角色更新事件。
type RoleUpdatedEvent struct {
	RoleID    uint      `json:"role_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

// RoleDeletedEvent 角色删除事件。
type RoleDeletedEvent struct {
	RoleID    uint      `json:"role_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PermissionCreatedEvent 权限创建事件。
type PermissionCreatedEvent struct {
	PermissionID uint      `json:"permission_id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Timestamp    time.Time `json:"timestamp"`
}

// PermissionUpdatedEvent 权限更新事件。
type PermissionUpdatedEvent struct {
	PermissionID uint      `json:"permission_id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Timestamp    time.Time `json:"timestamp"`
}

// PermissionDeletedEvent 权限删除事件。
type PermissionDeletedEvent struct {
	PermissionID uint      `json:"permission_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// ApprovalRequestCreatedEvent 审批申请创建事件。
type ApprovalRequestCreatedEvent struct {
	RequestID   uint      `json:"request_id"`
	RequesterID uint      `json:"requester_id"`
	ActionType  string    `json:"action_type"`
	Timestamp   time.Time `json:"timestamp"`
}

// ApprovalRequestApprovedEvent 审批通过事件。
type ApprovalRequestApprovedEvent struct {
	RequestID  uint      `json:"request_id"`
	ApproverID uint      `json:"approver_id"`
	Comment    string    `json:"comment"`
	Timestamp  time.Time `json:"timestamp"`
}

// ApprovalRequestRejectedEvent 审批拒绝事件。
type ApprovalRequestRejectedEvent struct {
	RequestID  uint      `json:"request_id"`
	ApproverID uint      `json:"approver_id"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

// SystemSettingUpdatedEvent 系统配置更新事件。
type SystemSettingUpdatedEvent struct {
	Key       string    `json:"key"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditLogCreatedEvent 审计日志创建事件。
type AuditLogCreatedEvent struct {
	LogID     uint      `json:"log_id"`
	UserID    uint      `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Timestamp time.Time `json:"timestamp"`
}
