// 变更说明：新增编排器（Orchestrator）功能，支持基于 Saga 模式的分布式事务协调、步骤管理及自动反向补偿。
// 假设：Saga 实例状态持久化于数据库，支持并发执行（分叉）及顺序执行（串行）。
package domain

import (
	"context"
	"time"
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
	ID            string      `json:"id"`        // 全局唯一事务ID
	SagaType      string      `json:"saga_type"` // Saga 类型（如：ORDER_PLACEMENT）
	Status        SagaStatus  `json:"status"`
	CurrentStep   int         `json:"current_step"` // 当前进度索引
	Steps         []*SagaStep `json:"steps"`        // 步骤明细
	ContextData   string      `json:"context_data"` // JSON 格式的共享上下文数据
	UserID        uint64      `json:"user_id"`
	OriginalRefID string      `json:"original_ref_id"` // 原始业务ID（如订单号）
	StartTime     time.Time   `json:"start_time"`
	EndTime       *time.Time  `json:"end_time"`
	UpdatedAt     time.Time   `json:"updated_at"`
	LastError     string      `json:"last_error"`
	MaxRetries    int         `json:"max_retries"` // 最大重试次数
}

// SagaStep Saga 步骤
type SagaStep struct {
	ID               uint64     `json:"id"`
	Name             string     `json:"name"`
	Service          string     `json:"service"`           // 负责的服务名
	Action           string     `json:"action"`            // 正向操作（如：DEDUCT_INVENTORY）
	CompensateAction string     `json:"compensate_action"` // 逆向补偿（如：RESTORE_INVENTORY）
	Status           StepStatus `json:"status"`
	Payload          string     `json:"payload"`   // 执行参数
	Response         string     `json:"response"`  // 执行结果
	TargetID         string     `json:"target_id"` // 步骤生成的ID（如：PaymentID）
	Error            string     `json:"error"`
	Retries          int        `json:"retries"`
	ExecutionTime    int64      `json:"execution_time_ms"` // 执行耗时
	ScheduledAt      *time.Time `json:"scheduled_at"`
	FinishedAt       *time.Time `json:"finished_at"`
}

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
	FindInstances(ctx context.Context, filter map[string]interface{}) ([]*SagaInstance, error)

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
