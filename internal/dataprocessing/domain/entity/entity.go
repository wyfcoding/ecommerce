package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// TaskStatus 定义了数据处理任务的生命周期状态。
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = 1 // 待处理：任务已提交，等待调度执行。
	TaskStatusRunning   TaskStatus = 2 // 运行中：任务正在执行。
	TaskStatusCompleted TaskStatus = 3 // 已完成：任务成功执行完毕。
	TaskStatusFailed    TaskStatus = 4 // 失败：任务执行过程中出现错误。
	TaskStatusCancelled TaskStatus = 5 // 已取消：任务被手动或系统取消。
)

// ProcessingTask 实体代表一个数据处理任务。
// 它包含了任务的名称、类型、状态、配置、执行结果和所属工作流等信息。
type ProcessingTask struct {
	gorm.Model              // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name         string     `gorm:"type:varchar(255);not null;comment:任务名称" json:"name"` // 任务名称。
	Type         string     `gorm:"type:varchar(64);not null;comment:任务类型" json:"type"`  // 任务类型，例如“数据清洗”、“数据转换”。
	Status       TaskStatus `gorm:"default:1;comment:状态" json:"status"`                  // 任务状态，默认为待处理。
	Config       string     `gorm:"type:text;comment:配置(JSON)" json:"config"`            // 任务的配置信息，通常以JSON字符串形式存储。
	Result       string     `gorm:"type:text;comment:结果(JSON)" json:"result"`            // 任务执行结果，通常以JSON字符串形式存储。
	ErrorMessage string     `gorm:"type:text;comment:错误信息" json:"error_message"`         // 任务失败时的错误信息。
	StartTime    *time.Time `gorm:"comment:开始时间" json:"start_time"`                      // 任务开始执行的时间。
	EndTime      *time.Time `gorm:"comment:结束时间" json:"end_time"`                        // 任务结束执行的时间。
	WorkflowID   uint64     `gorm:"index;comment:所属工作流ID" json:"workflow_id"`            // 关联的工作流ID，索引字段。
}

// WorkflowStatus 定义了数据处理工作流的状态。
type WorkflowStatus int

const (
	WorkflowStatusActive   WorkflowStatus = 1 // 激活：工作流正在运行或可被触发。
	WorkflowStatusInactive WorkflowStatus = 2 // 停用：工作流已停用。
	WorkflowStatusDraft    WorkflowStatus = 3 // 草稿：工作流已创建但尚未启用。
)

// ProcessingWorkflow 实体代表一个数据处理工作流。
// 它定义了一系列数据处理任务的编排和执行顺序。
type ProcessingWorkflow struct {
	gorm.Model                 // 嵌入gorm.Model。
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null;comment:工作流名称" json:"name"` // 工作流名称，唯一索引，不允许为空。
	Description string         `gorm:"type:text;comment:描述" json:"description"`                          // 工作流描述。
	Steps       string         `gorm:"type:text;comment:步骤定义(JSON)" json:"steps"`                        // 工作流的步骤定义，通常以JSON或YAML字符串形式存储。
	Status      WorkflowStatus `gorm:"default:3;comment:状态" json:"status"`                               // 工作流状态，默认为草稿。
}

// NewProcessingTask 创建并返回一个新的 ProcessingTask 实体实例。
// name: 任务名称。
// taskType: 任务类型。
// config: 任务配置。
// workflowID: 所属工作流ID。
func NewProcessingTask(name, taskType, config string, workflowID uint64) *ProcessingTask {
	return &ProcessingTask{
		Name:       name,
		Type:       taskType,
		Config:     config,
		WorkflowID: workflowID,
		Status:     TaskStatusPending, // 初始状态为待处理。
	}
}

// Start 启动数据处理任务，更新任务状态为“运行中”，并记录开始时间。
func (t *ProcessingTask) Start() {
	t.Status = TaskStatusRunning // 状态更新为“运行中”。
	now := time.Now()
	t.StartTime = &now // 记录开始时间。
}

// Complete 完成数据处理任务，更新任务状态为“已完成”，记录处理结果和结束时间。
// result: 任务执行结果。
func (t *ProcessingTask) Complete(result string) {
	t.Status = TaskStatusCompleted // 状态更新为“已完成”。
	t.Result = result              // 记录处理结果。
	now := time.Now()
	t.EndTime = &now // 记录结束时间。
}

// Fail 标记数据处理任务失败，更新任务状态为“失败”，记录错误信息和结束时间。
// errorMessage: 任务失败时的错误信息。
func (t *ProcessingTask) Fail(errorMessage string) {
	t.Status = TaskStatusFailed   // 状态更新为“失败”。
	t.ErrorMessage = errorMessage // 记录错误信息。
	now := time.Now()
	t.EndTime = &now // 记录结束时间。
}

// NewProcessingWorkflow 创建并返回一个新的 ProcessingWorkflow 实体实例。
// name: 工作流名称。
// description: 描述。
// steps: 步骤定义。
func NewProcessingWorkflow(name, description, steps string) *ProcessingWorkflow {
	return &ProcessingWorkflow{
		Name:        name,
		Description: description,
		Steps:       steps,
		Status:      WorkflowStatusDraft, // 初始状态为草稿。
	}
}
