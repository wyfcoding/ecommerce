package domain

import (
	"time"
)

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	Header EventHeader
	User   *User
}

// UserUpdatedEvent 用户更新事件
type UserUpdatedEvent struct {
	Header    EventHeader
	UserID    uint
	UpdatedAt time.Time
}

// UserPasswordChangedEvent 用户密码变更事件
type UserPasswordChangedEvent struct {
	Header EventHeader
	UserID uint
}

// EventHeader 事件头，包含元数据
type EventHeader struct {
	ID        string    `json:"id"`
	TraceID   string    `json:"trace_id"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}
