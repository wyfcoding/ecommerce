package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// SourceType 定义了数据源的类型。
type SourceType string

const (
	SourceTypeMySQL      SourceType = "mysql"      // MySQL数据库。
	SourceTypePostgreSQL SourceType = "postgresql" // PostgreSQL数据库。
	SourceTypeCSV        SourceType = "csv"        // CSV文件。
	SourceTypeAPI        SourceType = "api"        // 外部API接口。
)

// SourceStatus 定义了数据源的状态。
type SourceStatus int

const (
	SourceStatusActive   SourceStatus = 1 // 启用：数据源处于活跃状态，可以进行摄取。
	SourceStatusInactive SourceStatus = 2 // 停用：数据源已停用，不能进行摄取。
	SourceStatusError    SourceStatus = 3 // 错误：数据源连接或配置出现错误。
)

// IngestionSource 实体代表一个数据源的配置信息。
// 它包含了数据源的名称、类型、连接配置和当前状态。
type IngestionSource struct {
	gorm.Model               // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name        string       `gorm:"type:varchar(255);uniqueIndex;not null;comment:数据源名称" json:"name"` // 数据源名称，唯一索引，不允许为空。
	Type        SourceType   `gorm:"type:varchar(32);not null;comment:类型" json:"type"`                 // 数据源类型。
	Config      string       `gorm:"type:text;comment:配置(JSON)" json:"config"`                         // 数据源的连接配置，通常以JSON字符串形式存储。
	Description string       `gorm:"type:text;comment:描述" json:"description"`                          // 数据源描述。
	Status      SourceStatus `gorm:"default:1;comment:状态" json:"status"`                               // 数据源状态，默认为启用。
	LastRunAt   *time.Time   `gorm:"comment:上次运行时间" json:"last_run_at"`                                // 上次成功运行摄取任务的时间。
}

// JobStatus 定义了数据摄取任务的生命周期状态。
type JobStatus int

const (
	JobStatusPending   JobStatus = 1 // 待处理：任务已创建，等待调度。
	JobStatusRunning   JobStatus = 2 // 运行中：任务正在执行。
	JobStatusCompleted JobStatus = 3 // 已完成：任务成功完成。
	JobStatusFailed    JobStatus = 4 // 失败：任务执行失败。
	JobStatusCancelled JobStatus = 5 // 已取消：任务被手动取消。
)

// IngestionJob 实体代表一个数据摄取任务的执行记录。
// 它包含了任务的状态、执行时间、处理记录数和错误信息。
type IngestionJob struct {
	gorm.Model                  // 嵌入gorm.Model。
	SourceID         uint64     `gorm:"not null;index;comment:数据源ID" json:"source_id"`    // 关联的数据源ID，索引字段。
	Status           JobStatus  `gorm:"default:1;comment:状态" json:"status"`               // 任务状态，默认为待处理。
	StartTime        *time.Time `gorm:"comment:开始时间" json:"start_time"`                   // 任务开始执行的时间。
	EndTime          *time.Time `gorm:"comment:结束时间" json:"end_time"`                     // 任务结束执行的时间。
	RecordsProcessed int64      `gorm:"default:0;comment:处理记录数" json:"records_processed"` // 任务处理的记录数量。
	ErrorMessage     string     `gorm:"type:text;comment:错误信息" json:"error_message"`      // 任务失败时的错误信息。
}

// NewIngestionSource 创建并返回一个新的 IngestionSource 实体实例。
// name: 数据源名称。
// sourceType: 数据源类型。
// config: 配置信息。
// description: 描述。
func NewIngestionSource(name string, sourceType SourceType, config, description string) *IngestionSource {
	return &IngestionSource{
		Name:        name,
		Type:        sourceType,
		Config:      config,
		Description: description,
		Status:      SourceStatusActive, // 默认状态为启用。
	}
}

// NewIngestionJob 创建并返回一个新的 IngestionJob 实体实例。
// sourceID: 关联的数据源ID。
func NewIngestionJob(sourceID uint64) *IngestionJob {
	return &IngestionJob{
		SourceID: sourceID,
		Status:   JobStatusPending, // 初始状态为待处理。
	}
}

// Start 启动数据摄取任务，更新任务状态为“运行中”，并记录开始时间。
func (j *IngestionJob) Start() {
	j.Status = JobStatusRunning // 状态更新为“运行中”。
	now := time.Now()
	j.StartTime = &now // 记录开始时间。
}

// Complete 完成数据摄取任务，更新任务状态为“已完成”，记录处理记录数和结束时间。
// recordsProcessed: 任务处理的记录数量。
func (j *IngestionJob) Complete(recordsProcessed int64) {
	j.Status = JobStatusCompleted         // 状态更新为“已完成”。
	j.RecordsProcessed = recordsProcessed // 记录处理记录数。
	now := time.Now()
	j.EndTime = &now // 记录结束时间。
}

// Fail 标记数据摄取任务失败，更新任务状态为“失败”，记录错误信息和结束时间。
// errorMessage: 任务失败时的错误信息。
func (j *IngestionJob) Fail(errorMessage string) {
	j.Status = JobStatusFailed    // 状态更新为“失败”。
	j.ErrorMessage = errorMessage // 记录错误信息。
	now := time.Now()
	j.EndTime = &now // 记录结束时间。
}
