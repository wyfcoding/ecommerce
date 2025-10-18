package model

import "time"

// File represents metadata for an uploaded file in the business logic layer.
type File struct {
	ID          uint   // Database ID
	FileID      string // Unique ID for the file (e.g., UUID)
	FileName    string
	ContentType string
	BucketName  string
	FilePath    string // URL or path to the file
	FileSize    int64
	UploadedBy  uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
