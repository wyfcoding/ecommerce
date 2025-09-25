package data

import (
	"time"

	"gorm.io/gorm"
)

// ChangeEvent represents a captured database change event.
type ChangeEvent struct {
	gorm.Model
	EventID         string            `gorm:"uniqueIndex;not null;comment:事件唯一ID" json:"eventId"`
	TableName       string            `gorm:"size:255;not null;comment:表名" json:"tableName"`
	OperationType   string            `gorm:"size:50;not null;comment:操作类型 (INSERT, UPDATE, DELETE)" json:"operationType"`
	PrimaryKeyValue string            `gorm:"size:255;not null;comment:主键值" json:"primaryKeyValue"`
	OldData         string            `gorm:"type:json;comment:旧数据 (JSON字符串)" json:"oldData"`
	NewData         string            `gorm:"type:json;comment:新数据 (JSON字符串)" json:"newData"`
	EventTimestamp  time.Time         `gorm:"not null;comment:事件发生时间" json:"eventTimestamp"`
}

// TableName specifies the table name for ChangeEvent.
func (ChangeEvent) TableName() string {
	return "cdc_change_events"
}
