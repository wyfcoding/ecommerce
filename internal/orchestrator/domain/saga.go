// 变更说明：新增编排器（Orchestrator）功能，支持基于 Saga 模式的分布式事务协调、步骤管理及自动反向补偿。
// 假设：Saga 实例状态持久化于数据库，支持并发执行（分叉）及顺序执行（串行）。
package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// --- Saga 状态 ---

// SagaStatus Saga 实例状态
type SagaStatus string

const (
	SagaStarted      SagaStatus = "STARTED"      // 已启动
	SagaInProgress   SagaStatus = "IN_PROGRESS"  // 进行中
	SagaSucceeded    SagaStatus = "SUCCEEDED"    // 已成功完成
	SagaCompensating SagaStatus = "COMPENSATING" // 补偿中
	SagaFailed       SagaStatus = "FAILED"       // 业务失败且不可补偿
	SagaCompensated  SagaStatus = "COMPENSATED"  // 已成功补偿（回滚完成）
)

// --- Saga 步骤状态 ---

// StepStatus 步骤执行状态
type StepStatus string

const (
	StepPending      StepStatus = "PENDING"
	StepRunning      StepStatus = "RUNNING"
	StepCompleted    StepStatus = "COMPLETED"
	StepFailed       StepStatus = "FAILED"
	StepCompensating StepStatus = "COMPENSATING"
	StepCompensated  StepStatus = "COMPENSATED"
	StepSkipped      StepStatus = "SKIPPED"
)

// --- Saga 聚合根 ---

// SagaInstance Saga 实例
type SagaInstance struct {
	gorm.Model
	SagaID        string      `gorm:"column:saga_id;type:varchar(64);uniqueIndex;not null" json:"id"`    // 全局唯一事务ID
	SagaType      string      `gorm:"column:saga_type;type:varchar(64);index;not null" json:"saga_type"` // Saga 类型
	Status        SagaStatus  `gorm:"column:status;type:varchar(20);not null;default:'STARTED'" json:"status"`
	CurrentStep   int         `gorm:"column:current_step;not null;default:0" json:"current_step"`
	Steps         []*SagaStep `gorm:"foreignKey:SagaInstanceID" json:"steps"`
	ContextData   string      `gorm:"column:context_data;type:text" json:"context_data"`
	UserID        uint64      `gorm:"column:user_id;index;not null" json:"user_id"`
	OriginalRefID string      `gorm:"column:original_ref_id;type:varchar(64);index" json:"original_ref_id"`
	StartTime     time.Time   `gorm:"column:start_time;not null" json:"start_time"`
	EndTime       *time.Time  `gorm:"column:end_time" json:"end_time"`
	LastError     string      `gorm:"column:last_error;type:text" json:"last_error"`
	MaxRetries    int         `gorm:"column:max_retries;not null;default:3" json:"max_retries"`
}

func (SagaInstance) TableName() string { return "saga_instances" }

// SagaStep Saga 步骤
type SagaStep struct {
	gorm.Model
	SagaInstanceID   uint       `gorm:"column:saga_instance_id;index;not null"`
	Name             string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Service          string     `gorm:"column:service;type:varchar(64);not null" json:"service"`
	Action           string     `gorm:"column:action;type:varchar(64);not null" json:"action"`
	CompensateAction string     `gorm:"column:compensate_action;type:varchar(64);not null" json:"compensate_action"`
	Status           StepStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	Payload          string     `gorm:"column:payload;type:text" json:"payload"`
	Response         string     `gorm:"column:response;type:text" json:"response"`
	TargetID         string     `gorm:"column:target_id;type:varchar(64)" json:"target_id"`
	Error            string     `gorm:"column:error;type:text" json:"error"`
	Retries          int        `gorm:"column:retries;not null;default:0" json:"retries"`
	ExecutionTime    int64      `gorm:"column:execution_time_ms" json:"execution_time_ms"`
	ScheduledAt      *time.Time `gorm:"column:scheduled_at" json:"scheduled_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at" json:"finished_at"`
}

func (SagaStep) TableName() string { return "saga_steps" }

// --- Saga 定义 (Blueprint) ---

// SagaDefinition Saga 蓝图定义
type SagaDefinition struct {
	SagaType    string        `json:"saga_type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	StepConfigs []*StepConfig `json:"step_configs"`
}

// StepConfig 步骤配置
type StepConfig struct {
	Name             string `json:"name"`
	Service          string `json:"service"`
	Action           string `json:"action"`
	CompensateAction string `json:"compensate_action"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
}

// --- 编排器仓储接口 ---

// SagaRepository Saga 仓储接口
type SagaRepository interface {
	SaveInstance(ctx context.Context, instance *SagaInstance) error
	FindInstance(ctx context.Context, id string) (*SagaInstance, error)
	FindInstances(ctx context.Context, filter map[string]any) ([]*SagaInstance, error)

	// Saga 定义管理
	RegisterDefinition(ctx context.Context, def *SagaDefinition) error
	GetDefinition(ctx context.Context, sagaType string) (*SagaDefinition, error)
}

// --- 状态驱动方法 ---

// SetStatus 更新 Saga 状态并记录时间
func (s *SagaInstance) SetStatus(status SagaStatus) {
	s.Status = status
	if status == SagaSucceeded || status == SagaCompensated || status == SagaFailed {
		now := time.Now()
		s.EndTime = &now
	}
}

// CanCompensate 检查是否可以进行补偿
func (s *SagaInstance) CanCompensate() bool {
	return s.Status == SagaInProgress || s.Status == SagaFailed || s.Status == SagaCompensating
}

// NextStep 获取下一个待执行的步骤
func (s *SagaInstance) NextStep() *SagaStep {
	if s.Status == SagaSucceeded || s.Status == SagaCompensated {
		return nil
	}
	for i, step := range s.Steps {
		if step.Status == StepPending {
			s.CurrentStep = i
			return step
		}
	}
	return nil
}

// LastSuccessStep 获取最后一个执行成功的步骤（用于补偿）
func (s *SagaInstance) LastSuccessStep() *SagaStep {
	for i := len(s.Steps) - 1; i >= 0; i-- {
		if s.Steps[i].Status == StepCompleted {
			return s.Steps[i]
		}
	}
	return nil
}
