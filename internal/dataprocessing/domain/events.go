package domain

import "time"

// JobStartedEvent 数据处理任务开始事件。
type JobStartedEvent struct {
	JobID     uint64    `json:"job_id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// JobCompletedEvent 数据处理任务完成事件。
type JobCompletedEvent struct {
	JobID     uint64    `json:"job_id"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// JobFailedEvent 数据处理任务失败事件。
type JobFailedEvent struct {
	JobID     uint64    `json:"job_id"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// DataCleanedEvent 数据清洗完成事件。
type DataCleanedEvent struct {
	BatchID   string    `json:"batch_id"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}
