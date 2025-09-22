package data

import (
	"time"
)

// QueryResult represents a result from a data warehouse query.
type QueryResult struct {
	ColumnNames []string            `json:"column_names"`
	Rows        []map[string]string `json:"rows"` // Each row is a map of column name to value
	Message     string              `json:"message"`
	QueryTime   time.Time           `json:"query_time"`
}
