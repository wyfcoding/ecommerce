package domain

import (
	"time"

	"gorm.io/gorm"
)

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
