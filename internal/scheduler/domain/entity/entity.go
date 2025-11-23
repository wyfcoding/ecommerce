package entity

import (
	"time"

	"gorm.io/gorm"
)

// JobStatus 任务状态
type JobStatus int8

const (
	JobStatusDisabled JobStatus = 0 // 禁用
	JobStatusEnabled  JobStatus = 1 // 启用
	JobStatusRunning  JobStatus = 2 // 运行中
)

// Job 定时任务实体
type Job struct {
	gorm.Model
	Name        string     `gorm:"type:varchar(128);uniqueIndex;not null;comment:任务名称" json:"name"`
	Description string     `gorm:"type:varchar(255);comment:任务描述" json:"description"`
	CronExpr    string     `gorm:"type:varchar(64);not null;comment:Cron表达式" json:"cron_expr"`
	Handler     string     `gorm:"type:varchar(128);not null;comment:处理器名称" json:"handler"`
	Params      string     `gorm:"type:text;comment:参数" json:"params"`
	Status      JobStatus  `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	LastRunTime *time.Time `gorm:"comment:上次运行时间" json:"last_run_time"`
	NextRunTime *time.Time `gorm:"comment:下次运行时间" json:"next_run_time"`
	RunCount    int64      `gorm:"not null;default:0;comment:运行次数" json:"run_count"`
	FailCount   int64      `gorm:"not null;default:0;comment:失败次数" json:"fail_count"`
}

// JobLog 任务日志实体
type JobLog struct {
	gorm.Model
	JobID     uint64     `gorm:"index;not null;comment:任务ID" json:"job_id"`
	JobName   string     `gorm:"type:varchar(128);not null;comment:任务名称" json:"job_name"`
	Handler   string     `gorm:"type:varchar(128);not null;comment:处理器名称" json:"handler"`
	Params    string     `gorm:"type:text;comment:参数" json:"params"`
	Status    string     `gorm:"type:varchar(32);not null;comment:状态(RUNNING,SUCCESS,FAILED)" json:"status"`
	Result    string     `gorm:"type:text;comment:执行结果" json:"result"`
	Error     string     `gorm:"type:text;comment:错误信息" json:"error"`
	Duration  int64      `gorm:"comment:耗时(ms)" json:"duration"`
	StartTime time.Time  `gorm:"not null;comment:开始时间" json:"start_time"`
	EndTime   *time.Time `gorm:"comment:结束时间" json:"end_time"`
}
