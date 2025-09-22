package data

import (
	"time"
)

// DataLakeRecord represents a record to be ingested into the data lake (Hudi).
type DataLakeRecord struct {
	TableName string            `json:"table_name"`
	RecordID  string            `json:"record_id"`
	Data      map[string]string `json:"data"`
	EventTime time.Time         `json:"event_time"`
	IngestedAt time.Time        `json:"ingested_at"`
}
