package domain

import (
	"time"

	"gorm.io/gorm"
)

// SourceType 定义了数据源的类型。
type SourceType string

const (
	SourceTypeMySQL      SourceType = "mysql"
	SourceTypePostgreSQL SourceType = "postgresql"
	SourceTypeCSV        SourceType = "csv"
	SourceTypeAPI        SourceType = "api"
)

// SourceStatus 定义了数据源的状态。
type SourceStatus int

const (
	SourceStatusActive   SourceStatus = 1 // 启用。
	SourceStatusInactive SourceStatus = 2 // 停用。
	SourceStatusError    SourceStatus = 3 // 错误。
)

// IngestionSource 实体代表一个数据源的配置信息。
type IngestionSource struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(255);uniqueIndex;not null;comment:数据源名称" json:"name"`
	Type        SourceType   `gorm:"type:varchar(32);not null;comment:类型" json:"type"`
	Config      string       `gorm:"type:text;comment:配置(JSON)" json:"config"`
	Description string       `gorm:"type:text;comment:描述" json:"description"`
	Status      SourceStatus `gorm:"default:1;comment:状态" json:"status"`
	LastRunAt   *time.Time   `gorm:"comment:上次运行时间" json:"last_run_at"`
}

// JobStatus 定义了数据摄取任务的生命周期状态。
type JobStatus int

const (
	JobStatusPending   JobStatus = 1 // 待处理。
	JobStatusRunning   JobStatus = 2 // 运行中。
	JobStatusCompleted JobStatus = 3 // 已完成。
	JobStatusFailed    JobStatus = 4 // 失败。
	JobStatusCancelled JobStatus = 5 // 已取消。
)

// IngestionJob 实体代表一个数据摄取任务的执行记录。
type IngestionJob struct {
	gorm.Model
	SourceID         uint64     `gorm:"not null;index;comment:数据源ID" json:"source_id"`
	Status           JobStatus  `gorm:"default:1;comment:状态" json:"status"`
	StartTime        *time.Time `gorm:"comment:开始时间" json:"start_time"`
	EndTime          *time.Time `gorm:"comment:结束时间" json:"end_time"`
	RecordsProcessed int64      `gorm:"default:0;comment:处理记录数" json:"records_processed"`
	ErrorMessage     string     `gorm:"type:text;comment:错误信息" json:"error_message"`
}

// IngestedEvent 实体代表一个被摄取的原始事件。
type IngestedEvent struct {
	gorm.Model
	EventType string    `gorm:"type:varchar(255);index;comment:事件类型" json:"event_type"`
	EventData string    `gorm:"type:text;comment:事件数据(JSON)" json:"event_data"`
	Source    string    `gorm:"type:varchar(255);comment:来源" json:"source"`
	Timestamp time.Time `gorm:"index;comment:事件时间戳" json:"timestamp"`
}

// NewIngestionSource 创建并返回一个新的 IngestionSource 实体实例。
func NewIngestionSource(name string, sourceType SourceType, config, description string) *IngestionSource {
	return &IngestionSource{
		Name:        name,
		Type:        sourceType,
		Config:      config,
		Description: description,
		Status:      SourceStatusActive,
	}
}

// NewIngestionJob 创建并返回一个新的 IngestionJob 实体实例。
func NewIngestionJob(sourceID uint64) *IngestionJob {
	return &IngestionJob{
		SourceID: sourceID,
		Status:   JobStatusPending,
	}
}

// Start 启动数据摄取任务。
func (j *IngestionJob) Start() {
	j.Status = JobStatusRunning
	now := time.Now()
	j.StartTime = &now
}

// Complete 完成数据摄取任务。
func (j *IngestionJob) Complete(recordsProcessed int64) {
	j.Status = JobStatusCompleted
	j.RecordsProcessed = recordsProcessed
	now := time.Now()
	j.EndTime = &now
}

// Fail 标记数据摄取任务失败。
func (j *IngestionJob) Fail(errorMessage string) {
	j.Status = JobStatusFailed
	j.ErrorMessage = errorMessage
	now := time.Now()
	j.EndTime = &now
}

// NewIngestedEvent 创建一个新的 IngestedEvent 实体。
func NewIngestedEvent(eventType, eventData, source string, timestamp time.Time) *IngestedEvent {
	return &IngestedEvent{
		EventType: eventType,
		EventData: eventData,
		Source:    source,
		Timestamp: timestamp,
	}
}
