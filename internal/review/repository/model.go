package repository

import (
	"gorm.io/gorm"
)

// Review is the database model for a product review.
type Review struct {
	gorm.Model
	ProductID string `gorm:"type:varchar(100);not null;index"`
	UserID    string `gorm:"type:varchar(100);not null;index"`
	Rating    int32  `gorm:"type:int;not null"` // 1-5 stars
	Title     string `gorm:"type:varchar(255)"`
	Content   string `gorm:"type:text"`
}

// TableName specifies the table name for the Review model.
func (Review) TableName() string {
	return "reviews"
}
