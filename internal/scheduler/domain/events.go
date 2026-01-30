package domain

import "time"

// JobScheduledEvent 任务已调度事件。
type JobScheduledEvent struct {
	JobID     uint64    `json:"job_id"`
	Name      string    `json:"name"`
	GroupName string    `json:"group_name"`
	Timestamp time.Time `json:"timestamp"`
}

// JobExecutedEvent 任务已执行事件。
type JobExecutedEvent struct {
	JobID     uint64    `json:"job_id"`
	Success   bool      `json:"success"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// JobCancelledEvent 任务已取消事件。
type JobCancelledEvent struct {
	JobID     uint64    `json:"job_id"`
	Timestamp time.Time `json:"timestamp"`
}
