package entity

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = 1 // 待处理
	TaskStatusRunning   TaskStatus = 2 // 运行中
	TaskStatusCompleted TaskStatus = 3 // 已完成
	TaskStatusFailed    TaskStatus = 4 // 失败
	TaskStatusCancelled TaskStatus = 5 // 已取消
)

// ProcessingTask 数据处理任务实体
type ProcessingTask struct {
	gorm.Model
	Name         string     `gorm:"type:varchar(255);not null;comment:任务名称" json:"name"`
	Type         string     `gorm:"type:varchar(64);not null;comment:任务类型" json:"type"`
	Status       TaskStatus `gorm:"default:1;comment:状态" json:"status"`
	Config       string     `gorm:"type:text;comment:配置(JSON)" json:"config"`
	Result       string     `gorm:"type:text;comment:结果(JSON)" json:"result"`
	ErrorMessage string     `gorm:"type:text;comment:错误信息" json:"error_message"`
	StartTime    *time.Time `gorm:"comment:开始时间" json:"start_time"`
	EndTime      *time.Time `gorm:"comment:结束时间" json:"end_time"`
	WorkflowID   uint64     `gorm:"index;comment:所属工作流ID" json:"workflow_id"`
}

// WorkflowStatus 工作流状态
type WorkflowStatus int

const (
	WorkflowStatusActive   WorkflowStatus = 1 // 激活
	WorkflowStatusInactive WorkflowStatus = 2 // 停用
	WorkflowStatusDraft    WorkflowStatus = 3 // 草稿
)

// ProcessingWorkflow 数据处理工作流实体
type ProcessingWorkflow struct {
	gorm.Model
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null;comment:工作流名称" json:"name"`
	Description string         `gorm:"type:text;comment:描述" json:"description"`
	Steps       string         `gorm:"type:text;comment:步骤定义(JSON)" json:"steps"`
	Status      WorkflowStatus `gorm:"default:3;comment:状态" json:"status"`
}

// NewProcessingTask 创建任务
func NewProcessingTask(name, taskType, config string, workflowID uint64) *ProcessingTask {
	return &ProcessingTask{
		Name:       name,
		Type:       taskType,
		Config:     config,
		WorkflowID: workflowID,
		Status:     TaskStatusPending,
	}
}

// Start 开始任务
func (t *ProcessingTask) Start() {
	t.Status = TaskStatusRunning
	now := time.Now()
	t.StartTime = &now
}

// Complete 完成任务
func (t *ProcessingTask) Complete(result string) {
	t.Status = TaskStatusCompleted
	t.Result = result
	now := time.Now()
	t.EndTime = &now
}

// Fail 任务失败
func (t *ProcessingTask) Fail(errorMessage string) {
	t.Status = TaskStatusFailed
	t.ErrorMessage = errorMessage
	now := time.Now()
	t.EndTime = &now
}

// NewProcessingWorkflow 创建工作流
func NewProcessingWorkflow(name, description, steps string) *ProcessingWorkflow {
	return &ProcessingWorkflow{
		Name:        name,
		Description: description,
		Steps:       steps,
		Status:      WorkflowStatusDraft,
	}
}
