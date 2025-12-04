package entity

import (
	"gorm.io/gorm" // 导入GORM库。
)

// FileType 定义了文件的类型。
type FileType string

const (
	FileTypeImage    FileType = "image"    // 图片文件。
	FileTypeVideo    FileType = "video"    // 视频文件。
	FileTypeDocument FileType = "document" // 文档文件。
	FileTypeOther    FileType = "other"    // 其他类型文件。
)

// FileMetadata 实体代表一个文件的元数据信息。
// 它记录了文件的名称、大小、类型、存储路径和访问URL等。
type FileMetadata struct {
	gorm.Model          // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name       string   `gorm:"type:varchar(255);not null;comment:文件名" json:"name"`         // 文件名，不允许为空。
	Size       int64    `gorm:"not null;comment:文件大小(字节)" json:"size"`                      // 文件大小，单位为字节。
	Type       FileType `gorm:"type:varchar(32);not null;comment:文件类型" json:"type"`         // 文件类型。
	Path       string   `gorm:"type:varchar(512);not null;comment:存储路径" json:"path"`        // 文件在存储系统中的路径。
	URL        string   `gorm:"type:varchar(512);not null;comment:访问URL" json:"url"`        // 文件的可访问URL。
	Checksum   string   `gorm:"type:varchar(64);comment:文件校验和(MD5/SHA256)" json:"checksum"` // 文件内容的校验和，用于验证文件完整性。
	Bucket     string   `gorm:"type:varchar(64);comment:存储桶" json:"bucket"`                 // 文件所在的存储桶名称。
}

// NewFileMetadata 创建并返回一个新的 FileMetadata 实体实例。
// name: 文件名。
// size: 文件大小。
// fileType: 文件类型。
// path: 存储路径。
// url: 访问URL。
// checksum: 文件校验和。
// bucket: 存储桶名称。
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
