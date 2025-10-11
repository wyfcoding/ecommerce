package data

import (
	"context"
	"ecommerce/internal/data_lake_ingestion/biz"
	"ecommerce/internal/data_lake_ingestion/data/model"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	hudiDataLakePath = "/tmp/hudi_data_lake.log" // Placeholder for Hudi data path
)

type dataLakeIngestionRepo struct {
	data *Data // Placeholder for common data dependencies if any
	// TODO: Add Hudi client or API client for ingestion
}

// NewDataLakeIngestionRepo creates a new DataLakeIngestionRepo.
func NewDataLakeIngestionRepo(data *Data) biz.DataLakeIngestionRepo {
	return &dataLakeIngestionRepo{data: data}
}

// IngestData simulates ingesting data into a Hudi data lake.
func (r *dataLakeIngestionRepo) IngestData(ctx context.Context, tableName, recordID string, data map[string]string, eventTime time.Time) error {
	record := model.DataLakeRecord{
		TableName:  tableName,
		RecordID:   recordID,
		Data:       data,
		EventTime:  eventTime,
		IngestedAt: time.Now(),
	}

	// Simulate writing to a Hudi data lake (e.g., appending to a file)
	file, err := os.OpenFile(hudiDataLakePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open Hudi data lake file: %w", err)
	}
	defer file.Close()

	recordBytes, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal data lake record: %w", err)
	}

	_, err = file.WriteString(string(recordBytes) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write record to Hudi data lake file: %w", err)
	}

	fmt.Printf("Data ingested to Hudi: Table=%s, RecordID=%s\n", record.TableName, record.RecordID)
	return nil
}
