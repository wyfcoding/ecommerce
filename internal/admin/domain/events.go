package domain

import "time"

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
