package model

import "time"

// DataLakeRecord represents a record to be ingested into the data lake (Hudi) in the business logic layer.
type DataLakeRecord struct {
	TableName string
	RecordID  string
	Data      map[string]string
	EventTime time.Time
}
