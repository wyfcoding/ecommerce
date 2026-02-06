package domain

import "time"

const (
	AIModelCreatedEventType       = "aimodel.model.created"
	AIModelStatusUpdatedEventType = "aimodel.model.status.updated"
)

// AIModelCreatedEvent 模型创建事件。
type AIModelCreatedEvent struct {
	ModelID   uint64      `json:"model_id"`
	ModelNo   string      `json:"model_no"`
	Status    ModelStatus `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
}

// AIModelStatusUpdatedEvent 模型状态变更事件。
type AIModelStatusUpdatedEvent struct {
	ModelID   uint64      `json:"model_id"`
	OldStatus ModelStatus `json:"old_status"`
	NewStatus ModelStatus `json:"new_status"`
	Timestamp time.Time   `json:"timestamp"`
}
