package model

import "time"

// QueryResult represents a result from a data warehouse query in the business logic layer.
type QueryResult struct {
	ColumnNames []string
	Rows        []map[string]string
	Message     string
	QueryTime   time.Time
}
