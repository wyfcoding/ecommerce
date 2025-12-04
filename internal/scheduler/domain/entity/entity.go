package entity

import (
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// JobStatus 定义了定时任务的运行状态。
type JobStatus int8

const (
	JobStatusDisabled JobStatus = 0 // 禁用：任务不会被调度执行。
	JobStatusEnabled  JobStatus = 1 // 启用：任务会按照Cron表达式调度执行。
	JobStatusRunning  JobStatus = 2 // 运行中：任务当前正在执行。
)

// Job 实体是定时任务模块的聚合根。
// 它包含了定时任务的配置信息，如名称、Cron表达式、处理器和运行统计。
type Job struct {
	gorm.Model             // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name        string     `gorm:"type:varchar(128);uniqueIndex;not null;comment:任务名称" json:"name"` // 任务名称，唯一索引，不允许为空。
	Description string     `gorm:"type:varchar(255);comment:任务描述" json:"description"`               // 任务的简要描述。
	CronExpr    string     `gorm:"type:varchar(64);not null;comment:Cron表达式" json:"cron_expr"`      // 定时任务的Cron表达式。
	Handler     string     `gorm:"type:varchar(128);not null;comment:处理器名称" json:"handler"`         // 任务执行的处理器名称。
	Params      string     `gorm:"type:text;comment:参数" json:"params"`                              // 任务执行所需的参数，通常为JSON字符串。
	Status      JobStatus  `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`        // 任务状态，默认为启用。
	LastRunTime *time.Time `gorm:"comment:上次运行时间" json:"last_run_time"`                             // 任务最后一次运行的时间。
	NextRunTime *time.Time `gorm:"comment:下次运行时间" json:"next_run_time"`                             // 任务下次预定运行的时间。
	RunCount    int64      `gorm:"not null;default:0;comment:运行次数" json:"run_count"`                // 任务总运行次数。
	FailCount   int64      `gorm:"not null;default:0;comment:失败次数" json:"fail_count"`               // 任务失败次数。
}

// JobLog 实体代表一次定时任务的执行日志。
// 它记录了任务执行的详细信息，包括开始时间、结束时间、状态和结果。
type JobLog struct {
	gorm.Model            // 嵌入gorm.Model。
	JobID      uint64     `gorm:"index;not null;comment:任务ID" json:"job_id"`                                  // 关联的定时任务ID，索引字段。
	JobName    string     `gorm:"type:varchar(128);not null;comment:任务名称" json:"job_name"`                    // 任务名称。
	Handler    string     `gorm:"type:varchar(128);not null;comment:处理器名称" json:"handler"`                    // 任务处理器名称。
	Params     string     `gorm:"type:text;comment:参数" json:"params"`                                         // 任务执行时的参数。
	Status     string     `gorm:"type:varchar(32);not null;comment:状态(RUNNING,SUCCESS,FAILED)" json:"status"` // 任务执行状态，例如“RUNNING”、“SUCCESS”、“FAILED”。
	Result     string     `gorm:"type:text;comment:执行结果" json:"result"`                                       // 任务执行结果的详细信息。
	Error      string     `gorm:"type:text;comment:错误信息" json:"error"`                                        // 如果任务失败，记录错误信息。
	Duration   int64      `gorm:"comment:耗时(ms)" json:"duration"`                                             // 任务执行耗时（毫秒）。
	StartTime  time.Time  `gorm:"not null;comment:开始时间" json:"start_time"`                                    // 任务开始执行的时间。
	EndTime    *time.Time `gorm:"comment:结束时间" json:"end_time"`                                               // 任务结束执行的时间。
}
