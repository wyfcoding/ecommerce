package model

import "time"

// ChangeEvent represents a captured database change event in the business logic layer.
type ChangeEvent struct {
	ID              uint
	EventID         string
	TableName       string
	OperationType   string
	PrimaryKeyValue string
	OldData         string // JSON string
	NewData         string // JSON string
	EventTimestamp  time.Time
}
