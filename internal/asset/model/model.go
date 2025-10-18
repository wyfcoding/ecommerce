package model

import (
	"time"

	"gorm.io/gorm"
)

// AssetType 定义了资产的类型
type AssetType string

const (
	TypeProductImage AssetType = "PRODUCT_IMAGE"
	TypeCategoryImage AssetType = "CATEGORY_IMAGE"
	TypeBrandLogo    AssetType = "BRAND_LOGO"
	TypeUserAvatar   AssetType = "USER_AVATAR"
)

// Asset 资产元数据模型
type Asset struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	FileName   string    `gorm:"type:varchar(255);not null" json:"file_name"` // 原始文件名
	ObjectKey  string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"object_key"` // 在对象存储中的唯一键 (e.g., "products/2023/10/some-uuid.jpg")
	Size       int64     `gorm:"not null" json:"size"` // 文件大小 (bytes)
	MimeType   string    `gorm:"type:varchar(100);not null" json:"mime_type"` // MIME 类型 (e.g., "image/jpeg")
	AssetType  AssetType `gorm:"type:varchar(50);index" json:"asset_type"` // 资产类型
	OwnerID    uint      `gorm:"index" json:"owner_id"` // 所属者ID (例如，如果类型是 USER_AVATAR，这里就是 UserID)
	RelatedID  uint      `gorm:"index" json:"related_id"` // 关联实体ID (例如，如果类型是 PRODUCT_IMAGE，这里就是 ProductID)
	CreatedAt  time.Time `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 自定义表名
func (Asset) TableName() string {
	return "assets"
}