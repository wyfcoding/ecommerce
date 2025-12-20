package domain

import (
	"gorm.io/gorm"
)

// FileType 定义了文件的类型。
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeDocument FileType = "document"
	FileTypeOther    FileType = "other"
)

// FileMetadata 实体代表一个文件的元数据信息。
type FileMetadata struct {
	gorm.Model
	Name     string   `gorm:"type:varchar(255);not null;comment:文件名" json:"name"`
	Size     int64    `gorm:"not null;comment:文件大小(字节)" json:"size"`
	Type     FileType `gorm:"type:varchar(32);not null;comment:文件类型" json:"type"`
	Path     string   `gorm:"type:varchar(512);not null;comment:存储路径" json:"path"`
	URL      string   `gorm:"type:varchar(512);not null;comment:访问URL" json:"url"`
	Checksum string   `gorm:"type:varchar(64);comment:文件校验和(MD5/SHA256)" json:"checksum"`
	Bucket   string   `gorm:"type:varchar(64);comment:存储桶" json:"bucket"`
}

// NewFileMetadata 创建并返回一个新的 FileMetadata 实体实例。
func NewFileMetadata(name string, size int64, fileType FileType, path, url, checksum, bucket string) *FileMetadata {
	return &FileMetadata{
		Name:     name,
		Size:     size,
		Type:     fileType,
		Path:     path,
		URL:      url,
		Checksum: checksum,
		Bucket:   bucket,
	}
}
