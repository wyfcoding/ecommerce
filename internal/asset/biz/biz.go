package biz

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

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

// AssetRepo defines the interface for asset data access.
type AssetRepo interface {
	UploadFile(ctx context.Context, file *File, fileContent io.Reader) (*File, error)
}

// AssetUsecase is the business logic for asset management.
type AssetUsecase struct {
	repo AssetRepo
}

// NewAssetUsecase creates a new AssetUsecase.
func NewAssetUsecase(repo AssetRepo) *AssetUsecase {
	return &AssetUsecase{repo: repo}
}

// UploadFile handles the business logic for uploading a file.
func (uc *AssetUsecase) UploadFile(ctx context.Context, fileName, contentType, bucketName string, fileSize int64, uploadedBy uint64, fileContent io.Reader) (*File, error) {
	// Generate a unique file ID
	fileID := uuid.New().String()

	file := &File{
		FileID:      fileID,
		FileName:    fileName,
		ContentType: contentType,
		BucketName:  bucketName,
		FileSize:    fileSize,
		UploadedBy:  uploadedBy,
	}

	// Call the repository to upload and save metadata
	uploadedFile, err := uc.repo.UploadFile(ctx, file, fileContent)
	if err != nil {
		return nil, err
	}

	return uploadedFile, nil
}
