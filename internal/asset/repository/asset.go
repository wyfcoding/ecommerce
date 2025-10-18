package repository

import (
	"context"
	"io"

	"ecommerce/internal/asset/model"
)

// AssetRepo defines the interface for asset data access.
type AssetRepo interface {
	UploadFile(ctx context.Context, file *model.File, fileContent io.Reader) (*model.File, error)
}