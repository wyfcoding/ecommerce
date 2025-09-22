package data

import (
	"context"
	"ecommerce/internal/asset/biz"
	"ecommerce/internal/asset/data/model"
	"ecommerce/pkg/minio" // Assuming this package exists
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type assetRepo struct {
	data *Data
	minioClient *minio.Client
}

// NewAssetRepo creates a new AssetRepo.
func NewAssetRepo(data *Data, minioClient *minio.Client) biz.AssetRepo {
	return &assetRepo{data: data, minioClient: minioClient}
}

// UploadFile uploads a file to MinIO and saves its metadata to the database.
func (r *assetRepo) UploadFile(ctx context.Context, file *biz.File, fileContent io.Reader) (*biz.File, error) {
	// 1. Upload to MinIO
	info, err := r.minioClient.PutObject(ctx, file.BucketName, file.FileID, fileContent, file.FileSize, minio.PutObjectOptions{ContentType: file.ContentType})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}
	file.FilePath = fmt.Sprintf("%s/%s/%s", r.minioClient.Endpoint, file.BucketName, file.FileID) // Construct URL
	file.FileSize = info.Size

	// 2. Save metadata to database
	po := &model.File{
		FileID:      file.FileID,
		FileName:    file.FileName,
		ContentType: file.ContentType,
		BucketName:  file.BucketName,
		FilePath:    file.FilePath,
		FileSize:    file.FileSize,
		UploadedBy:  file.UploadedBy,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, fmt.Errorf("failed to save file metadata to database: %w", err)
	}
	file.ID = po.ID // Assign generated ID

	return file, nil
}
