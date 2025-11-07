package model

import "time"

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusActive   TaskStatus = "ACTIVE"   // 激活
	TaskStatusInactive TaskStatus = "INACTIVE" // 禁用
	TaskStatusPaused   TaskStatus = "PAUSED"   // 暂停
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCron   TaskType = "CRON"   // Cron表达式
	TaskTypeFixed  TaskType = "FIXED"  // 固定间隔
	TaskTypeOnce   TaskType = "ONCE"   // 一次性
)

// LogStatus 执行状态
type LogStatus string

const (
	LogStatusRunning LogStatus = "RUNNING" // 运行中
	LogStatusSuccess LogStatus = "SUCCESS" // 成功
	LogStatusFailed  LogStatus = "FAILED"  // 失败
)

// ScheduledTask 定时任务配置
type ScheduledTask struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	Name        string     `gorm:"type:varchar(100);uniqueIndex;not null;comment:任务名称" json:"name"`
	Description string     `gorm:"type:varchar(500);comment:任务描述" json:"description"`
	Type        TaskType   `gorm:"type:varchar(20);not null;comment:任务类型" json:"type"`
	CronExpr    string     `gorm:"type:varchar(100);comment:Cron表达式" json:"cronExpr"`
	Interval    int32      `gorm:"comment:执行间隔(秒)" json:"interval"`
	Handler     string     `gorm:"type:varchar(200);not null;comment:处理器名称" json:"handler"`
	Params      string     `gorm:"type:text;comment:任务参数JSON" json:"params"`
	Status      TaskStatus `gorm:"type:varchar(20);not null;comment:任务状态" json:"status"`
	Timeout     int32      `gorm:"not null;default:300;comment:超时时间(秒)" json:"timeout"`
	RetryCount  int32      `gorm:"not null;default:0;comment:重试次数" json:"retryCount"`
	LastRunAt   *time.Time `gorm:"comment:上次执行时间" json:"lastRunAt"`
	NextRunAt   *time.Time `gorm:"comment:下次执行时间" json:"nextRunAt"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

// TaskLog 任务执行日志
type TaskLog struct {
	ID         uint64     `gorm:"primarykey" json:"id"`
	TaskID     uint64     `gorm:"index;not null;comment:任务ID" json:"taskId"`
	TaskName   string     `gorm:"type:varchar(100);not null;comment:任务名称" json:"taskName"`
	Status     LogStatus  `gorm:"type:varchar(20);not null;comment:执行状态" json:"status"`
	StartTime  time.Time  `gorm:"not null;comment:开始时间" json:"startTime"`
	EndTime    *time.Time `gorm:"comment:结束时间" json:"endTime"`
	Duration   int64      `gorm:"comment:执行时长(毫秒)" json:"duration"`
	Result     string     `gorm:"type:text;comment:执行结果" json:"result"`
	ErrorMsg   string     `gorm:"type:text;comment:错误信息" json:"errorMsg"`
	RetryCount int32      `gorm:"not null;default:0;comment:重试次数" json:"retryCount"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}

// TaskLock 任务锁（防止重复执行）
type TaskLock struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	TaskName  string    `gorm:"type:varchar(100);uniqueIndex;not null;comment:任务名称" json:"taskName"`
	LockedAt  time.Time `gorm:"not null;comment:锁定时间" json:"lockedAt"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expiresAt"`
}

// TableName 指定表名
func (TaskLock) TableName() string {
	return "task_locks"
}

// IsExpired 判断锁是否过期
func (tl *TaskLock) IsExpired() bool {
	return time.Now().After(tl.ExpiresAt)
}
