package domain

import "time"

const (
	UserCreatedEventType         = "user.created"
	UserUpdatedEventType         = "user.updated"
	UserPasswordChangedEventType = "user.password.changed"
	UserAddressAddedEventType    = "user.address.added"
	UserAddressUpdatedEventType  = "user.address.updated"
	UserAddressDeletedEventType  = "user.address.deleted"
	UserAddressDefaultEventType  = "user.address.default_set"
)

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    int8      `json:"status"`
	CreatedAt int64     `json:"created_at"`
	OccurredOn time.Time `json:"occurred_on"`
}

// UserUpdatedEvent 用户更新事件
type UserUpdatedEvent struct {
	UserID    uint      `json:"user_id"`
	UpdatedAt int64     `json:"updated_at"`
	OccurredOn time.Time `json:"occurred_on"`
}

// UserPasswordChangedEvent 用户密码变更事件
type UserPasswordChangedEvent struct {
	UserID     uint      `json:"user_id"`
	ChangedAt  int64     `json:"changed_at"`
	OccurredOn time.Time `json:"occurred_on"`
}

// UserAddressAddedEvent 地址添加事件
type UserAddressAddedEvent struct {
	UserID      uint      `json:"user_id"`
	AddressID   uint      `json:"address_id"`
	IsDefault   bool      `json:"is_default"`
	OccurredOn  time.Time `json:"occurred_on"`
}

// UserAddressUpdatedEvent 地址更新事件
type UserAddressUpdatedEvent struct {
	UserID     uint      `json:"user_id"`
	AddressID  uint      `json:"address_id"`
	IsDefault  bool      `json:"is_default"`
	OccurredOn time.Time `json:"occurred_on"`
}

// UserAddressDeletedEvent 地址删除事件
type UserAddressDeletedEvent struct {
	UserID     uint      `json:"user_id"`
	AddressID  uint      `json:"address_id"`
	OccurredOn time.Time `json:"occurred_on"`
}

// UserAddressDefaultEvent 地址默认变更事件
type UserAddressDefaultEvent struct {
	UserID     uint      `json:"user_id"`
	AddressID  uint      `json:"address_id"`
	OccurredOn time.Time `json:"occurred_on"`
}

