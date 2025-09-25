package data

import (
	"time"

	"gorm.io/gorm"
)

// ContentPage is the database model for a content page.
type ContentPage struct {
	gorm.Model
	Title       string    `gorm:"type:varchar(255);not null"`
	Slug        string    `gorm:"type:varchar(255);uniqueIndex;not null"` // URL-friendly identifier
	ContentHTML string    `gorm:"type:longtext"`
	Status      string    `gorm:"type:varchar(50);not null"` // e.g., DRAFT, PUBLISHED, ARCHIVED
}

// TableName specifies the table name for the ContentPage model.
func (ContentPage) TableName() string {
	return "cms_content_pages"
}

// ContentBlock is the database model for a reusable content block.
type ContentBlock struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(255);uniqueIndex;not null"` // Internal name for the block
	ContentHTML string    `gorm:"type:longtext"`
	Type        string    `gorm:"type:varchar(50);not null"` // e.g., HTML, MARKDOWN, JSON
}

// TableName specifies the table name for the ContentBlock model.
func (ContentBlock) TableName() string {
	return "cms_content_blocks"
}
