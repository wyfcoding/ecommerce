package domain

import "time"

// DataIngestedEvent 数据摄入完成事件。
type DataIngestedEvent struct {
	IngestionID uint64    `json:"ingestion_id"`
	Source      string    `json:"source"`
	DataType    string    `json:"data_type"`
	Timestamp   time.Time `json:"timestamp"`
}

// IngestionFailedEvent 数据摄入失败事件。
type IngestionFailedEvent struct {
	IngestionID uint64    `json:"ingestion_id"`
	Source      string    `json:"source"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}
