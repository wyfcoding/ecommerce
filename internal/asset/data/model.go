package data

import (
	"time"

	"gorm.io/gorm"
)

// File represents metadata for an uploaded file.
type File struct {
	gorm.Model
	FileID      string `gorm:"uniqueIndex;not null;comment:文件唯一ID" json:"fileId"`
	FileName    string `gorm:"size:255;not null;comment:文件名" json:"fileName"`
	ContentType string `gorm:"size:100;comment:文件类型" json:"contentType"`
	BucketName  string `gorm:"size:100;not null;comment:存储桶名称" json:"bucketName"`
	FilePath    string `gorm:"size:512;not null;comment:文件路径/URL" json:"filePath"`
	FileSize    int64  `gorm:"comment:文件大小" json:"fileSize"`
	UploadedBy  uint64 `gorm:"comment:上传用户ID" json:"uploadedBy"` // Optional: if associated with a user
}

// TableName specifies the table name for File.
func (File) TableName() string {
	return "files"
}
